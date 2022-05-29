package server

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

type Server struct {
	config  config.Config
	service *services.Shortener
}

func (s *Server) Run() {
	r := s.NewRouter()
	httpServer := &http.Server{
		Addr:    ":" + config.Port(),
		Handler: r,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func (s *Server) NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	h := handlers.NewHandler(s.service)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.Expand)
		r.Post("/", h.Shorten)
	})
	return r
}

func New(config config.Config, service *services.Shortener) *Server {
	return &Server{
		config:  config,
		service: service,
	}
}
