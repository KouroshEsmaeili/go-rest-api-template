package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort        = "8080"
	defaultEnvironment = "development"
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
)

// Config contains the process configuration loaded from environment variables.
type Config struct {
	Port        string
	Environment string
	DatabaseURL string
}

// Load reads and validates configuration. Local-development values have safe
// defaults; production deployments should always set their own DATABASE_URL.
func Load() (Config, error) {
	databaseURL, databaseURLSet := os.LookupEnv("DATABASE_URL")
	if !databaseURLSet {
		databaseURL = defaultDatabaseURL
	}

	cfg := Config{
		Port:        envOrDefault("PORT", defaultPort),
		Environment: envOrDefault("APP_ENV", defaultEnvironment),
		DatabaseURL: databaseURL,
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must not be empty")
	}
	if cfg.Environment == "production" && !databaseURLSet {
		return Config{}, fmt.Errorf("DATABASE_URL must be set in production")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
