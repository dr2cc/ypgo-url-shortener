package handlers

import (
	"bytes"
	"errors"
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenerHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name    string
		request string
		want    want
		body    string
		method  string
	}{
		{
			name: "post with body",
			want: want{
				contentType: "",
				statusCode:  http.StatusCreated,
				body:        "http://localhost:8080/id",
			},
			request: "/",
			method:  http.MethodPost,
			body:    "url",
		},
		{
			name: "post without body",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "url required",
			},
			request: "/",
			method:  http.MethodPost,
			body:    "",
		},
		{
			name: "get with existing id",
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusTemporaryRedirect,
				body:        "<a href=\"/url\">Temporary Redirect</a>.",
			},
			request: "/id",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "get with null id",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "{id} required",
			},
			request: "/",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "get with missing id",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "cant find full url",
			},
			request: "/missing",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "not supported method",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusMethodNotAllowed,
				body:        "Unsupported method",
			},
			request: "/asd",
			method:  http.MethodDelete,
			body:    "",
		},
		{
			name: "it returns 500 when service fails on expanding",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
				body:        "error text",
			},
			request: "/error",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "it returns 500 when service fails on shortening",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
				body:        "err",
			},
			request: "/error",
			method:  http.MethodPost,
			body:    "error_on_shortening",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			mockRepo := new(mocks.MockRepo)
			mockRepo.On("Save", tt.body, "id").Return(nil)
			mockRepo.On("GetByID", "id").Return("url", nil)
			mockRepo.On("GetByID", "").Return("", nil)
			mockRepo.On("GetByID", "missing").Return("", nil)
			mockRepo.On("GetByID", "error").Return("", errors.New("error text"))

			mockGen := new(mocks.MockGen)
			mockGen.On("GenerateIDFromString", "url").Return("id", nil)
			mockGen.On("GenerateIDFromString", "").Return("", errors.New("err"))
			mockGen.On("GenerateIDFromString", "error_on_shortening").Return("", errors.New("err"))

			cfg := config.New()

			service := services.New(mockRepo, mockGen, cfg)

			h := ShortenerHandler(service)
			h.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			resBody, _ := io.ReadAll(result.Body)
			assert.Equal(t, tt.want.body, string(bytes.TrimSpace(resBody)))

		})
	}
}
