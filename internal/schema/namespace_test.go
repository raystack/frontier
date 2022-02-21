package schema

import (
	"context"
	"errors"
	"github.com/odpf/shield/model"
	"time"
)

func (s *ServiceTestSuite) TestUpdateNamespace_shouldThrowErrorOnStoreError() {
	ns := model.Namespace{
		Id:        "team2",
		Name:      "Team2",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	id := "team1"
	ctx := context.Background()
	expectedErr := errors.New("error updating namespace in store")
	s.store.On("UpdateNamespace", ctx, id, ns).Return(model.Namespace{}, expectedErr)

	actualNs, actualErr := s.service.UpdateNamespace(ctx, id, ns)

	s.Equal(model.Namespace{}, actualNs)
	s.ErrorIs(actualErr, expectedErr)
}

func (s *ServiceTestSuite) TestUpdateNamespace_shouldReturnUpdatedNamespace() {
	ns := model.Namespace{
		Id:        "team2",
		Name:      "Team2",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	id := "team1"
	ctx := context.Background()
	expectedNs := model.Namespace{
		Id:        "team1",
		Name:      "Team2",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	s.store.On("UpdateNamespace", ctx, id, ns).Return(expectedNs, nil)

	actualNs, actualErr := s.service.UpdateNamespace(ctx, id, ns)

	s.Equal(expectedNs, actualNs)
	s.Nil(actualErr)
}
