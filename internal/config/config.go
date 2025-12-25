package config

import (
	"log"
	"os"
	"strings"

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

	// Keycloak OIDC configuration
	KeycloakClientID     string
	KeycloakClientSecret string
	KeycloakServerURL    string // OIDC well-known URL

	// Frontend URL for redirects
	FrontendBaseURL string

	// CORS allowed origins (comma-separated for multiple origins)
	CORSAllowedOrigins string

	// Environment: development, production
	Environment string
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

		// Keycloak
		KeycloakClientID:     getEnv("KEYCLOAK_CLIENT_ID", ""),
		KeycloakClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		KeycloakServerURL:    getEnv("KEYCLOAK_SERVER_URL", ""),

		// Frontend
		FrontendBaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:5173"),

		// CORS
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),

		// Environment
		Environment: getEnv("GIN_MODE", "debug"),
	}
}

// IsKeycloakEnabled returns true if Keycloak is configured
func (c *Config) IsKeycloakEnabled() bool {
	return c.KeycloakClientID != "" && c.KeycloakClientSecret != "" && c.KeycloakServerURL != ""
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "release" || c.Environment == "production"
}

// GetCORSOrigins returns the list of allowed CORS origins
func (c *Config) GetCORSOrigins() []string {
	if c.CORSAllowedOrigins == "" {
		return []string{"http://localhost:5173"}
	}
	origins := []string{}
	for _, origin := range splitAndTrim(c.CORSAllowedOrigins, ",") {
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		parts = append(parts, strings.TrimSpace(part))
	}
	return parts
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
