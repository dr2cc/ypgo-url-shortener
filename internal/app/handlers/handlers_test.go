package handlers

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	t.Helper()

	var err error
	var req *http.Request
	var resp *http.Response
	var respBody []byte

	req, err = http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Do(req)
	require.NoError(t, err)

	respBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	require.NoError(t, err)

	return resp, string(bytes.TrimSpace(respBody))
}

func testGzippedRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	t.Helper()

	var err error
	var req *http.Request
	var resp *http.Response
	var respBody []byte

	var b bytes.Buffer

	w, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)

	_, err = w.Write([]byte(body))
	if err != nil {
		return nil, err.Error()
	}

	err = w.Close()
	if err != nil {
		return nil, err.Error()
	}

	req, err = http.NewRequest(method, ts.URL+path, bytes.NewReader(b.Bytes()))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Do(req)
	require.NoError(t, err)

	respBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	require.NoError(t, err)

	return resp, string(bytes.TrimSpace(respBody))
}

func TestHandler_Expand(t *testing.T) {
	type want struct {
		statusCode int
		location   string
		body       string
	}
	tests := []struct {
		name    string
		request string
		want    want
		method  string
	}{
		{
			name: "get with existing id",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "/url",
				body:       "",
			},
			request: "/id",
			method:  http.MethodGet,
		},
		{
			name: "get with missing id",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
				body:       "cant find full url",
			},
			request: "/missing",
			method:  http.MethodGet,
		},
		{
			name: "get with null id",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				location:   "",
				body:       "",
			},
			request: "/",
			method:  http.MethodGet,
		},
		{
			name: "not supported method",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				location:   "",
				body:       "",
			},
			request: "/asd",
			method:  http.MethodDelete,
		},
		{
			name: "it returns 500 when service fails on expanding",
			want: want{
				statusCode: http.StatusInternalServerError,
				location:   "",
				body:       "error text",
			},
			request: "/error",
			method:  http.MethodGet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockRepo)
			mockRepo.On("GetByID", "id").Return("url", nil)
			mockRepo.On("GetByID", "missing").Return("", nil)
			mockRepo.On("GetByID", "error").Return("", errors.New("error text"))

			mockGen := new(mocks.MockGen)
			cfg := config.New()
			service := services.New(mockRepo, mockGen, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, tt.request, "")
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}

func TestHandler_Shorten(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name   string
		want   want
		body   string
		method string
	}{
		{
			name: "post with url",
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://localhost:8080/id",
			},
			method: http.MethodPost,
			body:   "url",
		},
		{
			name: "post without url",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "url required",
			},
			method: http.MethodPost,
			body:   "",
		},
		{
			name: "not supported method",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				body:       "",
			},
			method: http.MethodGet,
			body:   "",
		},
		{
			name: "it returns 500 when service fails on shortening",
			want: want{
				statusCode: http.StatusInternalServerError,
				body:       "err",
			},
			method: http.MethodPost,
			body:   "error_on_shortening",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockRepo)
			mockRepo.On("Save", "url", "id").Return(nil)

			mockGen := new(mocks.MockGen)
			mockGen.On("GenerateIDFromString", "url").Return("id", nil)
			mockGen.On("GenerateIDFromString", "").Return("", errors.New("err"))
			mockGen.On("GenerateIDFromString", "error_on_shortening").Return("", errors.New("err"))

			cfg := config.New()
			service := services.New(mockRepo, mockGen, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, "/", tt.body)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)

			result, body = testGzippedRequest(t, ts, tt.method, "/", tt.body)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)
			fmt.Println(result.Cookies())
		})
	}
}

func TestHandler_ShortenAPI(t *testing.T) {
	type want struct {
		statusCode  int
		body        string
		contentType string
	}
	tests := []struct {
		name   string
		want   want
		body   string
		method string
	}{
		{
			name: "post with url",
			want: want{
				statusCode:  http.StatusCreated,
				body:        "{\"result\":\"http://localhost:8080/id\"}",
				contentType: "application/json",
			},
			method: http.MethodPost,
			body:   "{\"url\":\"url\"}",
		},
		{
			name: "post without url",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "url required",
			},
			method: http.MethodPost,
			body:   "{\"url\":\"\"}",
		},
		{
			name: "post without url param",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "url required",
			},
			method: http.MethodPost,
			body:   "{\"some param\":\"some value\"}",
		},
		{
			name: "post without json",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "cannot decode json",
			},
			method: http.MethodPost,
			body:   "url",
		},
		{
			name: "not supported method",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				body:       "",
			},
			method: http.MethodGet,
			body:   "",
		},
		{
			name: "it returns 500 when service fails on shortening",
			want: want{
				statusCode: http.StatusInternalServerError,
				body:       "err",
			},
			method: http.MethodPost,
			body:   "{\"url\":\"error_on_shortening\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockRepo)
			mockRepo.On("Save", "url", "id").Return(nil)

			mockGen := new(mocks.MockGen)
			mockGen.On("GenerateIDFromString", "url").Return("id", nil)
			mockGen.On("GenerateIDFromString", "").Return("", errors.New("err"))
			mockGen.On("GenerateIDFromString", "error_on_shortening").Return("", errors.New("err"))

			cfg := config.New()
			service := services.New(mockRepo, mockGen, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, "/api/shorten", tt.body)
			defer result.Body.Close()

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)

			result, body = testGzippedRequest(t, ts, tt.method, "/api/shorten", tt.body)
			defer result.Body.Close()

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)
		})
	}
}
