package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"GoNotification/internal"
	"GoNotification/internal/consumer"
	"GoNotification/internal/email"
)

func main() {
	cfg := internal.MustLoad()

	sender := email.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	cons := consumer.New(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, sender)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := cons.Start(ctx); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	log.Println("notification-svc started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	log.Println("notification-svc stopped")
}
