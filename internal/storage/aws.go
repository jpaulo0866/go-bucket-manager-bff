package storage

import (
	"context"
	internalConfig "go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/models"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Strategy struct {
	client *s3.Client
}

func NewS3Strategy(cfg *internalConfig.Config) (*S3Strategy, error) {
	config := aws.NewConfig()
	config.Region = cfg.AWSRegion

	awsCfg := &aws.Config{
		Region:      cfg.AWSRegion,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
	}

	client := s3.NewFromConfig(*awsCfg)
	return &S3Strategy{client: client}, nil
}

func (s *S3Strategy) ListFiles(bucket string) ([]models.FileInfo, error) {
	resp, err := s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}
	files := make([]models.FileInfo, 0, len(resp.Contents))
	for _, obj := range resp.Contents {
		files = append(files, models.FileInfo{
			FileName:     *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
		})
	}
	return files, nil
}

func (s *S3Strategy) DownloadFile(bucket, file string) (io.ReadCloser, error) {
	resp, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *S3Strategy) UploadFile(bucket, file string, data io.Reader) error {
	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
		Body:   data,
	})
	return err
}

func (s *S3Strategy) PresignedURL(bucket, file string) (string, error) {
	psClient := s3.NewPresignClient(s.client)
	resp, err := psClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}
