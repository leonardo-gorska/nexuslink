package messaging

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	Conn *amqp.Connection
}

func NewConnection(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return &Connection{Conn: conn}, nil
}

func (c *Connection) Close() error {
	return c.Conn.Close()
}

func SetupInfrastructure(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// DLX & DLQ
	err = ch.ExchangeDeclare("nexuslink.events.dlx", "topic", true, false, false, false, nil)
	if err != nil { return err }

	_, err = ch.QueueDeclare("nexuslink.clicks.dlq", true, false, false, false, nil)
	if err != nil { return err }

	err = ch.QueueBind("nexuslink.clicks.dlq", "click.created", "nexuslink.events.dlx", false, nil)
	if err != nil { return err }

	// Main Exchange & Queue
	err = ch.ExchangeDeclare("nexuslink.events", "topic", true, false, false, false, nil)
	if err != nil { return err }

	args := amqp.Table{
		"x-dead-letter-exchange":    "nexuslink.events.dlx",
		"x-dead-letter-routing-key": "click.created",
	}
	_, err = ch.QueueDeclare("nexuslink.clicks", true, false, false, false, args)
	if err != nil { return err }

	err = ch.QueueBind("nexuslink.clicks", "click.created", "nexuslink.events", false, nil)
	if err != nil { return err }

	return nil
}
