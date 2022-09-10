package handlers

import (
	"context"
	"encoding/json"
	"net/http"
)

func (h *Handler) DeleteUrls(w http.ResponseWriter, r *http.Request) {
	var ids []string

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(reader).Decode(&ids); err != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	go h.service.DeleteUrls(context.Background(), ids, userID) //nolint:contextcheck

	w.WriteHeader(http.StatusAccepted)
}
