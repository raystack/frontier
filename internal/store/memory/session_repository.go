package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/authenticate/session"
)

type SessionRepository struct {
	data map[string]*session.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		data: make(map[string]*session.Session),
	}
}

func (s *SessionRepository) Set(ctx context.Context, session *session.Session) error {
	s.data[session.ID.String()] = session
	return nil
}

func (s *SessionRepository) Get(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	if val, ok := s.data[id.String()]; ok {
		return val, nil
	}
	return nil, session.ErrNoSession
}

func (s *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(s.data, id.String())
	return nil
}

func (s *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	for _, sess := range s.data {
		if sess.ExpiresAt.Before(time.Now().UTC()) {
			delete(s.data, sess.ID.String())
		}
	}
	return nil
}

func (s *SessionRepository) UpdateValidity(ctx context.Context, id uuid.UUID, validity time.Duration) error {
	if val, ok := s.data[id.String()]; ok {
		val.ExpiresAt = val.ExpiresAt.Add(validity)
		return nil
	}
	return session.ErrNoSession
}
