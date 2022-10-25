// Package server is wrapper around built in http server.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	handlers "github.com/belamov/ypgo-url-shortener/internal/app/http_handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
)

type HTTP struct {
	server *http.Server
}

// Run starts http server.
func (s *HTTP) Run() error {
	return s.server.ListenAndServe()
}

func (s *HTTP) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func NewHTTP(config *config.Config, ipChecker services.IPCheckerInterface, service *services.Shortener) (Server, error) {
	httpServer := &http.Server{
		Addr:              config.ServerAddress,
		Handler:           handlers.NewRouter(service, ipChecker, config),
		ReadHeaderTimeout: 1 * time.Second,
	}
	server := &HTTP{
		server: httpServer,
	}
	return server, nil
}
