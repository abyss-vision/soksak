package server

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all server configuration loaded from environment variables.
type Config struct {
	Port           string
	Host           string
	DatabaseURL    string
	ServeUI        bool
	DeploymentMode string
	JWTSecret      string
	StorageBaseDir string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
// DATABASE_URL is optional at this stage (Phase 1A allows running without DB).
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "3200"),
		Host:           getEnv("HOST", ""),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		ServeUI:        getEnvBool("SERVE_UI", false),
		DeploymentMode: getEnv("DEPLOYMENT_MODE", "dev"),
		JWTSecret:      os.Getenv("SOKSAK_AGENT_JWT_SECRET"),
		StorageBaseDir: getEnv("STORAGE_BASE_DIR", "./data/storage"),
	}
	return cfg, nil
}

// Addr returns the combined host:port address for the server to listen on.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
