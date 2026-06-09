package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort    string
	DatabaseURL   string
	RedisAddr     string
	CacheTTL      time.Duration
	DefaultExpiry time.Duration
	RateLimitRPS  float64
	RateLimitBurst int
}

func Load() *Config {
	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "host=localhost user=urlshort password=urlshort dbname=urlshortener port=5432 sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		CacheTTL:      24 * time.Hour,
		DefaultExpiry: 12 * time.Hour,
		RateLimitRPS:  10.0 / 60.0,
		RateLimitBurst: 10,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
