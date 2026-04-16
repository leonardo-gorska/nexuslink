package output

import (
	"context"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

// EventConsumer defines the output port for consuming events in batch.
type EventConsumer interface {
	Consume(ctx context.Context, batchSize int, timeout time.Duration) ([]entity.ClickEvent, error)
	Ack(deliveryTag uint64) error
	Nack(deliveryTag uint64, requeue bool) error
}
