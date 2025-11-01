package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"rabbitmq-go-lab/internal/config"
	"rabbitmq-go-lab/internal/rabbit"
)

func main() {
	_ = godotenv.Load()
	c := config.Load()

	setup := flag.Bool("setup", false, "declare exchanges/queues/bindings")
	republishDLQ := flag.Bool("republish-dlq", false, "move messages from DLQ to main")
	limit := flag.Int("limit", 50, "republish limit")
	flag.Parse()

	conn, ch, err := rabbit.Dial(c.URL)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	if *setup {
		if err := rabbit.Declare(ch, rabbit.Topology{
			Exchange: c.Exchange, RetryExchange: c.RetryExchange, DLX: c.DLX,
			Queue: c.Queue, RetryQueue: c.RetryQueue, DLQ: c.DLQ, RetryTTLms: c.RetryTTLms,
		}); err != nil {
			log.Fatalf("declare: %v", err)
		}
		log.Println("topology set.")
		return
	}

	if *republishDLQ {
		for i := 0; i < *limit; i++ {
			msg, ok, err := ch.Get(c.DLQ, false)
			if err != nil {
				log.Fatalf("get dlq: %v", err)
			}
			if !ok {
				log.Println("DLQ empty.")
				break
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err = ch.PublishWithContext(ctx, c.Exchange, "demo.info", false, false, amqp.Publishing{
				ContentType:  msg.ContentType,
				DeliveryMode: amqp.Persistent,
				Body:         msg.Body,
				Headers:      msg.Headers,
				MessageId:    msg.MessageId,
			})
			cancel()
			if err != nil {
				_ = msg.Nack(false, true)
				log.Fatalf("republish err: %v", err)
			}
			_ = msg.Ack(false)
			fmt.Println("republished:", msg.MessageId)
		}
		return
	}

	flag.Usage()
}
