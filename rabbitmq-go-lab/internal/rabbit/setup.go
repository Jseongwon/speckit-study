package rabbit

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Topology struct {
	Exchange      string
	RetryExchange string
	DLX           string
	Queue         string
	RetryQueue    string
	DLQ           string
	RetryTTLms    int
}

func Declare(ch *amqp.Channel, t Topology) error {
	// exchanges
	if err := ch.ExchangeDeclare(t.Exchange, "direct", true, false, false, false, nil); err != nil { return err }
	if err := ch.ExchangeDeclare(t.RetryExchange, "direct", true, false, false, false, nil); err != nil { return err }
	if err := ch.ExchangeDeclare(t.DLX, "fanout", true, false, false, false, nil); err != nil { return err }

	// main queue: dead-letter to retry exchange
	mainArgs := amqp.Table{
		"x-dead-letter-exchange": t.RetryExchange,
	}
	if _, err := ch.QueueDeclare(t.Queue, true, false, false, false, mainArgs); err != nil { return err }
	if err := ch.QueueBind(t.Queue, "demo.info", t.Exchange, false, nil); err != nil { return err }

	// retry queue: TTL then dead-letter back to main exchange
	retryArgs := amqp.Table{
		"x-dead-letter-exchange": t.Exchange,
		"x-message-ttl": int32(t.RetryTTLms),
	}
	if _, err := ch.QueueDeclare(t.RetryQueue, true, false, false, false, retryArgs); err != nil { return err }
	if err := ch.QueueBind(t.RetryQueue, "demo.info", t.RetryExchange, false, nil); err != nil { return err }

	// DLQ bound to DLX
	if _, err := ch.QueueDeclare(t.DLQ, true, false, false, false, nil); err != nil { return err }
	if err := ch.QueueBind(t.DLQ, "", t.DLX, false, nil); err != nil { return err }

	return nil
}

func RetryCount(headers amqp.Table) int {
	// from x-death header injected by RabbitMQ on dead-lettering
	if v, ok := headers["x-death"]; ok {
		if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
			if m, ok := arr[0].(amqp.Table); ok {
				if cnt, ok := m["count"].(int64); ok { return int(cnt) }
				if cnt, ok := m["count"].(int32); ok { return int(cnt) }
				if cnt, ok := m["count"].(int); ok { return cnt }
			}
		}
	}
	return 0
}

func HeadersString(h amqp.Table) string {
	return fmt.Sprintf("%v", h)
}
