// Package server is wrapper around built in http server.
package server

import (
	"log"
	"net/http"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
)

type Server struct {
	config  *config.Config      // configuration object that we'll use to configure our server
	service *services.Shortener // service that will be used to shorten URLs
}

// Run starts http server.
func (s *Server) Run() {
	r := handlers.NewRouter(s.service, s.config)

	httpServer := &http.Server{
		Addr:              s.config.ServerAddress,
		Handler:           r,
		ReadHeaderTimeout: 1 * time.Second,
	}
	log.Fatal(httpServer.ListenAndServe())
}

// New creates new http server.
func New(config *config.Config, service *services.Shortener) *Server {
	return &Server{
		config:  config,
		service: service,
	}
}
