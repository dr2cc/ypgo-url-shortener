package server

import (
	"log"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
)

type Server struct {
	config      *config.Config
	service     *services.Shortener
	idGenerator random.UserIDGenerator
}

func (s *Server) Run() {
	r := handlers.NewRouter(s.service, s.config, s.idGenerator)

	httpServer := &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: r,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func New(config *config.Config, service *services.Shortener, generator *random.UUIDGenerator) *Server {
	return &Server{
		config:      config,
		service:     service,
		idGenerator: generator,
	}
}
