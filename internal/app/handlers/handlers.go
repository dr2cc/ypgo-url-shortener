package handlers

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"io"
	"net/http"
	"strings"
)

func ShortenerHandler(service *services.Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			get(service, w, r)
		case http.MethodPost:
			add(service, w, r)
		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
	}
}

func add(service *services.Shortener, w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if string(url) == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	su, err := service.Shorten(string(url))
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

func get(service *services.Shortener, w http.ResponseWriter, r *http.Request) {
	uID := strings.TrimPrefix(r.URL.Path, "/")
	if uID == "" {
		http.Error(w, "{id} required", http.StatusBadRequest)
		return
	}
	fu, err := service.Expand(uID)
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
