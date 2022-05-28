package main

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

func main() {
	cfg := config.New()
	repo := storage.NewInMemoryRepository()
	gen := &generator.HashGenerator{}
	service := services.New(repo, gen, cfg)
	srv := server.New(cfg, service)

	srv.Run()
}
