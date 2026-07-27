package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type SMTPConfig struct {
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

type Config struct {
	Env         string
	HTTPAddr    string
	DatabaseURL string
	RedisURL    string
	CORSOrigins []string
	LogLevel    string
	LogFile     string
	JWT_SECRET  string
	SMTPConfig  *SMTPConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env.local", ".env")

	smtmpConfig := &SMTPConfig{
		SMTPHost: getEnv("SMTP_HOST", "localhost"),
		SMTPPort: getEnv("SMTP_PORT", "1025"),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", "noreply@dvapi.local"),
	}

	cfg := &Config{
		Env:         getEnv("APP_ENV", "development"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		CORSOrigins: splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFile:     getEnv("LOG_FILE", ".log/access.log"),
		JWT_SECRET:  generateRandomString(32),
		SMTPConfig:  smtmpConfig,
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRandomString(n int) string {
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
