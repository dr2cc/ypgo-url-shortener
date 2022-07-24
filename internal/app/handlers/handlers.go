package handlers

import (
	"compress/flate"
	"compress/gzip"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const UserIDCookieName = "shortener-user-id"

func NewRouter(service *services.Shortener, config *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service, config)

	r.Get("/{id}", h.Expand)
	r.Post("/", h.Shorten)
	r.Post("/api/shorten", h.ShortenAPI)
	r.Post("/api/shorten/batch", h.ShortenBatchAPI)
	r.Get("/api/user/urls", h.UserURLs)
	r.Get("/ping", h.Ping)
	r.Delete("/api/user/urls", h.DeleteUrls)

	return r
}

type Handler struct {
	Mux     *chi.Mux
	service *services.Shortener
	crypto  crypto.Cryptographer
}

func NewHandler(service *services.Shortener, config *config.Config) *Handler {
	cryptographer := crypto.GCMAESCryptographer{Key: config.EncryptionKey, Random: service.Random}
	return &Handler{
		Mux:     chi.NewMux(),
		service: service,
		crypto:  &cryptographer,
	}
}

func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

func (h *Handler) addEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
	encryptedUserID, err := h.crypto.Encrypt([]byte(userID))
	if err != nil {
		return err
	}

	encodedCookieValue := hex.EncodeToString(encryptedUserID)

	http.SetCookie(
		*w,
		&http.Cookie{
			Name:  UserIDCookieName,
			Value: encodedCookieValue,
		},
	)
	return nil
}

func (h *Handler) getUserID(r *http.Request) string {
	encodedCookie, err := r.Cookie(UserIDCookieName)
	if err != nil {
		return h.service.GenerateNewUserID()
	}

	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
	if err != nil {
		return h.service.GenerateNewUserID()
	}

	decryptedUserID, err := h.crypto.Decrypt(decodedCookie)
	if err != nil {
		return h.service.GenerateNewUserID()
	}

	return string(decryptedUserID)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
