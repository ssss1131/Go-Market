package internal

import (
	"os"
	"strconv"
)

type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
	SMTPHost     string
	SMTPPort     int
	SMTPFrom     string
}

func MustLoad() *Config {
	return &Config{
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "user.registered"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "notification-service"),
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getInt("SMTP_PORT", 1025),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@gomarket.local"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
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
