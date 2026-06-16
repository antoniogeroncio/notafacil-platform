// Package config loads backend configuration from the environment.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds all runtime configuration for the API.
type Config struct {
	MongoURI   string
	MongoDB    string
	HTTPAddr   string
	JWTSecret  string
	SessionTTL time.Duration
	InviteTTL  time.Duration
	AppBaseURL string
	SMTP       SMTPConfig
}

// SMTPConfig holds transactional e-mail credentials.
type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// Load reads configuration from environment variables, applying sane defaults
// for local development.
func Load() (Config, error) {
	cfg := Config{
		MongoURI:   env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:    env("MONGO_DB", "notafacil"),
		HTTPAddr:   env("HTTP_ADDR", ":8080"),
		JWTSecret:  env("JWT_SECRET", "dev-insecure-change-me"),
		AppBaseURL: env("APP_BASE_URL", "http://localhost:3000"),
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     env("SMTP_PORT", "587"),
			User:     os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     env("SMTP_FROM", "no-reply@notafacil.local"),
		},
	}

	var err error
	if cfg.SessionTTL, err = duration("SESSION_TTL", 30*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.InviteTTL, err = duration("INVITE_TTL", 48*time.Hour); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: invalid duration for %s: %w", key, err)
	}
	return d, nil
}
