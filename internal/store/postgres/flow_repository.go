package postgres

import (
	"errors"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/authenticate"
)

type FlowRepository struct {
	data map[string]*authenticate.Flow
}

// TODO(kushsharma): instead of inmemory, persist these models in db
func NewFlowRepository() *FlowRepository {
	return &FlowRepository{
		data: make(map[string]*authenticate.Flow),
	}
}

func (s *FlowRepository) Set(session *authenticate.Flow) error {
	s.data[session.ID.String()] = session
	return nil
}

func (s *FlowRepository) Get(id uuid.UUID) (*authenticate.Flow, error) {
	if val, ok := s.data[id.String()]; ok {
		return val, nil
	}
	return nil, errors.New("no session")
}

func (s *FlowRepository) Delete(id uuid.UUID) error {
	delete(s.data, id.String())
	return nil
}
