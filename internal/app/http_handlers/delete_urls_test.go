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

func TestHandler_DeleteUrls(t *testing.T) {
	tests := []struct {
		name             string
		ids              string
		wantResponseCode int
	}{
		{
			name:             "it accepts urls to delete",
			ids:              "[\"id1\", \"id2\"]",
			wantResponseCode: http.StatusAccepted,
		},
		{
			name:             "it accepts urls to delete",
			ids:              "[\"\", \"\"]",
			wantResponseCode: http.StatusAccepted,
		},
		{
			name:             "it responses with error when request is not valid",
			ids:              "id1, id2",
			wantResponseCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().DeleteUrls(gomock.Any(), []models.ShortURL{{ID: "id2", CreatedByID: "new user id"}, {ID: "id1", CreatedByID: "new user id"}}).AnyTimes()
			mockRepo.EXPECT().DeleteUrls(gomock.Any(), []models.ShortURL{{ID: "id1", CreatedByID: "new user id"}, {ID: "id2", CreatedByID: "new user id"}}).AnyTimes()
			mockRepo.EXPECT().DeleteUrls(gomock.Any(), []models.ShortURL{{CreatedByID: "new user id"}, {CreatedByID: "new user id"}}).AnyTimes()
			mockGen := mocks.NewMockURLGenerator(ctrl)

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateNewUserID().Return("new user id").AnyTimes()
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

			result, _ := testRequest(t, ts, http.MethodDelete, "/api/user/urls", tt.ids, nil)
			defer result.Body.Close()

			assert.Equal(t, tt.wantResponseCode, result.StatusCode)
		})
	}
}
