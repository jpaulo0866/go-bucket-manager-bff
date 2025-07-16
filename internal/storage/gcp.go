package storage

import (
	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"context"
	"encoding/base64"
	"go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/models"
	"io"
	"time"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCSStrategy struct {
	client    *storage.Client
	clientCfg *jwt.Config
}

func NewGCSStrategy(cfg *config.Config) (*GCSStrategy, error) {
	credsJSON, err := base64.StdEncoding.DecodeString(cfg.GCPCredentialsBase64)
	if err != nil {
		return nil, err
	}

	googleCfg, _ := google.JWTConfigFromJSON(credsJSON)

	client, err := storage.NewClient(context.TODO(), option.WithCredentialsJSON(credsJSON))
	if err != nil {
		return nil, err
	}
	return &GCSStrategy{client: client, clientCfg: googleCfg}, nil
}

func (g *GCSStrategy) ListFiles(bucket string) ([]models.FileInfo, error) {
	ctx := context.TODO()
	it := g.client.Bucket(bucket).Objects(ctx, nil)
	files := []models.FileInfo{}
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, models.FileInfo{
			FileName:     objAttrs.Name,
			Size:         objAttrs.Size,
			LastModified: objAttrs.Updated,
		})
	}
	return files, nil
}

func (g *GCSStrategy) DownloadFile(bucket, file string) (io.ReadCloser, error) {
	ctx := context.TODO()
	rc, err := g.client.Bucket(bucket).Object(file).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (g *GCSStrategy) UploadFile(bucket, file string, data io.Reader) error {
	ctx := context.TODO()
	wc := g.client.Bucket(bucket).Object(file).NewWriter(ctx)
	if _, err := io.Copy(wc, data); err != nil {
		_ = wc.Close()
		return err
	}
	return wc.Close()
}

func (g *GCSStrategy) PresignedURL(bucket, file string) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
		Scheme:         storage.SigningSchemeV4,
		GoogleAccessID: g.clientCfg.Email,
		PrivateKey:     g.clientCfg.PrivateKey,
	}
	url, err := storage.SignedURL(bucket, file, opts)
	if err != nil {
		return "", err
	}
	return url, nil
}
