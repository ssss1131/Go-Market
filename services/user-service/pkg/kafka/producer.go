package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers string) *Producer {
	log.Printf("Initializing Kafka producer with brokers: %s", brokers)
	writer := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}
	return &Producer{writer: writer}
}

func (p *Producer) Send(ctx context.Context, topic string, key string, value interface{}) error {
	log.Printf("Sending event to topic %v", topic)
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})

	if err != nil {
		log.Printf("Failed to send message to topic %s: %v", topic, err)
		return err
	}

	log.Printf("Successfully sent event to topic %s (key: %s)", topic, key)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
