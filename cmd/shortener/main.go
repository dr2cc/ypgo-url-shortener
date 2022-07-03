package main

import (
	"flag"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

func main() {
	cfg := config.New()

	cfg.Init()
	flag.Parse()

	gen := &generator.HashGenerator{}
	repo := storage.GetRepo(cfg)
	defer repo.Close()
	service := services.New(repo, gen, cfg)
	userIDGenerator := &random.UUIDGenerator{}
	srv := server.New(cfg, service, userIDGenerator)

	srv.Run()
}
