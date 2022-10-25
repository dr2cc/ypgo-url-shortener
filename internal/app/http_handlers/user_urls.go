package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
)

func (h *Handler) UserURLs(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	URLs, err := h.service.GetUrlsCreatedBy(r.Context(), userID)
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
