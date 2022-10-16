package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string, cookies map[string]string) (*http.Response, string) {
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

	if len(cookies) > 0 {
		for name, value := range cookies {
			req.AddCookie(&http.Cookie{
				Name:     name,
				Value:    value,
				Secure:   true,
				HttpOnly: true,
			})
		}
	}
	resp, err = client.Do(req)
	require.NoError(t, err)

	respBody, err = io.ReadAll(resp.Body)
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

	if _, err = w.Write([]byte(body)); err != nil {
		return nil, err.Error()
	}

	if err = w.Close(); err != nil {
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

	respBody, err = io.ReadAll(resp.Body)
	defer resp.Body.Close()

	require.NoError(t, err)

	return resp, string(bytes.TrimSpace(respBody))
}

func TestHandler_getUserID(t *testing.T) {
	tests := []struct {
		name           string
		cookieRawValue string
		cookieValue    string
		want           string
	}{
		{
			name:           "it returns passed user id correctly",
			cookieRawValue: "user id",
			cookieValue:    "",
			want:           "user id",
		},
		{
			name:           "it generates new user id if no cookie was passed",
			cookieRawValue: "",
			cookieValue:    "",
			want:           "new user id",
		},
		{
			name:           "it generates new user id if no cookie was passed",
			cookieRawValue: "",
			cookieValue:    "75770aa0e5a26e916fc2c4bd76da15aba5284340f1fdea947d25a56dcf9b15354d3bda",
			want:           "new user id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockGen := mocks.NewMockURLGenerator(ctrl)

			mockRandom := mocks.NewMockGenerator(ctrl)
			mockRandom.EXPECT().GenerateNewUserID().Return("new user id").AnyTimes()
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

			req, err := http.NewRequest(http.MethodGet, "", nil)

			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			h := NewHandler(service, cfg)

			if tt.cookieRawValue != "" {
				encryptedCookieValue, errEncrypt := h.crypto.Encrypt([]byte(tt.cookieRawValue))
				encodedCookieValue := hex.EncodeToString(encryptedCookieValue)
				require.NoError(t, errEncrypt)
				req.AddCookie(&http.Cookie{
					Name:     UserIDCookieName,
					Value:    encodedCookieValue,
					Secure:   true,
					HttpOnly: true,
				})
			}

			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:     UserIDCookieName,
					Value:    tt.cookieValue,
					Secure:   true,
					HttpOnly: true,
				})
			}

			res := h.getUserID(req)

			assert.Equal(t, tt.want, res)
		})
	}
}
