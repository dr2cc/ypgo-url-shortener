package handlers

import (
	"crypto/aes"
	"encoding/json"
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

func TestHandler_Stats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockGen := mocks.NewMockURLGenerator(ctrl)
	mockRandom := mocks.NewMockGenerator(ctrl)
	cfg := &config.Config{
		BaseURL:       "http://localhost:8080",
		ServerAddress: ":8080",
		EncryptionKey: make([]byte, 2*aes.BlockSize),
	}

	t.Run("from trusted subnet", func(t *testing.T) {
		mockChecker := mocks.NewMockIPCheckerInterface(ctrl)
		mockChecker.EXPECT().IsRequestFromTrustedSubnet(gomock.Any()).Return(true, nil)

		service := services.New(mockRepo, mockGen, mockRandom, cfg)
		r := NewRouter(service, mockChecker, cfg)
		ts := httptest.NewServer(r)
		defer ts.Close()

		result, body := testRequest(t, ts, http.MethodGet, "/api/internal/stats", "", nil)
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
		expectedJSON, err := json.Marshal(models.Stats{})
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), body)
	})

	t.Run("not from trusted subnet", func(t *testing.T) {
		mockChecker := mocks.NewMockIPCheckerInterface(ctrl)
		mockChecker.EXPECT().IsRequestFromTrustedSubnet(gomock.Any()).Return(false, nil)

		service := services.New(mockRepo, mockGen, mockRandom, cfg)
		r := NewRouter(service, mockChecker, cfg)
		ts := httptest.NewServer(r)
		defer ts.Close()

		result, body := testRequest(t, ts, http.MethodGet, "/api/internal/stats", "", nil)
		defer result.Body.Close()

		assert.Equal(t, http.StatusForbidden, result.StatusCode)
		assert.Equal(t, "forbidden", body)
	})
}
