package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	cfg, err := config.New()
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
		return
	}

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1) //nolint:gomnd
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigint
		if errShutdown := srv.Shutdown(); errShutdown != nil {
			log.Printf("HTTP server Shutdown: %v", errShutdown)
		}
		close(idleConnsClosed)
	}()

	if errRun := srv.Run(); errRun != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", errRun)
		return
	}
	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")
}
