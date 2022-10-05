package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg := config.New()

	cfg.Init()
	flag.Parse()

	gen := &generator.HashGenerator{}
	repo := storage.GetRepo(cfg)
	defer repo.Close(context.Background())
	random := &random.TrulyRandomGenerator{}
	service := services.New(repo, gen, random, cfg)
	srv := server.New(cfg, service)

	srv.Run()
}
