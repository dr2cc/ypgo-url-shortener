package handlers

import (
	"crypto/aes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//nolint:goconst
func TestHandler_ShortenBatchAPI(t *testing.T) {
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
				body:        "[{\"correlation_id\":\"corId1\",\"short_url\":\"http://localhost:8080/id1\"},{\"correlation_id\":\"corId2\",\"short_url\":\"http://localhost:8080/id2\"}]",
				contentType: "application/json",
			},
			method: http.MethodPost,
			body:   "[{\"correlation_id\":\"corId1\",\"original_url\":\"url1\"},{\"correlation_id\":\"corId2\",\"original_url\":\"url2\"}]",
		},
		{
			name: "post without url",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "url required",
			},
			method: http.MethodPost,
			body:   "[{\"correlation_id\":\"corId1\",\"original_url\":\"\"},{\"correlation_id\":\"cor2\",\"original_url\":\"url2\"}]",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().SaveBatch(gomock.Any(), []models.ShortURL{
				{
					OriginalURL:   "url1",
					ID:            "id1",
					CreatedByID:   "user id",
					CorrelationID: "corId1",
				},
				{
					OriginalURL:   "url2",
					ID:            "id2",
					CreatedByID:   "user id",
					CorrelationID: "corId2",
				},
			}).Return(nil).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)
			mockGen.EXPECT().GenerateIDFromString("url1").Return("id1", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("url2").Return("id2", nil).AnyTimes()

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateNewUserID().Return("user id").AnyTimes()
			mockRandom.EXPECT().GenerateRandomBytes(12).Return(make([]byte, 12), nil).AnyTimes()

			mockChecker := mocks.NewMockIPCheckerInterface(ctrl)

			cfg := &config.Config{
				BaseURL:       "http://localhost:8080",
				ServerAddress: ":8080",
				EncryptionKey: make([]byte, 2*aes.BlockSize),
			}

			service := services.New(mockRepo, mockGen, mockRandom, cfg)
			r := NewRouter(service, mockChecker, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			result, body := testRequest(t, ts, tt.method, "/api/shorten/batch", tt.body, nil)
			defer result.Body.Close()

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)

			result, body = testGzippedRequest(t, ts, tt.method, "/api/shorten/batch", tt.body)
			defer result.Body.Close()

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.body, body)
		})
	}
}
