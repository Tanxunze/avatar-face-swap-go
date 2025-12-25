package config

import "os"

type Config struct {
	Port          string
	JWTSecret     string
	AdminPassword string
	DatabaseURL   string
	StorageDir    string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "5001"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-key"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		DatabaseURL:   getEnv("DATABASE_URL", "./data/app.db"),
		StorageDir:    getEnv("STORAGE_DIR", "./data/storage"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
