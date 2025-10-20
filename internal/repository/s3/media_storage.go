package s3

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type MediaStorage interface {
	Upload(file io.Reader, filename string, contentType string) (string, error)
	GetURL(key string) string
}

type localStackS3Storage struct {
	s3Client *s3.S3
	bucket   string
	endpoint string
}

func NewMediaStorage(endpoint, region, bucket string) (MediaStorage, error) {
	log.Printf("Initializing S3 with endpoint: %s, region: %s, bucket: %s", endpoint, region, bucket)

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		S3ForcePathStyle: aws.Bool(true), // Important for LocalStack
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)

	// Skip bucket existence check for now - just log and proceed
	log.Printf("S3 client initialized for bucket: %s", bucket)

	return &localStackS3Storage{
		s3Client: client,
		bucket:   bucket,
		endpoint: endpoint,
	}, nil
}

func (s *localStackS3Storage) Upload(file io.Reader, filename string, contentType string) (string, error) {
	key := fmt.Sprintf("posts/%d-%s", time.Now().Unix(), filename)
	log.Printf("Uploading file to S3: bucket=%s, key=%s, contentType=%s", s.bucket, key, contentType)

	// Read the file content into bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		log.Printf("S3 upload failed: %v", err)
		return "", err
	}

	log.Printf("S3 upload successful: key=%s", key)
	return key, nil
}

func (s *localStackS3Storage) GetURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
}
