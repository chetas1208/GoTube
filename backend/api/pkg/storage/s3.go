package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
}

func NewS3Storage(cfg Config) (*S3Storage, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("object storage endpoint is required")
	}

	if strings.TrimSpace(cfg.PublicEndpoint) == "" {
		cfg.PublicEndpoint = cfg.Endpoint
	}

	internalAWS, err := buildAWSConfig(normalizeEndpoint(cfg.Endpoint), cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, err
	}
	publicAWS, err := buildAWSConfig(normalizeEndpoint(cfg.PublicEndpoint), cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(internalAWS, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})
	presignClient := s3.NewPresignClient(s3.NewFromConfig(publicAWS, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	}))

	return &S3Storage{
		client:        client,
		presignClient: presignClient,
	}, nil
}

func buildAWSConfig(endpoint, accessKey, secretKey string) (aws.Config, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	region := "us-east-1"
	if isR2Endpoint(endpoint) {
		region = "auto"
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithEndpointResolverWithOptions(resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("load object storage config: %w", err)
	}

	return cfg, nil
}

func normalizeEndpoint(endpoint string) string {
	return strings.TrimRight(strings.TrimSpace(endpoint), "/")
}

func isR2Endpoint(endpoint string) bool {
	endpoint = strings.ToLower(endpoint)
	return strings.Contains(endpoint, "cloudflarestorage.com") || strings.Contains(endpoint, ".r2.dev")
}

func (s *S3Storage) PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *S3Storage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3Storage) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:          key,
		Size:         *out.ContentLength,
		ContentType:  aws.ToString(out.ContentType),
		LastModified: aws.ToTime(out.LastModified),
	}, nil
}

func (s *S3Storage) GeneratePresignedPutURL(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3Storage) GeneratePresignedGetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
