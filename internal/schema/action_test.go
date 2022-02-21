package schema

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

func (s *ServiceTestSuite) TestUpdateAction_ShouldThrowErrorOnStoreError() {
	id := "456"
	action := model.Action{Id: "123", Name: "Read", NamespaceId: "team"}
	ctx := context.Background()
	expectedErr := errors.New("error updating action in store")
	s.store.On("UpdateAction", ctx, model.Action{
		Id:          id,
		Name:        action.Name,
		NamespaceId: action.NamespaceId,
	}).Return(model.Action{}, expectedErr)

	actualAction, actualErr := s.service.UpdateAction(ctx, id, action)

	s.Equal(model.Action{}, actualAction)
	s.ErrorIs(actualErr, expectedErr)
}

func (s *ServiceTestSuite) TestUpdateAction_ShouldReturnUpdatedAction() {
	id := "456"
	action := model.Action{Id: "123", Name: "Read", NamespaceId: "team"}
	ctx := context.Background()

	expectedAction := model.Action{Id: "456", Name: "Read", NamespaceId: "team"}

	s.store.On("UpdateAction", ctx, model.Action{
		Id:          id,
		Name:        action.Name,
		NamespaceId: action.NamespaceId,
	}).Return(expectedAction, nil)

	actualAction, actualErr := s.service.UpdateAction(ctx, id, action)

	s.Equal(expectedAction, actualAction)
	s.Nil(actualErr)
}

func (s *ServiceTestSuite) TestCreateAction_ShouldThrowErrorOnStoreError() {
	action := model.Action{Name: "Read", NamespaceId: "team"}
	ctx := context.Background()
	expectedErr := errors.New("error creating action in store")
	s.store.On("CreateAction", ctx, action).Return(model.Action{}, expectedErr)

	actualAction, actualErr := s.service.CreateAction(ctx, action)

	s.Equal(model.Action{}, actualAction)
	s.ErrorIs(actualErr, expectedErr)
}

func (s *ServiceTestSuite) TestCreateAction_ShouldReturnCreatedAction() {
	action := model.Action{Name: "Read", NamespaceId: "team"}
	ctx := context.Background()

	expectedAction := model.Action{Id: "456", Name: "Read", NamespaceId: "team"}
	s.store.On("CreateAction", ctx, action).Return(expectedAction, nil)

	actualAction, actualErr := s.service.CreateAction(ctx, action)

	s.Equal(expectedAction, actualAction)
	s.Nil(actualErr)
}
