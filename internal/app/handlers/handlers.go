package handlers

import (
	"compress/flate"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
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

	shortURL, err := h.service.Shorten(string(url), userID)
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		writeShorteningResult(w, h, shortURL, http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeShorteningResult(w, h, shortURL, http.StatusCreated)
}

func writeShorteningResult(w http.ResponseWriter, h *Handler, shortURL models.ShortURL, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	shortenedURL := h.service.FormatShortURL(shortURL.ID)
	if _, err := w.Write([]byte(shortenedURL)); err != nil {
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

	shortURL, err := h.service.Shorten(v.OriginalURL, userID)
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		writeShorteningAPIResult(w, h, shortURL, http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeShorteningAPIResult(w, h, shortURL, http.StatusCreated)
}

func writeShorteningAPIResult(w http.ResponseWriter, h *Handler, shortURL models.ShortURL, status int) {
	res := responses.ShorteningResult{Result: h.service.FormatShortURL(shortURL.ID)}

	out, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(out); err != nil {
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
			responses.UsersShortURL{ShortURL: h.service.FormatShortURL(URL.ID), OriginalURL: URL.OriginalURL},
		)
	}

	out, err := json.Marshal(formattedURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(out); err != nil {
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
	err := h.service.HealthCheck()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ShortenBatchAPI(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	var input []request

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(reader).Decode(&input); err != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	batch := make([]models.ShortURL, len(input))

	for i, shortURLInput := range input {
		if shortURLInput.OriginalURL == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			return
		}
		batch[i] = models.ShortURL{
			OriginalURL:   shortURLInput.OriginalURL,
			CorrelationID: shortURLInput.CorrelationID,
		}
	}

	userID := h.getUserID(r)

	shortURLBatches, err := h.service.ShortenBatch(batch, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := make([]responses.ShorteningBatchResult, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = responses.ShorteningBatchResult{
			CorrelationID: shortURLBatch.CorrelationID,
			ShortURL:      h.service.FormatShortURL(shortURLBatch.ID),
		}
	}

	out, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
