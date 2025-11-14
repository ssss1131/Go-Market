package internal

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr     string
	PGURL        string
	KafkaBrokers string
}

func MustLoad() *Config {
	// Загружаем .env (если есть)
	_ = godotenv.Load(".env")

	cfg := &Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8081"),          // другой порт, чтобы не конфликтовать с user-service
		PGURL:        mustEnv("PG_URL"),                     // строка подключения к базе
		KafkaBrokers: getEnv("KAFKA_BROKERS", "kafka:9092"), // можно оставить как в user-service
	}

	return cfg
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing required env: %s", k)
	}
	return v
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
