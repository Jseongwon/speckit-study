package rabbit

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Dial(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil { return nil, nil, err }
	ch, err := conn.Channel()
	if err != nil { conn.Close(); return nil, nil, err }
	return conn, ch, nil
}

func PublishConfirm(ctx context.Context, ch *amqp.Channel, exchange, routingKey string, msg amqp.Publishing) error {
	// publisher confirms
	if err := ch.Confirm(false); err != nil { return err }
	acks := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, msg); err != nil {
		return err
	}
	select {
	case cf := <-acks:
		if !cf.Ack { return amqp.ErrClosed }
		return nil
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}
