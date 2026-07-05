package main

import (
	"os"
	"strconv"
	"time"
)

// FingguConfig holds every runtime setting for the service, sourced from
// environment variables so the tool behaves the same in Docker, systemd,
// or a plain `go run`.
type FingguConfig struct {
	Port                string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	PollInterval        time.Duration
	MetricRetention     int // number of samples kept per metric series
	CPUThresholdPercent float64
	MemThresholdPercent float64
	GoroutineThreshold  int
	SlackWebhookURL     string
	DiscordWebhookURL   string
	AlertCooldown       time.Duration
	Environment         string
}

// FingguLoadConfig reads configuration from environment variables, falling
// back to sane defaults so the tool runs out of the box with zero setup.
func FingguLoadConfig() *FingguConfig {
	return &FingguConfig{
		Port:                fingguFn_GetEnv("FINGGU_PORT", "8080"),
		RedisAddr:           fingguFn_GetEnv("FINGGU_REDIS_ADDR", "localhost:6379"),
		RedisPassword:       fingguFn_GetEnv("FINGGU_REDIS_PASSWORD", ""),
		RedisDB:             fingguFn_GetEnvInt("FINGGU_REDIS_DB", 0),
		PollInterval:        time.Duration(fingguFn_GetEnvInt("FINGGU_POLL_INTERVAL_SECONDS", 5)) * time.Second,
		MetricRetention:     fingguFn_GetEnvInt("FINGGU_METRIC_RETENTION", 500),
		CPUThresholdPercent: fingguFn_GetEnvFloat("FINGGU_CPU_THRESHOLD", 80.0),
		MemThresholdPercent: fingguFn_GetEnvFloat("FINGGU_MEM_THRESHOLD", 85.0),
		GoroutineThreshold:  fingguFn_GetEnvInt("FINGGU_GOROUTINE_THRESHOLD", 5000),
		SlackWebhookURL:     fingguFn_GetEnv("FINGGU_SLACK_WEBHOOK_URL", ""),
		DiscordWebhookURL:   fingguFn_GetEnv("FINGGU_DISCORD_WEBHOOK_URL", ""),
		AlertCooldown:       time.Duration(fingguFn_GetEnvInt("FINGGU_ALERT_COOLDOWN_SECONDS", 60)) * time.Second,
		Environment:         fingguFn_GetEnv("FINGGU_ENV", "development"),
	}
}

func fingguFn_GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func fingguFn_GetEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func fingguFn_GetEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return fallback
}
