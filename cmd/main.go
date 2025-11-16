package main

import (
	"context"
	"nbf-s3/internal/config"
	grpc_server "nbf-s3/internal/presentation"
	"os"
	"os/signal"
	"syscall"

	cfgtools "github.com/hesoyamTM/nbf-auth/pkg/config"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
)

func main() {
	cfg := cfgtools.MustParseConfig[config.Config]()
	ctx, err := logger.SetupLogger(context.Background(), cfg.Env)
	if err != nil {
		panic(err)
	}

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(err)
	}

	log.Debug("Logger is working")

	gRPCserver, err := grpc_server.NewGrpcServer(ctx, cfg)
	if err != nil {
		panic(err)
	}

	go gRPCserver.MustStart(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	gRPCserver.MustStop(ctx)

	log.Info("Server is gracefully stopped")
}
