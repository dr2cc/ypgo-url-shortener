package handlers

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

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
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if string(url) == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	su, err := h.service.Shorten(string(url))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(su))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Expand(w http.ResponseWriter, r *http.Request) {
	uID := chi.URLParam(r, "id")

	fu, err := h.service.Expand(uID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fu == "" {
		http.Error(w, "cant find full url", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, fu, http.StatusTemporaryRedirect)
}
