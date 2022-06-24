package handlers

import (
	"compress/flate"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"io"
	"net/http"
)

const cookieName = "shortener-user-id"

func NewRouter(service *services.Shortener) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.Expand)
		r.Post("/", h.Shorten)
		r.Post("/api/shorten", h.ShortenAPI)
	})
	return r
}

type Handler struct {
	Mux     *chi.Mux
	service *services.Shortener
}

func NewHandler(service *services.Shortener) *Handler {
	return &Handler{
		Mux:     chi.NewMux(),
		service: service,
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

	userId := h.getUserId(r)

	if userId == "" {
		userId = generateUserId()

		var encryptedUserId []byte

		err = encryptCookie([]byte(userId), encryptedUserId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.SetCookie(
			w,
			&http.Cookie{
				Name:     cookieName,
				Value:    string(encryptedUserId),
				Secure:   true,
				HttpOnly: true,
			},
		)
	}

	su, err := h.service.Shorten(string(url))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(su.GetShortURL()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func encryptCookie(src []byte, dst []byte) error {
	// будем использовать AES256, создав ключ длиной 32 байта
	key, err := generateRandom(2 * aes.BlockSize) // ключ шифрования
	if err != nil {
		return err
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return err
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return err
	}

	dst = aesgcm.Seal(nil, nonce, src, nil) // зашифровываем

	return nil
}

func generateUserId() string {
	return uuid.NewString()
}

func (h *Handler) getUserId(r *http.Request) string {
	encryptedCookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	return encryptedCookie.Value
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

	su, err := h.service.Shorten(v.OriginalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := responses.ShorteningResult{Result: su.GetShortURL()}

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

	fullURL, err := h.service.Expand(uID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if fullURL == "" {
		http.Error(w, "cant find full url", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}
