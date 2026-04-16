package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch   *amqp.Channel
	msgs <-chan amqp.Delivery
}

func NewConsumer(conn *Connection) (*Consumer, error) {
	ch, err := conn.Conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.Qos(
		500,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(
		"nexuslink.clicks", // queue
		"",                 // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return &Consumer{ch: ch, msgs: msgs}, nil
}

func (c *Consumer) Consume(ctx context.Context, batchSize int, timeout time.Duration) ([]entity.ClickEvent, error) {
	var batch []entity.ClickEvent
	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return batch, ctx.Err()
		case <-timeoutChan:
			return batch, nil
		case d, ok := <-c.msgs:
			if !ok {
				return batch, fmt.Errorf("channel closed")
			}

			var msg ClickEventMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				slog.Error("failed to unmarshal message, sending to DLQ", slog.Any("error", err))
				d.Nack(false, false)
				continue
			}

			t, _ := time.Parse("2006-01-02T15:04:05.000Z07:00", msg.Timestamp)
			if t.IsZero() {
				t = time.Now()
			}

			event := entity.ClickEvent{
				LinkHash:  msg.Data.LinkHash,
				IP:        msg.Data.IPAddress,
				UserAgent: msg.Data.UserAgent,
				Referer:   msg.Data.Referer,
				ClickedAt: t,
			}
			event.ID = int64(d.DeliveryTag)

			batch = append(batch, event)

			if len(batch) >= batchSize {
				return batch, nil
			}
		}
	}
}

func (c *Consumer) Ack(deliveryTag uint64) error {
	return c.ch.Ack(deliveryTag, true) // multiple = true -> acks all up to this tag
}

func (c *Consumer) Nack(deliveryTag uint64, requeue bool) error {
	return c.ch.Nack(deliveryTag, true, requeue) // multiple = true -> nacks all up to this tag
}

func (c *Consumer) Close() error {
	return c.ch.Close()
}
