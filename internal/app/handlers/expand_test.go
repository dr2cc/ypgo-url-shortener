package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Expand(t *testing.T) {
	type want struct {
		location   string
		body       string
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		method  string
		want    want
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
		{
			name: "it returns 410 when trying to expand deleted url",
			want: want{
				statusCode: http.StatusGone,
				location:   "",
				body:       "url is deleted",
			},
			request: "/deleted",
			method:  http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().GetByID(gomock.Any(), "id").Return(models.ShortURL{OriginalURL: "url"}, nil).AnyTimes()
			mockRepo.EXPECT().GetByID(gomock.Any(), "missing").Return(models.ShortURL{}, nil).AnyTimes()
			mockRepo.EXPECT().GetByID(gomock.Any(), "error").Return(models.ShortURL{}, errors.New("error text")).AnyTimes()
			mockRepo.EXPECT().GetByID(gomock.Any(), "deleted").Return(models.ShortURL{
				OriginalURL: "url",
				ID:          "deleted",
				CreatedByID: "user id",
				DeletedAt:   time.Now(),
			}, nil).AnyTimes()
			mockGen := mocks.NewMockURLGenerator(ctrl)

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateNewUserID().Return("user id").AnyTimes()

			cfg := config.New()
			service := services.New(mockRepo, mockGen, mockRandom, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, tt.request, "", nil)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}
