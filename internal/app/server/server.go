package server

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"log"
	"net/http"
)

type Server struct {
	config  config.Config
	service *services.Shortener
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.ShortenerHandler(s.service))
	return mux
}

func (s *Server) Run() {
	httpServer := &http.Server{
		Addr:    ":" + config.Port(),
		Handler: s.Handler(),
	}

	log.Fatal(httpServer.ListenAndServe())
}

func New(config config.Config, service *services.Shortener) *Server {
	return &Server{
		config:  config,
		service: service,
	}
}
