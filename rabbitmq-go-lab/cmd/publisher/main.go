package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"rabbitmq-go-lab/internal/config"
	"rabbitmq-go-lab/internal/rabbit"
	"rabbitmq-go-lab/internal/types"
)

func main() {
	_ = godotenv.Load()
	c := config.Load()

	count := flag.Int("count", 1, "messages to publish")
	key := flag.String("key", "demo.info", "routing key")
	fail := flag.Bool("fail", false, "set header to force failure path")
	flag.Parse()

	conn, ch, err := rabbit.Dial(c.URL)
	if err != nil { log.Fatalf("dial: %v", err) }
	defer conn.Close(); defer ch.Close()

	if err := rabbit.Declare(ch, rabbit.Topology{
		Exchange: c.Exchange, RetryExchange: c.RetryExchange, DLX: c.DLX,
		Queue: c.Queue, RetryQueue: c.RetryQueue, DLQ: c.DLQ, RetryTTLms: c.RetryTTLms,
	}); err != nil { log.Fatalf("declare: %v", err) }

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < *count; i++ {
		evt := types.DemoEvent{
			MessageID: fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Type: "demo",
			Version: 1,
			Payload: map[string]any{"n": i, "ts": time.Now().Format(time.RFC3339Nano)},
		}
		body, _ := json.Marshal(evt)

		headers := amqp.Table{"schemaVersion": 1}
		if *fail { headers["forceFail"] = true }

		msg := amqp.Publishing{
			ContentType: "application/json",
			DeliveryMode: amqp.Persistent,
			Body: body,
			Headers: headers,
			MessageId: evt.MessageID,
			Type: evt.Type,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := rabbit.PublishConfirm(ctx, ch, c.Exchange, *key, msg); err != nil {
			log.Fatalf("publish confirm err: %v", err)
		}
		cancel()
		log.Printf("published %s key=%s", evt.MessageID, *key)
	}
}
