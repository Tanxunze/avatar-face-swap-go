package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	JWTSecret        string
	AdminPassword    string
	DatabaseURL      string
	StorageDir       string
	TencentSecretID  string
	TencentSecretKey string
	TencentRegion    string
}

func Load() *Config {
	// Load .env file (ignore error if not exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	return &Config{
		Port:             getEnv("PORT", "5001"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-key"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin123"),
		DatabaseURL:      getEnv("DATABASE_URL", "./data/app.db"),
		StorageDir:       getEnv("STORAGE_DIR", "./data/storage"),
		TencentSecretID:  getEnv("TENCENTCLOUD_SECRET_ID", ""),
		TencentSecretKey: getEnv("TENCENTCLOUD_SECRET_KEY", ""),
		TencentRegion:    getEnv("TENCENT_REGION", "ap-guangzhou"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
