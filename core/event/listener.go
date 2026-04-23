package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/raystack/frontier/core/audit"
)

// ChanListener listens for audit logs and processes them blocking
// one event at a time
type ChanListener struct {
	logs      <-chan audit.Log
	processor *Service
}

func NewChanListener(inputChan <-chan audit.Log, processor *Service) *ChanListener {
	return &ChanListener{
		logs:      inputChan,
		processor: processor,
	}
}

// Listen listens for audit logs and processes them
func (l *ChanListener) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case log := <-l.logs:
			// process log
			l.Process(ctx, log)
		}
	}
}

func (l *ChanListener) Process(ctx context.Context, log audit.Log) {
	switch log.Action {
	case audit.OrgCreatedEvent.String():
		if err := l.processor.EnsureDefaultPlan(ctx, log.OrgID); err != nil {
			slog.ErrorContext(ctx, fmt.Errorf("EnsureDefaultPlan: %w", err).Error(), "event", log.Action)
		}
	}
}
