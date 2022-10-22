// Package server is wrapper around built in http server.
package server

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
)

type Server interface {
	Run() error
	Shutdown() error
}

func New(config *config.Config, ipChecker services.IPCheckerInterface, service *services.Shortener) (Server, error) {
	if config.EnableHTTPS {
		return NewHTTPS(config, ipChecker, service)
	} else {
		return NewHTTP(config, ipChecker, service)
	}
}
