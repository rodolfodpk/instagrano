package s3

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type MediaStorage interface {
    Upload(file io.Reader, filename string, contentType string) (string, error)
    GetURL(key string) string
}

type localStackS3Storage struct {
    s3Client *s3.Client
    bucket   string
    endpoint string
}

func NewMediaStorage(endpoint, region, bucket string) (MediaStorage, error) {
    customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
        if service == s3.ServiceID {
            return aws.Endpoint{
                URI: endpoint,
            }, nil
        }
        return aws.Endpoint{}, &aws.EndpointNotFoundError{}
    })

    cfg, err := config.LoadDefaultAWSConfig(context.TODO(),
        config.WithRegion(region),
        config.WithEndpointResolverWithOptions(customResolver),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    client := s3.NewFromConfig(cfg)

    // Create bucket if it doesn't exist
    _, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
        Bucket: aws.String(bucket),
    })
    if err != nil && !isBucketAlreadyOwnedByYouError(err) {
        return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
    }

    return &localStackS3Storage{
        s3Client: client,
        bucket:   bucket,
        endpoint: endpoint,
    }, nil
}

func isBucketAlreadyOwnedByYouError(err error) bool {
    var apiErr interface{ ErrorCode() string }
    if ok := (err).(interface{ Unwrap() error }); ok != nil {
        if ok := (ok.Unwrap()).(interface{ ErrorCode() string }); ok != nil {
            apiErr = ok
        }
    }
    return apiErr != nil && apiErr.ErrorCode() == "BucketAlreadyOwnedByYou"
}

func (s *localStackS3Storage) Upload(file io.Reader, filename string, contentType string) (string, error) {
    key := fmt.Sprintf("posts/%d-%s", time.Now().Unix(), filename)

    _, err := s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String(contentType),
    })

    return key, err
}

func (s *localStackS3Storage) GetURL(key string) string {
    return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
}
