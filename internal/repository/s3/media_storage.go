package s3

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

type MediaStorage interface {
	Upload(file io.Reader, filename string, contentType string) (string, error)
	GetURL(key string) string
}

type localStackS3Storage struct {
	s3Client *s3.S3
	bucket   string
	endpoint string
	logger   *zap.Logger
}

func NewMediaStorage(endpoint, region, bucket string) (MediaStorage, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	
	logger.Info("initializing s3 storage",
		zap.String("endpoint", endpoint),
		zap.String("region", region),
		zap.String("bucket", bucket),
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
		s3Client: client,
		bucket:   bucket,
		endpoint: endpoint,
		logger:   logger,
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
