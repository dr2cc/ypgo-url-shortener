package handlers

import (
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_UserURLs(t *testing.T) {
	type want struct {
		body       []responses.UsersShortURL
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		userID  string
		method  string
		want    want
	}{
		{
			name: "get user's urls",
			want: want{
				body: []responses.UsersShortURL{
					{ShortURL: "http://localhost:8080/id", OriginalURL: "url"},
				},
				statusCode: http.StatusOK,
			},
			userID: "user id with urls",
		},
		{
			name: "get another user's urls ",
			want: want{
				body:       []responses.UsersShortURL{},
				statusCode: http.StatusNoContent,
			},
			userID: "user id without urls",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().GetUsersUrls(gomock.Any(), "user id with urls").Return([]models.ShortURL{{OriginalURL: "url", ID: "id"}}, nil).AnyTimes()
			mockRepo.EXPECT().GetUsersUrls(gomock.Any(), "user id without urls").Return([]models.ShortURL{}, nil).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateRandomBytes(12).Return(make([]byte, 12), nil).AnyTimes()

			cfg := &config.Config{
				BaseURL:       "http://localhost:8080",
				ServerAddress: ":8080",
				EncryptionKey: make([]byte, 2*aes.BlockSize),
			}

			service := services.New(mockRepo, mockGen, mockRandom, cfg)
			r := NewRouter(service, cfg)
			ts := httptest.NewServer(r)
			defer ts.Close()

			cryptographer := crypto.GCMAESCryptographer{Key: cfg.EncryptionKey, Random: mockRandom}
			encryptedCookieValue, _ := cryptographer.Encrypt([]byte(tt.userID))
			cookies := map[string]string{
				UserIDCookieName: hex.EncodeToString(encryptedCookieValue),
			}
			result, body := testRequest(t, ts, http.MethodGet, "/api/user/urls", "", cookies)
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			if tt.want.statusCode == http.StatusNoContent {
				assert.Equal(t, "", body)
				return
			}
			expectedJSON, err := json.Marshal(tt.want.body)
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), body)
		})
	}
}
