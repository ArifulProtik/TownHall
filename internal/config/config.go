package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	AppEnv           string
	LogLevel         string
	AutoMigrate      bool
	DatabaseURL      string
	JWTSecret        string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	RedisURL         string
	UploadthingToken string
}

// IsProd reports whether the app runs in production (JSON logs, secure cookies).
func (c *Config) IsProd() bool {
	return c.AppEnv == "production"
}

// New loads configuration. Missing .env is fine when real env vars exist;
// missing required values are an error for main to report.
func New() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		AppEnv:           getEnv("APP_ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		UploadthingToken: getFirstEnv([]string{"UPLOADTHING_TOKEN", "UPLOADTHING_SECRET", "UPLOADTHING_API_KEY"}, ""),
	}
	var err error
	if cfg.AutoMigrate, err = getBool("AUTO_MIGRATE", false); err != nil {
		return nil, err
	}
	if cfg.DatabaseURL, err = mustGetEnv("DATABASE_URL"); err != nil {
		return nil, err
	}
	if cfg.JWTSecret, err = mustGetEnv("JWT_SECRET"); err != nil {
		return nil, err
	}
	if cfg.JWTAccessTTL, err = getDuration("JWT_ACCESS_TTL", 15*time.Minute); err != nil {
		return nil, err
	}
	if cfg.JWTRefreshTTL, err = getDuration("JWT_REFRESH_TTL", 30*24*time.Hour); err != nil {
		return nil, err
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getFirstEnv(keys []string, fallback string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return fallback
}

func getBool(key string, fallback bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("invalid boolean for %s: %w", key, err)
	}
	return b, nil
}

func mustGetEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required env var %s is not set", key)
	}
	return v, nil
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	return d, nil
}
