package server

import (
	"log"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
)

type Server struct {
	config  *config.Config
	service *services.Shortener
}

func (s *Server) Run() {
	r := handlers.NewRouter(s.service)

	httpServer := &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: r,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func New(config *config.Config, service *services.Shortener) *Server {
	return &Server{
		config:  config,
		service: service,
	}
}
