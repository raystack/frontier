package event

import (
	"context"

	"github.com/raystack/frontier/core/audit"
)

// ChanPublisher is a blocking audit event publisher
type ChanPublisher struct {
	target chan<- audit.Log
}

func NewChanPublisher(target chan<- audit.Log) *ChanPublisher {
	return &ChanPublisher{
		target: target,
	}
}

func (p *ChanPublisher) Publish(ctx context.Context, log audit.Log) {
	p.target <- log
}
