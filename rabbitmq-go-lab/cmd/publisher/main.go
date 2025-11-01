// cmd/publisher/main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
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
	timeout := flag.Duration("timeout", 30*time.Second, "per-message publish+confirm timeout")
	flag.Parse()

	// 1) Dial & Channel
	conn, ch, err := rabbit.Dial(c.URL)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	// 2) Declare topology (idempotent)
	if err := rabbit.Declare(ch, rabbit.Topology{
		Exchange: c.Exchange, RetryExchange: c.RetryExchange, DLX: c.DLX,
		Queue: c.Queue, RetryQueue: c.RetryQueue, DLQ: c.DLQ, RetryTTLms: c.RetryTTLms,
	}); err != nil {
		log.Fatalf("declare: %v", err)
	}

	// 3) Confirm mode ON (채널당 딱 한 번)
	if err := ch.Confirm(false); err != nil {
		log.Fatalf("confirm mode: %v", err)
	}

	// in-flight 수보다 충분히 큰 버퍼 권장
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1024))

	// (선택) 미라우팅 감지하고 싶으면 mandatory=true로 퍼블리시하고 returnsCh 처리
	// returnsCh := ch.NotifyReturn(make(chan amqp.Return, 64))

	for i := 0; i < *count; i++ {
		evt := types.DemoEvent{
			MessageID: "msg-" + time.Now().Format("20060102T150405.000000000"),
			Type:      "demo",
			Version:   1,
			Payload:   map[string]any{"n": i, "ts": time.Now().Format(time.RFC3339Nano)},
		}
		body, _ := json.Marshal(evt)

		headers := amqp.Table{"schemaVersion": 1}
		if *fail {
			headers["forceFail"] = true
		}

		msg := amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      headers,
			MessageId:    evt.MessageID,
			Type:         evt.Type,
		}

		// 4) Publish
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		err := ch.PublishWithContext(
			ctx,
			c.Exchange,
			*key,
			false, // mandatory (미라우팅 감지하려면 true로 바꾸고 returnsCh 처리)
			false, // immediate (deprecated)
			msg,
		)
		cancel()
		if err != nil {
			log.Fatalf("publish err: %v", err)
		}

		// 5) Confirm 대기
		select {
		case cfm := <-confirms:
			if !cfm.Ack {
				log.Fatalf("publish nack (deliveryTag=%d)", cfm.DeliveryTag)
			}
			log.Printf("published %s key=%s (confirmed)", evt.MessageID, *key)

		case <-time.After(*timeout):
			log.Fatalf("publish confirm err: context deadline exceeded")
		}

		// (선택) mandatory=true일 때 미라우팅 감지
		// select {
		// case r := <-returnsCh:
		// 	log.Fatalf("unroutable: code=%d text=%s key=%s", r.ReplyCode, r.ReplyText, r.RoutingKey)
		// default:
		// }
	}
}
