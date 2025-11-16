package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(
	endpoint string,
	accessKey string,
	secretKey string,
	useSSL bool,
	bucketName string,
) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	//дописать проверку подключения

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bucketExists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	if !bucketExists {
		if err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
		//log
	}
	//log

	return &MinioClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (m *MinioClient) Upload(ctx context.Context, objectName string, data io.Reader, fileSize int64, contentType string) error {
	if _, err := m.client.PutObject(
		ctx,
		m.bucketName,
		objectName,
		data,
		fileSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	); err != nil {
		return fmt.Errorf("failed to upload photo: %w", err)
	}

	return nil
}

func (m *MinioClient) GetPresignedUrl(ctx context.Context, objectName string, expiryHours int) (string, error) {
	url, err := m.client.PresignedGetObject(
		ctx,
		m.bucketName,
		objectName,
		time.Duration(expiryHours)*time.Hour,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}

	return url.String(), nil
}

func (m *MinioClient) Delete(ctx context.Context, objectName string) error {
	if err := m.client.RemoveObject(
		ctx,
		m.bucketName,
		objectName,
		minio.RemoveObjectOptions{},
	); err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	return nil
}
