package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
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
	for i := 0; i < 15; i++ {
		client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		bucketExists, err := client.BucketExists(ctx, bucketName)
		cancel()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if !bucketExists {
			if err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
		}

		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucketName + `/*"]}]}`

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		err = client.SetBucketPolicy(ctx, bucketName, policy)
		cancel()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		return &MinioClient{
			client:     client,
			bucketName: bucketName,
		}, nil
	}

	return nil, fmt.Errorf("Failed to start minio")
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
	presignedUrl, err := m.client.PresignedGetObject(
		ctx,
		m.bucketName,
		objectName,
		time.Duration(expiryHours)*time.Hour,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}

	url := strings.Replace(presignedUrl.String(), "minio:9000", "localhost:9000", 1)
	//url := presignedUrl.String()

	return url, nil
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

func (m *MinioClient) ObjectExists(ctx context.Context, objectName string) bool {
	if _, err := m.client.StatObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{}); err != nil {
		return false
	}

	return true
}
