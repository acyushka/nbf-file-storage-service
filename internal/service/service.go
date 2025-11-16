package service

import (
	"context"
	"fmt"
	"io"
	"nbf-s3/internal/models"
	"nbf-s3/internal/storage"
	"path/filepath"
	"time"
)

type MinioService struct {
	storage     *storage.MinioClient
	expiryHours int
}

func NewMinioService(s3 *storage.MinioClient, expiryHours int) *MinioService {
	return &MinioService{
		storage:     s3,
		expiryHours: expiryHours,
	}
}

func (s *MinioService) UploadAvatar(ctx context.Context, userID string, data io.Reader, fileName string, fileSize int64, contentType string) (string, error) {
	extension := filepath.Ext(fileName)
	objectName := fmt.Sprintf("%s/avatar%s", userID, extension)

	_ = s.storage.Delete(ctx, objectName)

	if err := s.storage.Upload(ctx, objectName, data, fileSize, contentType); err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	url, err := s.storage.GetPresignedUrl(ctx, objectName, s.expiryHours)
	if err != nil {
		return url, fmt.Errorf("failed to get presigned url for avatar: %w", err)
	}

	return url, nil
}

func (s *MinioService) UploadPhotos(ctx context.Context, userID string, photos []models.PhotoData) ([]string, error) {
	if len(photos) == 0 || len(photos) > 5 {
		return nil, fmt.Errorf("invalid length slice of photos")
	}

	var URLs []string

	for i, photo := range photos {
		timestamp := time.Now().Unix()
		extension := filepath.Ext(photo.FileName)
		objectName := fmt.Sprintf("%s/form/%d_%d%s", userID, i, timestamp, extension)

		if err := s.storage.Upload(ctx, objectName, photo.Data, photo.FileSize, photo.ContentType); err != nil {
			return nil, fmt.Errorf("failed to upload photo %d: %w", i+1, err)
		}

		url, err := s.storage.GetPresignedUrl(ctx, objectName, s.expiryHours)
		if err != nil {
			return nil, fmt.Errorf("failed to get presigned url for photo %d: %w", i+1, err)
		}

		URLs = append(URLs, url)

	}

	return URLs, nil
}
