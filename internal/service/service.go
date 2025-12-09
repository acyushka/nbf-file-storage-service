package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/acyushka/nbf-file-storage-service/internal/models"
	"github.com/acyushka/nbf-file-storage-service/internal/storage"

	"github.com/google/uuid"
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
	photoUUID := uuid.New().String()
	extension := filepath.Ext(fileName)
	objectName := fmt.Sprintf("%s/photos/%s%s", userID, photoUUID, extension)

	if err := s.storage.Upload(ctx, objectName, data, fileSize, contentType); err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	publicURL, err := s.storage.GetPublicUrl(ctx, objectName)
	if err != nil {
		return "", fmt.Errorf("failed to get public url: %w", err)
	}

	return publicURL, nil
}

func (s *MinioService) UploadPhotos(ctx context.Context, userID string, photos []models.PhotoData) ([]string, error) {
	if len(photos) == 0 || len(photos) > 5 {
		return nil, fmt.Errorf("invalid length slice of photos")
	}

	var uuids []string

	for i, photo := range photos {
		photoUUID := uuid.New().String()
		extension := filepath.Ext(photo.FileName)
		objectName := fmt.Sprintf("%s/photos/%s%s", userID, photoUUID, extension)

		if err := s.storage.Upload(ctx, objectName, photo.Data, photo.FileSize, photo.ContentType); err != nil {
			return nil, fmt.Errorf("failed to upload photo %d: %w", i+1, err)
		}

		uuids = append(uuids, photoUUID+extension)

	}

	return uuids, nil
}

func (s *MinioService) GetPhotoURL(ctx context.Context, userID string, uuid string) (string, error) {
	objectName := fmt.Sprintf("%s/photos/%s", userID, uuid)

	if !s.storage.ObjectExists(ctx, objectName) {
		return "", fmt.Errorf("photo does not exists")
	}

	url, err := s.storage.GetPresignedUrl(ctx, objectName, s.expiryHours)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url for %s: %w", uuid, err)
	}

	return url, nil
}
