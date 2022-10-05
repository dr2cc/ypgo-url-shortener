package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

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

	shortURL, err := h.service.Shorten(r.Context(), string(url), userID)
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

	if errDecode := json.NewDecoder(reader).Decode(&v); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	if v.OriginalURL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	shortURL, err := h.service.Shorten(r.Context(), v.OriginalURL, userID)
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
