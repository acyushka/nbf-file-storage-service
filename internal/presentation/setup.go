package grpc_server

import (
	"context"
	"fmt"
	"nbf-file-storage-service/internal/config"
	"nbf-file-storage-service/internal/service"
	"nbf-file-storage-service/internal/storage"
	s3_v1 "nbf-file-storage-service/pkg/pb/gen"
	"net"

	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	server *grpc.Server
	host   string
	port   int
}

func NewGrpcServer(ctx context.Context, cfg *config.Config) (*GrpcServer, error) {
	const op = "grpc.NewGrpcServer"

	//init storage
	storageClient, err := storage.NewMinioClient(
		cfg.Minio.Endpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.UseSSL,
		cfg.Minio.BucketName,
	)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	//init service
	fileStorageService := service.NewMinioService(storageClient, cfg.PresignedUrl.ExpiryHours)

	//init server
	fileStorageServer := NewMinioServer(fileStorageService)

	//create grpc server
	logInterceptor, err := logger.NewLoggingInterceptor(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(logInterceptor))

	//init FileStorageService
	s3_v1.RegisterFileStorageServiceServer(server, fileStorageServer)

	//init Reflection
	reflection.Register(server)

	return &GrpcServer{
		server: server,
		host:   cfg.Host,
		port:   cfg.Port,
	}, nil
}

func (s *GrpcServer) MustStart(ctx context.Context) {
	const op = "grpc.MustStart"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	log.Info("grpc server is starting")

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		panic(fmt.Errorf("%s:%w", op, err))
	}

	log.Info("Server is started")

	if err := s.server.Serve(lis); err != nil {
		panic(fmt.Errorf("%s:%w", op, err))
	}

}

func (s *GrpcServer) MustStop(ctx context.Context) {
	const op = "grpc.MustStop"

	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	log.Info("grpc server is stopping")

	s.server.GracefulStop()
}
