package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/newBa1t/grpc_server.git/internal/config"
	"github.com/newBa1t/grpc_server.git/internal/logger"
	"github.com/newBa1t/grpc_server.git/internal/repo"
	"github.com/newBa1t/grpc_server.git/internal/service"
	auth "github.com/newBa1t/grpc_server.git/protos/gen"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := godotenv.Load("local.env"); err != nil {
		log.Fatal(errors.Wrap(err, "Error loading .env file"))
	}

	var cfg config.AuthConfig

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "Error processing config"))
	}

	logger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error creating logger"))
	}

	ctx := context.Background()
	repository, err := repo.NewRepository(ctx, cfg.PostgreSQL)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error initializing repository"))
	}

	authSrv := service.NewServerAuth(cfg, repository, logger)

	grpcServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, authSrv)

	listen, err := net.Listen("tcp", cfg.Grpc.Port)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "Error initializing listener"))
	}

	go func() {
		logger.Infof("gRPC server started on %s", cfg.Grpc.Port)
		if err := grpcServer.Serve(listen); err != nil {
			logger.Fatal(errors.Wrap(err, "Error initializing server"))
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	logger.Info("Получен сигнал завершения, корректное завершение работы сервера gRPC...")
	grpcServer.GracefulStop()
}
