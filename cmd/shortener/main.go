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

	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err)
	}

	gen := &generator.HashGenerator{}

	// Здесь выбор хранилища
	// Это финальная версия приложения и тут выбор между
	// NewPgRepository и NewFileRepository (нет map)
	repo := storage.GetRepo(cfg)

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

	// Вот это уровень iter9 (29.07.2025) или выше
	// restServer имеет type Server interface {
	// 								Run() error
	// 								Shutdown() error
	// 								}
	restServer, err := server.New(cfg, ipChecker, service)
	if err != nil {
		log.Fatal().Err(err)
	}

	cryptographer := &crypto.GCMAESCryptographer{
		Random: randomGenerator,
		Key:    cfg.EncryptionKey,
	}

	// Это не знаю еще какой iter
	grpcServer, err := pb.NewGRPCServer(cfg, ipChecker, service, cryptographer)
	if err != nil {
		log.Fatal().Err(err)
	}

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(2) //nolint:gomnd

	// здесь старт в двух горутинах
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
