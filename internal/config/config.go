package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string

	SecurityEnabled bool
	JWTSecret       string

	GoogleClientID     string
	GoogleClientSecret string

	AWSEnabled         bool
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	GCPEnabled           bool
	GCPProjectID         string
	GCPCredentialsBase64 string

	AzureEnabled          bool
	AzureConnectionString string
}

func LoadConfig(envPath string) (*Config, error) {
	_ = godotenv.Load(envPath)
	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8000"),
		SecurityEnabled:    getEnv("SECURITY_ENABLED", "true") == "true",
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		AWSEnabled:         getEnv("AWS_ENABLED", "false") == "true",
		AWSRegion:          os.Getenv("AWS_REGION"),
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),

		GCPEnabled:            getEnv("GCP_ENABLED", "false") == "true",
		GCPProjectID:          os.Getenv("GCP_PROJECT_ID"),
		GCPCredentialsBase64:  os.Getenv("GCP_CREDENTIALS_BASE64"),
		AzureEnabled:          getEnv("AZURE_ENABLED", "false") == "true",
		AzureConnectionString: os.Getenv("AZURE_STORAGE_CONNECTION_STRING"),
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
