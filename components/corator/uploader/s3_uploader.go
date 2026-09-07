package uploader

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/luhtaf/corator/config"
)

// S3Uploader adalah implementasi uploader untuk S3 compatible storage.
type S3Uploader struct {
	client *s3.Client
	bucket string
	region string
}

// NewS3Uploader membuat instance baru dari S3Uploader.
func NewS3Uploader(cfg config.S3Config) (*S3Uploader, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		// Jika endpoint kosong, kembalikan default resolver
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat konfigurasi AWS: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Uploader{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}

// Upload mengunggah file ke bucket S3.
func (u *S3Uploader) Upload(ctx context.Context, fileReader io.Reader, uniqueFilename string) (string, error) {
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(uniqueFilename),
		Body:   fileReader,
	})

	if err != nil {
		return "", fmt.Errorf("gagal mengunggah objek ke S3: %w", err)
	}

	// Kembalikan S3 URI atau URL
	uploadPath := fmt.Sprintf("s3://%s/%s", u.bucket, uniqueFilename)
	return uploadPath, nil
}
