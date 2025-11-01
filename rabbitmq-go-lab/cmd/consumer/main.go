package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"rabbitmq-go-lab/internal/config"
	"rabbitmq-go-lab/internal/rabbit"
)

type payload struct {
	MessageID string `json:"messageId"`
	Type      string `json:"type"`
	Version   int    `json:"version"`
	Payload   any    `json:"payload"`
}

func main() {
	_ = godotenv.Load()
	c := config.Load()

	conn, ch, err := rabbit.Dial(c.URL)
	if err != nil { log.Fatalf("dial: %v", err) }
	defer conn.Close(); defer ch.Close()

	if err := rabbit.Declare(ch, rabbit.Topology{
		Exchange: c.Exchange, RetryExchange: c.RetryExchange, DLX: c.DLX,
		Queue: c.Queue, RetryQueue: c.RetryQueue, DLQ: c.DLQ, RetryTTLms: c.RetryTTLms,
	}); err != nil { log.Fatalf("declare: %v", err) }

	if err := ch.Qos(10, 0, false); err != nil { log.Fatalf("qos: %v", err) }

	deliveries, err := ch.Consume(c.Queue, "go-consumer", false, false, false, false, nil)
	if err != nil { log.Fatalf("consume: %v", err) }

	log.Printf("consumer up. queue=%s maxRetries=%d", c.Queue, c.MaxRetries)

	for d := range deliveries {
		var p payload
		_ = json.Unmarshal(d.Body, &p)
		retries := rabbit.RetryCount(d.Headers)

		// simulate processing and conditional failure
		forceFail := false
		if v, ok := d.Headers["forceFail"].(bool); ok && v { forceFail = true }

		log.Printf("recv id=%s retries=%d headers=%v", p.MessageID, retries, rabbit.HeadersString(d.Headers))

		if forceFail && retries < c.MaxRetries {
			log.Printf("forcing failure. nack -> retry exchange")
			_ = d.Nack(false, false) // dead-letter to retry exchange via queue args
			continue
		}

		if forceFail && retries >= c.MaxRetries {
			log.Printf("max retries reached -> send to DLQ")
			// publish to DLX (DLQ is bound to DLX)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = ch.PublishWithContext(ctx, c.DLX, "", false, false, amqp.Publishing{
				ContentType: "application/json",
				DeliveryMode: amqp.Persistent,
				Body: d.Body,
				Headers: d.Headers,
				MessageId: p.MessageID,
			})
			cancel()
			_ = d.Ack(false)
			continue
		}

		// normal success
		log.Printf("OK id=%s", p.MessageID)
		_ = d.Ack(false)
	}
}
