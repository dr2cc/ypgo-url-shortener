package handlers

import (
	"compress/flate"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const UserIDCookieName = "shortener-user-id"

func NewRouter(service *services.Shortener, config *config.Config, generator random.UserIDGenerator) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service, config, generator)

	r.Get("/{id}", h.Expand)
	r.Post("/", h.Shorten)
	r.Post("/api/shorten", h.ShortenAPI)
	r.Get("/api/user/urls", h.UserURLs)

	return r
}

type Handler struct {
	Mux             *chi.Mux
	service         *services.Shortener
	crypto          crypto.Cryptographer
	userIDGenerator random.UserIDGenerator
}

func NewHandler(service *services.Shortener, config *config.Config, userIDGenerator random.UserIDGenerator) *Handler {
	cryptographer := crypto.GCMAESCryptographer{Key: config.EncryptionKey}
	return &Handler{
		Mux:             chi.NewMux(),
		service:         service,
		crypto:          &cryptographer,
		userIDGenerator: userIDGenerator,
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if string(url) == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	su, err := h.service.Shorten(string(url), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.addEncryptedUserIDToCookie(&w, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.service.FormatShortURL(su)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var v models.ShortURL

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(reader).Decode(&v); err != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	if v.OriginalURL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	su, err := h.service.Shorten(v.OriginalURL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.addEncryptedUserIDToCookie(&w, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := responses.ShorteningResult{Result: h.service.FormatShortURL(su)}

	out, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Expand(w http.ResponseWriter, r *http.Request) {
	uID := chi.URLParam(r, "id")

	shortURL, err := h.service.Expand(uID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if shortURL.OriginalURL == "" {
		http.Error(w, "cant find full url", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	http.Redirect(w, r, shortURL.OriginalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) UserURLs(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	URLs, err := h.service.GetUrlsCreatedBy(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(URLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	formattedURLs := make([]responses.UsersShortURL, 0)
	for _, URL := range URLs {
		formattedURLs = append(
			formattedURLs,
			responses.UsersShortURL{ShortURL: h.service.FormatShortURL(URL), OriginalURL: URL.OriginalURL},
		)
	}

	out, err := json.Marshal(formattedURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		return h.userIDGenerator.GenerateUserID()
	}

	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
	if err != nil {
		return h.userIDGenerator.GenerateUserID()
	}

	decryptedUserID, err := h.crypto.Decrypt(decodedCookie)
	if err != nil {
		return h.userIDGenerator.GenerateUserID()
	}

	return string(decryptedUserID)
}
