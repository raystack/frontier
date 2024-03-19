package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Logger struct {
	service *Service
	ctx     context.Context
	orgID   string

	Now func() time.Time
}

func NewLogger(ctx context.Context, orgID string) *Logger {
	return &Logger{
		service: GetService(ctx),
		ctx:     ctx,
		orgID:   orgID,
		Now:     time.Now().UTC,
	}
}

func (s *Logger) Log(action EventName, target Target) error {
	return s.LogWithAttrs(action, target, nil)
}

func (s *Logger) LogWithAttrs(action EventName, target Target, attrs map[string]string) error {
	l := &Log{
		ID:        uuid.NewString(),
		OrgID:     s.orgID,
		Source:    s.service.source,
		Action:    action.String(),
		Target:    target,
		CreatedAt: s.Now(),
		Metadata:  map[string]string{},
	}

	// extract metadata
	if s.service.metadataExtractor != nil {
		md, ok := s.service.metadataExtractor(s.ctx)
		if ok {
			l.Metadata = md
		}
	}
	// merge existing metadata with attrs
	for k, v := range attrs {
		l.Metadata[k] = v
	}

	// extract actor
	if s.service.actorExtractor != nil {
		actor, ok := s.service.actorExtractor(s.ctx)
		if ok {
			l.Actor = actor
		}
	}
	return s.service.Create(s.ctx, l)
}
