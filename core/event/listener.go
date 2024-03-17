package event

import (
	"context"
	"fmt"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/audit"
	"go.uber.org/zap"
)

// ChanListener listens for audit logs and processes them blocking
// one event at a time
type ChanListener struct {
	logs      <-chan audit.Log
	processor *Processor
}

func NewChanListener(inputChan <-chan audit.Log, processor *Processor) *ChanListener {
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
	stdLogger := grpczap.Extract(ctx).With(zap.String("event", log.Action))
	switch log.Action {
	case audit.OrgCreatedEvent.String():
		if err := l.processor.EnsureDefaultPlan(ctx, log.OrgID); err != nil {
			stdLogger.Error(fmt.Errorf("EnsureDefaultPlan: %w", err).Error())
		}
	}
}
