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

type HTTP struct {
	server *http.Server
}

// Run starts http server.
func (s *HTTP) Run() {
	log.Fatal(s.server.ListenAndServe())
}

func NewHTTP(config *config.Config, service *services.Shortener) (Server, error) {
	httpServer := &http.Server{
		Addr:              config.ServerAddress,
		Handler:           handlers.NewRouter(service, config),
		ReadHeaderTimeout: 1 * time.Second,
	}
	server := &HTTP{
		server: httpServer,
	}
	return server, nil
}
