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
	client       *minio.Client
	publicClient *minio.Client
	bucketName   string
}

func NewMinioClient(
	endpoint string,
	publicEndpoint string,
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
			fmt.Printf("[minio] connect error: %v\n", err)
			continue
		}

		ctx := context.Background()
		bucketExists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			fmt.Printf("[minio] bucket exists error: %v\n", err)
			continue
		}

		if !bucketExists {
			if err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
				fmt.Printf("[minio] make bucket error: %v\n", err)
				continue
			}
		}

		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucketName + `/*"]}]}`

		err = client.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			fmt.Printf("[minio] set policy error: %v\n", err)
			continue
		}

		publicClient, err := minio.New(publicEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create public minio client: %w", err)
		}

		return &MinioClient{
			client:       client,
			publicClient: publicClient,
			bucketName:   bucketName,
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
	presignedUrl, err := m.publicClient.PresignedGetObject(
		ctx,
		m.bucketName,
		objectName,
		time.Duration(expiryHours)*time.Hour,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}

	url := presignedUrl.String()

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
