package output

import (
	"context"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

// EventPublisher defines the output port for publishing events.
type EventPublisher interface {
	Publish(ctx context.Context, event *entity.ClickEvent) error
}
