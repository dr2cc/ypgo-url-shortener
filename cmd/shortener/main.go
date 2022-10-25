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
	pb "github.com/belamov/ypgo-url-shortener/internal/app/proto"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
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

	ipChecker, err := services.NewIPChecker(cfg)
	if err != nil {
		log.Print(err)
		return
	}

	restServer, err := server.New(cfg, ipChecker, service)
	if err != nil {
		log.Print(err)
		return
	}

	cryptographer := &crypto.GCMAESCryptographer{
		Random: randomGenerator,
		Key:    cfg.EncryptionKey,
	}
	grpcServer, err := pb.NewGRPCServer(cfg, ipChecker, service, cryptographer)
	if err != nil {
		log.Print(err)
		return
	}

	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 2) //nolint:gomnd
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	terminateCh := make(chan struct{})
	go func() {
		<-sigint
		close(terminateCh)
	}()

	go runRestServer(restServer, terminateCh, idleConnsClosed)
	go runGrpcServer(grpcServer, terminateCh)
	<-idleConnsClosed
	fmt.Println("Http Shutdown gracefully")
}

func runGrpcServer(grpcServer server.Server, sigint chan struct{}) {
	go func() {
		<-sigint
		if errShutdown := grpcServer.Shutdown(); errShutdown != nil {
			log.Printf("GRPC server Shutdown: %v", errShutdown)
		}
	}()

	if errRun := grpcServer.Run(); errRun != http.ErrServerClosed && errRun != nil {
		log.Printf("GRPC server ListenAndServe: %v", errRun)
	}
}

func runRestServer(srv server.Server, sigint chan struct{}, idleConnsClosed chan struct{}) {
	go func() {
		<-sigint
		if errShutdown := srv.Shutdown(); errShutdown != nil {
			log.Printf("HTTP server Shutdown: %v", errShutdown)
		}
		close(idleConnsClosed)
	}()

	if errRun := srv.Run(); errRun != http.ErrServerClosed && errRun != nil {
		log.Printf("HTTP server ListenAndServe: %v", errRun)
	}
}
