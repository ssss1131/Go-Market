package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"GoNotification/internal/email"

	"github.com/segmentio/kafka-go"
)

type UserRegisteredEvent struct {
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
}

type Consumer struct {
	reader *kafka.Reader
	sender *email.Sender
}

func New(brokers, topic, groupID string, sender *email.Sender) *Consumer {
	log.Printf("Initializing Kafka consumer - brokers: %s, topic: %s, groupID: %s", brokers, topic, groupID)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(brokers, ","),
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		reader: reader,
		sender: sender,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Println("consumer: started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			return c.reader.Close()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				log.Printf("consumer: read error: %v", err)
				continue
			}

			log.Printf("consumer: received message from topic %s, partition %d, offset %d", msg.Topic, msg.Partition, msg.Offset)
			c.handleMessage(msg)
		}
	}
}

func (c *Consumer) handleMessage(msg kafka.Message) {
	var event UserRegisteredEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("consumer: unmarshal error: %v", err)
		return
	}

	log.Printf("consumer: received event for user %d (%s)", event.UserID, event.Email)

	if err := c.sender.SendVerification(event.Email, event.Token, event.BaseURL); err != nil {
		log.Printf("consumer: send email error: %v", err)
		return
	}

	log.Printf("consumer: verification email sent to %s", event.Email)
}
