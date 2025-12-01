package internal

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr     string
	PGURL        string
	JWTSecret    string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	KafkaBrokers string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
}

func MustLoad() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		PGURL:        mustEnv("USER_PG_URL"),
		JWTSecret:    mustEnv("JWT_SECRET"),
		AccessTTL:    getDuration("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:   getDuration("REFRESH_TTL", 14*24*time.Hour),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "kafka:9092"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
	}

	validateConfig(cfg)
	return cfg
}

func validateConfig(cfg *Config) {
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		log.Fatal("ACCESS_TTL and REFRESH_TTL must be > 0")
	}
	if cfg.AccessTTL >= cfg.RefreshTTL {
		log.Fatalf("ACCESS_TTL (%s) must be less than REFRESH_TTL (%s)", cfg.AccessTTL, cfg.RefreshTTL)
	}
	if len(cfg.JWTSecret) < 16 {
		log.Fatal("JWT_SECRET is too short; use at least 16–32 characters")
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env: %s", k)
	}
	return v
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
