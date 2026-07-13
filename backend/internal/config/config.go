package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration parsed from environment variables.
type Config struct {
	// Kafka
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	// TimescaleDB
	DatabaseURL string
	DBMaxOpen   int
	DBMaxIdle   int
	DBMaxLifetime time.Duration

	// Batch flusher
	BatchSize    int
	BatchFlushInterval time.Duration

	// API server
	APIPort           string
	APIQueryTimeout   time.Duration
	APIReadTimeout    time.Duration
	APIWriteTimeout   time.Duration

	// WebSocket
	WSThrottleInterval time.Duration

	// Logging
	LogLevel string
}

// Load reads environment variables and returns a validated Config.
func Load() (*Config, error) {
	cfg := &Config{
		KafkaBrokers:       splitCSV(envOrDefault("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:         envOrDefault("KAFKA_TOPIC", "metrics"),
		KafkaGroupID:       envOrDefault("KAFKA_GROUP_ID", "nexus-ingestor"),
		DatabaseURL:        envOrDefault("DATABASE_URL", "postgres://nexus:nexus@localhost:5432/nexus?sslmode=disable"),
		DBMaxOpen:          envIntOrDefault("DB_MAX_OPEN", 25),
		DBMaxIdle:          envIntOrDefault("DB_MAX_IDLE", 25),
		DBMaxLifetime:      envDurationOrDefault("DB_MAX_LIFETIME", 5*time.Minute),
		BatchSize:          envIntOrDefault("BATCH_SIZE", 1000),
		BatchFlushInterval: envDurationOrDefault("BATCH_FLUSH_INTERVAL", 500*time.Millisecond),
		APIPort:            envOrDefault("API_PORT", "8080"),
		APIQueryTimeout:    envDurationOrDefault("API_QUERY_TIMEOUT", 2*time.Second),
		APIReadTimeout:     envDurationOrDefault("API_READ_TIMEOUT", 10*time.Second),
		APIWriteTimeout:    envDurationOrDefault("API_WRITE_TIMEOUT", 10*time.Second),
		WSThrottleInterval: envDurationOrDefault("WS_THROTTLE_INTERVAL", 250*time.Millisecond),
		LogLevel:           envOrDefault("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if len(c.KafkaBrokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS must not be empty")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("BATCH_SIZE must be > 0")
	}
	if c.BatchFlushInterval <= 0 {
		return fmt.Errorf("BATCH_FLUSH_INTERVAL must be positive")
	}
	if c.APIQueryTimeout <= 0 {
		return fmt.Errorf("API_QUERY_TIMEOUT must be positive")
	}
	return nil
}

// --- helpers ---

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if i > start {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
