package handlers

import (
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

	if errDecode := json.NewDecoder(reader).Decode(&ids); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	go h.service.DeleteUrls(r.Context(), ids, userID)

	w.WriteHeader(http.StatusAccepted)
}
