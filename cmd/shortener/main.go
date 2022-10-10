package main

import (
	"context"
	"fmt"
	"log"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg := config.New()

	err := cfg.Init()
	if err != nil {
		log.Fatal(err)
	}

	gen := &generator.HashGenerator{}
	repo := storage.GetRepo(cfg)
	defer func(ctx context.Context, repo storage.Repository) {
		errClose := repo.Close(ctx)
		if errClose != nil {
			log.Fatal(errClose)
		}
	}(context.Background(), repo)

	randomGenerator := &random.TrulyRandomGenerator{}
	service := services.New(repo, gen, randomGenerator, cfg)

	srv, err := server.New(cfg, service)
	if err != nil {
		log.Print(err)
	} else {
		srv.Run()
	}
}
