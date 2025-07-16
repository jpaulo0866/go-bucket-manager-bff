package storage

import (
	"context"
	"go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/models"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type AzureStrategy struct {
	serviceClient *azblob.Client
}

func NewAzureStrategy(cfg *config.Config) (*AzureStrategy, error) {
	client, err := azblob.NewClientFromConnectionString(cfg.AzureConnectionString, nil)
	if err != nil {
		return nil, err
	}
	return &AzureStrategy{serviceClient: client}, nil
}

func (a *AzureStrategy) ListFiles(container string) ([]models.FileInfo, error) {
	ctx := context.TODO()
	containerClient := a.serviceClient.ServiceClient().NewContainerClient(container)
	pager := containerClient.NewListBlobsFlatPager(&azblob.ListBlobsFlatOptions{})
	files := []models.FileInfo{}

    for pager.More() {
        resp, err := pager.NextPage(ctx)
        if err != nil {
            return nil, err
        } else {
            for _, blob := range resp.Segment.BlobItems {
                files = append(files, models.FileInfo{
                    FileName:     *blob.Name,
                    Size:         *blob.Properties.ContentLength,
                    LastModified: *blob.Properties.LastModified,
                })
            }
        }
    }

	return files, nil
}

func (a *AzureStrategy) DownloadFile(container, file string) (io.ReadCloser, error) {
	ctx := context.TODO()
	blobClient := a.serviceClient.ServiceClient().NewContainerClient(container).NewBlobClient(file)
	resp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (a *AzureStrategy) UploadFile(container, file string, data io.Reader) error {
	ctx := context.TODO()
	blobClient := a.serviceClient.ServiceClient().NewContainerClient(container).NewBlockBlobClient(file)
	_, err := blobClient.UploadStream(ctx, data, nil)
	return err
}

func (a *AzureStrategy) PresignedURL(container, file string) (string, error) {
	// Generate a SAS token for the blob
	blobClient := a.serviceClient.ServiceClient().NewContainerClient(container).NewBlobClient(file)
	// Set permissions and expiry
	permissions := sas.BlobPermissions{Read: true}
	expiry := time.Now().UTC().Add(15 * time.Minute)
	sasUrl, err := blobClient.GetSASURL(permissions, expiry, nil)
	if err != nil {
		return "", err
	}
	return sasUrl, nil
}
