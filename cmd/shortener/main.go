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

// В таждую функцию file_repository добавить тестувую печать
// Вот это в отладочном json
// "program": "${workspaceFolder}/cmd/shortener/main.go"
// Позволяет отлаживать из любого каталога!

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Configuration🧹🏦
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err)
	}

	// Repository🧹🏦
	repo := storage.GetRepo(cfg)

	// Use case🧹🏦
	gen := &generator.HashGenerator{}

	randomGenerator := &random.TrulyRandomGenerator{}

	// ❗TODO: список главных структур services.Shortener (видимо аналог App в zha-go-clean-architecture)
	// реализует методы интерфейса services.ShortenerInterface
	// ❗другие ключевые структуры- handlers.Handler - models.ShortURL
	//
	// Здесь начало цепочки, следующий шаг- restServer
	// service имеет тип services.Shortener struct — основной сервис приложения
	service := services.New(repo, gen, randomGenerator, cfg)
	// ЦЕПОЧКА ОБРАБОТЧИКОВ
	//
	// services.New (internal\app\server\server.go) -->
	// server.NewHTTP (internal\app\server\http.go) -->
	// handlers.NewRouter (internal\app\http_handlers\handlers.go)

	ipChecker, err := services.NewIPChecker(cfg)
	if err != nil {
		log.Fatal().Err(err)
	}

	cryptographer := &crypto.GCMAESCryptographer{
		Random: randomGenerator,
		Key:    cfg.EncryptionKey,
	}

	// HTTP Server🧹🏦
	restServer, err := server.New(cfg, ipChecker, service)
	// restServer реализует тип server.Server interface {
	// 								Run() error
	// 								Shutdown() error
	// 								}
	if err != nil {
		log.Fatal().Err(err)
	}

	// GRPC Server🧹🏦
	grpcServer, err := pb.NewGRPCServer(cfg, ipChecker, service, cryptographer)
	if err != nil {
		log.Fatal().Err(err)
	}

	// Waiting signal🧹🏦
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(2) //nolint:gomnd

	// Старт в двух горутинах
	go runServer(ctx, wg, restServer, "REST HTTP server")
	go runServer(ctx, wg, grpcServer, "GRPC server")
	wg.Wait()

	// Shutdown🧹🏦
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
