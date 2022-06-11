package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/requests"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/go-chi/render"
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
				body:       "<a href=\"/url\">Temporary Redirect</a>.",
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
			r := NewRouter(service)
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
			name: "post with location",
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://localhost:8080/id",
			},
			method: http.MethodPost,
			body:   "url",
		},
		{
			name: "post without location",
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
			mockRepo.On("Save", tt.body, "id").Return(nil)

			mockGen := new(mocks.MockGen)
			mockGen.On("GenerateIDFromString", "url").Return("id", nil)
			mockGen.On("GenerateIDFromString", "").Return("", errors.New("err"))
			mockGen.On("GenerateIDFromString", "error_on_shortening").Return("", errors.New("err"))

			cfg := config.New()
			service := services.New(mockRepo, mockGen, cfg)
			r := NewRouter(service)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, "/", tt.body)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)
		})
	}
}

func TestHandler_ShortenAPI(t *testing.T) {
	type want struct {
		statusCode int
		response   render.Renderer
	}

	tests := []struct {
		name    string
		want    want
		Req     string
		method  string
		request requests.ShortenURLRequest
	}{
		{
			name: "post with url",
			want: want{
				statusCode: http.StatusCreated,
				response:   responses.NewShortURLResponse(models.ShortURL{ID: "id"}),
			},
			method:  http.MethodPost,
			request: requests.ShortenURLRequest{OriginalURL: "url"},
		},
		{
			name: "post without url",
			want: want{
				statusCode: http.StatusBadRequest,
				response:   responses.ErrInvalidRequest(errors.New("missing required url field")),
			},
			method:  http.MethodPost,
			request: requests.ShortenURLRequest{},
		},
		{
			name: "it returns 500 when service fails on shortening",
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   responses.ErrInternal(errors.New("err")),
			},
			method:  http.MethodPost,
			request: requests.ShortenURLRequest{OriginalURL: "error_on_shortening"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockRepo)
			mockRepo.On("Save", tt.request.OriginalURL, "id").Return(nil)

			mockGen := new(mocks.MockGen)
			mockGen.On("GenerateIDFromString", "url").Return("id", nil)
			mockGen.On("GenerateIDFromString", "").Return("", errors.New("err"))
			mockGen.On("GenerateIDFromString", "error_on_shortening").Return("", errors.New("err"))

			cfg := config.New()

			service := services.New(mockRepo, mockGen, cfg)
			r := NewRouter(service)
			ts := httptest.NewServer(r)
			defer ts.Close()

			bodyJSON, err := json.Marshal(tt.request)
			require.NoError(t, err)

			result, body := testRequest(t, ts, tt.method, "/api/shorten", string(bodyJSON))
			defer result.Body.Close()

			expectedResponse, err := json.Marshal(tt.want.response)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, string(expectedResponse), body)
		})
	}
}
