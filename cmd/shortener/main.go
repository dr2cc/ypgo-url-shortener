package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	pb "github.com/belamov/ypgo-url-shortener/internal/app/proto"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err)
	}

	gen := &generator.HashGenerator{}
	repo := storage.GetRepo(cfg)

	randomGenerator := &random.TrulyRandomGenerator{}

	// ОТСЮДА НАЧИНАЕТСЯ ЦЕПОЧКА ОБРАБОТЧИКОВ
	// В скобках сам файл проекта:
	// services.New (internal\app\server\server.go) -->
	// server.NewHTTP (internal\app\server\http.go) или server.NewHTTPS (internal\app\server\https.go) -->
	// тут все обработчики!
	// handlers.NewRouter (internal\app\http_handlers\handlers.go)
	service := services.New(repo, gen, randomGenerator, cfg)

	ipChecker, err := services.NewIPChecker(cfg)
	if err != nil {
		log.Fatal().Err(err)
	}

	restServer, err := server.New(cfg, ipChecker, service)
	if err != nil {
		log.Fatal().Err(err)
	}

	cryptographer := &crypto.GCMAESCryptographer{
		Random: randomGenerator,
		Key:    cfg.EncryptionKey,
	}
	grpcServer, err := pb.NewGRPCServer(cfg, ipChecker, service, cryptographer)
	if err != nil {
		log.Fatal().Err(err)
	}

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(2) //nolint:gomnd
	go runServer(ctx, wg, restServer, "REST HTTP server")
	go runServer(ctx, wg, grpcServer, "GRPC server")
	wg.Wait()

	log.Info().Msg("trying to shutdown storage gracefully")

	errClose := repo.Close(context.Background()) //nolint:contextcheck
	if errClose != nil {
		log.Fatal().Err(errClose)
	} else {
		log.Info().Msg("storage closed gracefully")
	}

	log.Info().Msg("Goodbye")
}

func runServer(ctx context.Context, wg *sync.WaitGroup, server server.Server, serverName string) {
	log.Info().Msgf("%s started", serverName)

	go func() {
		<-ctx.Done()
		log.Info().Msgf("trying to shutdown %s gracefully", serverName)

		if errShutdown := server.Shutdown(); errShutdown != nil {
			log.Info().Msgf("%s server Shutdown: %v", serverName, errShutdown)
		} else {
			log.Info().Msgf("%s shutted down gracefully", serverName)
		}
		wg.Done()
	}()

	if errRun := server.Run(); errRun != http.ErrServerClosed && errRun != nil {
		log.Info().Msgf("%s could not have started: %v", serverName, errRun)
	}
}
