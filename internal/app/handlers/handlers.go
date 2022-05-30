package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/belamov/ypgo-url-shortener/internal/app/services"

	"github.com/go-chi/chi/v5"
)

func NewRouter(service *services.Shortener) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	h := NewHandler(service)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.Expand)
		r.Post("/", h.Shorten)
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

	fullURL, err := h.service.Expand(uID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if fullURL == "" {
		http.Error(w, "cant find full url", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}
