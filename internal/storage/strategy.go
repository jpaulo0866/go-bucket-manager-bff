package storage

import (
	"go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/models"
	"io"
)

type CloudProvider string

const (
	ProviderAWS   CloudProvider = "AWS"
	ProviderGCP   CloudProvider = "GCP"
	ProviderAzure CloudProvider = "AZURE"
)

type StorageStrategy interface {
	ListFiles(bucket string) ([]models.FileInfo, error)
	DownloadFile(bucket, file string) (io.ReadCloser, error)
	UploadFile(bucket, file string, data io.Reader) error
	PresignedURL(bucket, file string) (string, error)
}

type Strategies map[CloudProvider]StorageStrategy

func NewStrategies(cfg *config.Config) Strategies {
    strategies := make(Strategies)
    if cfg.AWSEnabled {
        if s3strat, err := NewS3Strategy(cfg); err == nil {
            strategies[ProviderAWS] = s3strat
        }
    }
    if cfg.GCPEnabled {
        if gcsstrat, err := NewGCSStrategy(cfg); err == nil {
            strategies[ProviderGCP] = gcsstrat
        }
    }
    if cfg.AzureEnabled {
        if azstrat, err := NewAzureStrategy(cfg); err == nil {
            strategies[ProviderAzure] = azstrat
        }
    }
    return strategies
}
