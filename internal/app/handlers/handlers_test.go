package handlers

import (
	"errors"
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenerHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
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
			},
			request: "/missing",
			method:  http.MethodGet,
			body:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			mockRepo := new(mocks.MockRepo)
			mockRepo.On("Save", tt.body, "id").Return(nil)
			mockRepo.On("GetById", "id").Return("url", nil)
			mockRepo.On("GetById", "").Return("", nil)
			mockRepo.On("GetById", "missing").Return("", nil)

			mockGen := new(mocks.MockGen)
			mockGen.On("GenerateIdFromString", "url").Return("id", nil)
			mockGen.On("GenerateIdFromString", "").Return("", errors.New("err"))

			cfg := config.New()
			service := services.New(mockRepo, mockGen, cfg)
			h := ShortenerHandler(service)
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
