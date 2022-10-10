package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Shorten(t *testing.T) {
	type want struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name   string
		body   string
		method string
		want   want
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
		{
			name: "it returns 409 when url already exists",
			want: want{
				statusCode: http.StatusConflict,
				body:       "http://localhost:8080/id",
			},
			method: http.MethodPost,
			body:   "existingURL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().GetUsersUrls(gomock.Any(), "user id").Return(nil, nil).AnyTimes()
			mockRepo.EXPECT().Save(gomock.Any(), models.ShortURL{
				OriginalURL: "existingURL",
				ID:          "id",
				CreatedByID: "user id",
			}).Return(storage.ErrNotUnique).AnyTimes()
			mockRepo.EXPECT().Save(gomock.Any(), models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
				CreatedByID: "user id",
			}).Return(nil).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)
			mockGen.EXPECT().GenerateIDFromString("url").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("existingURL").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("").Return("", errors.New("err")).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("error_on_shortening").Return("", errors.New("err")).AnyTimes()

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateNewUserID().Return("user id").AnyTimes()
			mockRandom.EXPECT().GenerateRandomBytes(12).Return(make([]byte, 12), nil).AnyTimes()

			cfg := config.New()
			cfg.BaseURL = "http://localhost:8080"

			service := services.New(mockRepo, mockGen, mockRandom, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, "/", tt.body, nil)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)

			result, body = testGzippedRequest(t, ts, tt.method, "/", tt.body)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)
			if tt.want.statusCode == http.StatusCreated {
				assert.NotEmpty(t, result.Header.Get("Set-Cookie"))
			}
		})
	}
}

func TestHandler_ShortenAPI(t *testing.T) {
	type want struct {
		body        string
		contentType string
		statusCode  int
	}
	tests := []struct {
		name   string
		body   string
		method string
		want   want
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
		{
			name: "it returns 409 when url already exists",
			want: want{
				statusCode: http.StatusConflict,
				body:       "{\"result\":\"http://localhost:8080/id\"}",
			},
			method: http.MethodPost,
			body:   "{\"url\":\"existingURL\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().Save(gomock.Any(), models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
				CreatedByID: "user id",
			}).Return(nil).AnyTimes()
			mockRepo.EXPECT().Save(gomock.Any(), models.ShortURL{
				OriginalURL: "existingURL",
				ID:          "id",
				CreatedByID: "user id",
			}).Return(storage.ErrNotUnique).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)
			mockGen.EXPECT().GenerateIDFromString("url").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("existingURL").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("").Return("", errors.New("err")).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("error_on_shortening").Return("", errors.New("err")).AnyTimes()

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateNewUserID().Return("user id").AnyTimes()
			mockRandom.EXPECT().GenerateRandomBytes(12).Return(make([]byte, 12), nil).AnyTimes()

			cfg := config.New()
			cfg.BaseURL = "http://localhost:8080"

			service := services.New(mockRepo, mockGen, mockRandom, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, "/api/shorten", tt.body, nil)
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
