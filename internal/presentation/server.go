package grpc_server

import (
	"bytes"
	"context"
	"nbf-s3/internal/models"
	"nbf-s3/internal/service"
	s3_v1 "nbf-s3/pkg/pb/gen/go"

	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MinioServer struct {
	s3_v1.UnimplementedFileStorageServiceServer
	service *service.MinioService
}

func NewMinioServer(service *service.MinioService) *MinioServer {
	return &MinioServer{
		service: service,
	}
}

func (s *MinioServer) UploadAvatar(ctx context.Context, req *s3_v1.UploadAvatarRequest) (*s3_v1.UploadAvatarResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to init logger")
	}

	if req.GetUserId() == "" {
		log.Error("Error: user_id is empty")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if len(req.FileData) == 0 {
		log.Error("Error: file_data is empty")
		return nil, status.Error(codes.InvalidArgument, "file_data is required")
	}

	fileReader := bytes.NewReader(req.FileData)

	url, err := s.service.UploadAvatar(ctx, req.GetUserId(), fileReader, req.GetFileName(), int64(len(req.FileData)), req.GetContentType())
	if err != nil {
		log.Error("Error: failed to upload avatar")
		return nil, status.Errorf(codes.Internal, "failed to upload avatar: %v", err)
	}

	log.Info("Avatar uploaded successfuly")

	return &s3_v1.UploadAvatarResponse{
		Url: url,
	}, nil
}

func (s *MinioServer) UploadPhotos(ctx context.Context, req *s3_v1.UploadPhotosRequest) (*s3_v1.UploadPhotosResponse, error) {
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to init logger")
	}

	if req.GetUserId() == "" {
		log.Error("Error: user_id is empty")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	photos := make([]models.PhotoData, len(req.Photos))
	for i, pbPhoto := range req.Photos {
		photos[i] = models.PhotoData{
			Data:        bytes.NewReader(pbPhoto.FileData),
			FileSize:    int64(len(pbPhoto.FileData)),
			FileName:    pbPhoto.FileName,
			ContentType: pbPhoto.ContentType,
		}
	}

	urls, err := s.service.UploadPhotos(ctx, req.GetUserId(), photos)
	if err != nil {
		log.Error("Error: failed to upload photos")
		return nil, status.Errorf(codes.Internal, "failed to upload photos: %v", err)
	}

	log.Info("All photos uploaded successfuly")

	return &s3_v1.UploadPhotosResponse{
		Urls: urls,
	}, nil
}
