package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rodolfodpk/instagrano/internal/webclient"
	"go.uber.org/zap"
)

type MediaStorage interface {
	Upload(file io.Reader, filename string, contentType string) (string, error)
	UploadFromURL(url string) (string, string, error) // NEW: returns (key, contentType, error)
	GetURL(key string) string
}

type localStackS3Storage struct {
	s3Client   *s3.S3
	bucket     string
	endpoint   string
	logger     *zap.Logger
	httpClient webclient.HTTPClient
}

func NewMediaStorage(endpoint, region, bucket string, webclientConfig webclient.Config) (MediaStorage, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("initializing s3 storage",
		zap.String("endpoint", endpoint),
		zap.String("region", region),
		zap.String("bucket", bucket),
		zap.Bool("use_mock_controller", webclientConfig.UseMockController),
	)

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		S3ForcePathStyle: aws.Bool(true), // Important for LocalStack
	})
	if err != nil {
		logger.Error("failed to create aws session", zap.Error(err))
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)

	// Skip bucket existence check for now - just log and proceed
	logger.Info("s3 client initialized", zap.String("bucket", bucket))

	return &localStackS3Storage{
		s3Client:   client,
		bucket:     bucket,
		endpoint:   endpoint,
		logger:     logger,
		httpClient: webclient.NewDefaultHTTPClient(webclientConfig),
	}, nil
}

func (s *localStackS3Storage) Upload(file io.Reader, filename string, contentType string) (string, error) {
	key := fmt.Sprintf("posts/%d-%s", time.Now().Unix(), filename)
	s.logger.Info("uploading file to s3",
		zap.String("bucket", s.bucket),
		zap.String("key", key),
		zap.String("content_type", contentType),
		zap.String("filename", filename),
	)

	// Read the file content into bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("failed to read file", zap.Error(err))
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		s.logger.Error("s3 upload failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return "", err
	}

	s.logger.Info("s3 upload successful",
		zap.String("key", key),
		zap.Int("size_bytes", len(fileBytes)),
	)
	return key, nil
}

func (s *localStackS3Storage) GetURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
}

// UploadFromURL downloads media from a URL and uploads it to S3
func (s *localStackS3Storage) UploadFromURL(url string) (string, string, error) {
	s.logger.Info("downloading media from URL", zap.String("url", url))

	// Download from URL using webclient
	result, err := s.httpClient.Download(context.Background(), url)
	if err != nil {
		s.logger.Error("failed to download from URL", zap.String("url", url), zap.Error(err))
		return "", "", fmt.Errorf("failed to download from URL: %w", err)
	}
	defer result.Content.(io.ReadCloser).Close()

	// Generate filename from URL
	filename := filepath.Base(url)
	if filename == "" || filename == "." {
		filename = fmt.Sprintf("media-%d", time.Now().Unix())
	}

	s.logger.Info("downloaded media successfully",
		zap.String("url", url),
		zap.String("content_type", result.ContentType),
		zap.String("filename", filename),
		zap.Int64("size", result.Size),
	)

	// Upload to S3
	key, err := s.Upload(result.Content, filename, result.ContentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload downloaded media: %w", err)
	}

	return key, result.ContentType, nil
}
