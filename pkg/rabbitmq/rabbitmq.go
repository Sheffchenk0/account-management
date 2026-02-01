package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn *amqp.Connection
}

func New(url string) (*RabbitMQ, error) {
	var (
		conn  *amqp.Connection
		err   error
		count int64
	)

	const maxReties = 5

	for {
		conn, err = amqp.Dial(url)
		if err == nil {
			// TODO: add logs
			break
		}

		count++
		if count > maxReties {
			return nil, fmt.Errorf("could not connect to rabbitmq after %d attempts: %w", maxReties, err)
		}

		time.Sleep(time.Duration(count) * time.Second)
	}

	return &RabbitMQ{Conn: conn}, nil
}

func (r *RabbitMQ) Close() error {
	return r.Conn.Close()
}
