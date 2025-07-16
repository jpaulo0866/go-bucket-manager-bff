package test

import (
	"go-bucket-manager-bff/internal/api"
	"go-bucket-manager-bff/internal/auth"
	"go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/models"
	"go-bucket-manager-bff/internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	jwtManager := auth.NewJWTManager("testsecret")
	token, err := jwtManager.Generate("test@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	email, err := jwtManager.Validate(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", email)
	}
}

type mockStrategy struct{}

func (m *mockStrategy) ListFiles(bucket string) ([]models.FileInfo, error) {
	return []models.FileInfo{{FileName: "test.log", Size: 123, LastModified: models.FileInfo{}.LastModified}}, nil
}
func (m *mockStrategy) DownloadFile(bucket, file string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("logdata")), nil
}
func (m *mockStrategy) UploadFile(bucket, file string, data io.Reader) error { return nil }
func (m *mockStrategy) PresignedURL(bucket, file string) (string, error) {
	return "http://example.com", nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	cfg := &config.Config{JWTSecret: "testsecret"}
	jwt := auth.NewJWTManager(cfg.JWTSecret)
	strategies := storage.Strategies{
		storage.ProviderAWS: &mockStrategy{},
	}
	api.RegisterAPIRoutes(r, strategies, jwt, cfg)
	return r
}

func TestListFilesHandler(t *testing.T) {
	r := setupRouter()
	jwt := auth.NewJWTManager("testsecret")
	token, _ := jwt.Generate("test@example.com")
	req, _ := http.NewRequest("GET", "/api/v1/providers/AWS/buckets/test-bucket/files", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "test.log")
}

func TestJWTMiddlewareRejectsInvalidToken(t *testing.T) {
	r := setupRouter()
	req, _ := http.NewRequest("GET", "/api/v1/providers/AWS/buckets/test-bucket/files", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}
