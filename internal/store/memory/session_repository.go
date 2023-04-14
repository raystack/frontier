package memory

import (
	"context"

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
