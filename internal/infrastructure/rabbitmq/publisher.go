package rabbitmq

import (
	"account-manager/pkg/rabbitmq"
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn         *rabbitmq.RabbitMQ
	exchangeName string
}

func NewPublisher(conn *rabbitmq.RabbitMQ, exchangeName string) (*Publisher, error) {
	ch, err := conn.Conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Publisher{
		conn,
		exchangeName,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, payload []byte) error {
	ch, err := p.conn.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
		Timestamp:    time.Now(),
	}
	err = ch.PublishWithContext(
		ctx,
		p.exchangeName,
		routingKey,
		false,
		false,
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}
