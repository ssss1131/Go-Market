package internal

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr   string
	PGURL      string
	JWTSecret  string
	ProductURL string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func MustLoad() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		HTTPAddr:   getEnv("HTTP_ADDR", ":8083"),
		PGURL:      mustEnv("CART_PG_URL"),
		JWTSecret:  mustEnv("JWT_SECRET"),
		ProductURL: mustEnv("PRODUCT_URL"),
		AccessTTL:  getDuration("ACCESS_TTL", 15*time.Minute),
		RefreshTTL: getDuration("REFRESH_TTL", 14*24*time.Hour),
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
		log.Fatal("JWT_SECRET is too short")
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
