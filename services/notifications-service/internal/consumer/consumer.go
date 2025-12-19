package consumer

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"

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
	brokers string
	topic   string
	groupID string
	sender  *email.Sender
	reader  *kafka.Reader
}

func New(brokers, topic, groupID string, sender *email.Sender) *Consumer {
	return &Consumer{
		brokers: brokers,
		topic:   topic,
		groupID: groupID,
		sender:  sender,
	}
}

func (c *Consumer) ensureTopicExists(ctx context.Context) error {
	brokerList := strings.Split(c.brokers, ",")

	var conn *kafka.Conn
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		conn, err = kafka.Dial("tcp", brokerList[0])
		if err == nil {
			break
		}
		log.Printf("consumer: attempt %d - waiting for Kafka: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, "9092"))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             c.topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil && !strings.Contains(err.Error(), "Topic with this name already exists") {
		log.Printf("consumer: create topic warning: %v", err)
	}

	log.Printf("consumer: topic '%s' ready", c.topic)
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("consumer: initializing - brokers: %s, topic: %s", c.brokers, c.topic)

	if err := c.ensureTopicExists(ctx); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(c.brokers, ","),
		Topic:          c.topic,
		GroupID:        c.groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	})

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
				time.Sleep(time.Second)
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
