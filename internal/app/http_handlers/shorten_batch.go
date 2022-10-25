package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
)

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

	if errDecode := json.NewDecoder(reader).Decode(&input); errDecode != nil {
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

	shortURLBatches, err := h.service.ShortenBatch(r.Context(), batch, userID)
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
