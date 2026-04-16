package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
	"github.com/leonardo-gorska/nexuslink/pkg/metrics"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(conn *Connection) (*Publisher, error) {
	ch, err := conn.Conn.Channel()
	if err != nil {
		return nil, err
	}
	// Enable publisher confirms
	if err := ch.Confirm(false); err != nil {
		return nil, err
	}
	return &Publisher{ch: ch}, nil
}

type ClickEventMessage struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Version   int    `json:"version"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		LinkHash  string `json:"link_hash"`
		IPAddress string `json:"ip_address"`
		UserAgent string `json:"user_agent"`
		Referer   string `json:"referer"`
		RequestID string `json:"request_id"`
	} `json:"data"`
}

func (p *Publisher) Publish(ctx context.Context, event *entity.ClickEvent) error {
	msg := ClickEventMessage{
		EventID:   uuid.NewString(),
		EventType: "click.created",
		Version:   1,
		Timestamp: event.ClickedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	msg.Data.LinkHash = event.LinkHash
	msg.Data.IPAddress = event.IP
	msg.Data.UserAgent = event.UserAgent
	msg.Data.Referer = event.Referer
	msg.Data.RequestID = "" // We can pass this by expanding entity if needed, or omit it.

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = p.ch.PublishWithContext(ctx,
		"nexuslink.events", // exchange
		"click.created",    // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	if err != nil {
		metrics.EventsPublishedTotal.WithLabelValues("error").Inc()
		return fmt.Errorf("failed to publish message: %w", err)
	}

	metrics.EventsPublishedTotal.WithLabelValues("success").Inc()
	return nil
}
