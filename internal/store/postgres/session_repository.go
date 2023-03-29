package postgres

import (
	"github.com/google/uuid"
	"github.com/odpf/shield/core/authenticate"
)

type SessionRepository struct {
	data map[string]*authenticate.Session
}

// TODO(kushsharma): instead of inmemory, persist these models in db
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		data: make(map[string]*authenticate.Session),
	}
}

func (s *SessionRepository) Set(session *authenticate.Session) error {
	s.data[session.ID.String()] = session
	return nil
}

func (s *SessionRepository) Get(id uuid.UUID) (*authenticate.Session, error) {
	if val, ok := s.data[id.String()]; ok {
		return val, nil
	}
	return nil, authenticate.ErrNoSession
}

func (s *SessionRepository) Delete(id uuid.UUID) error {
	delete(s.data, id.String())
	return nil
}
