package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/authenticate/session/mocks"
	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

func TestService_Create(t *testing.T) {
	t.Run("should create a session when parameters are passed correctly", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		mockRepository.On("Set", mock.Anything, mock.AnythingOfType("*session.Session")).Run(func(args mock.Arguments) {
			arg := args.Get(1)
			r := arg.(*session.Session)
			assert.Equal(t, r.UserID, "1")
		}).Return(nil)

		userID := "1"
		sess, err := svc.Create(context.Background(), userID)

		assert.Nil(t, err)
		assert.Equal(t, sess.UserID, "1")
	})

	t.Run("should return an error when session is not successfully set", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		mockRepository.On("Set", mock.Anything, mock.AnythingOfType("*session.Session")).Run(func(args mock.Arguments) {
			arg := args.Get(1)
			r := arg.(*session.Session)
			assert.Equal(t, r.UserID, "1")
		}).Return(errors.New("internal-error"))

		userID := "1"
		_, err := svc.Create(context.Background(), userID)

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "internal-error")
	})
}

func TestService_Refresh(t *testing.T) {
	t.Run("should refresh a session successfully", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		mockSessionID := uuid.New()
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		mockRepository.On("UpdateValidity", mock.Anything, mockSessionID, 24*time.Hour).Return(nil)

		err := svc.Refresh(context.Background(), mockSessionID)

		assert.Nil(t, err)
	})

	t.Run("should return an error if refresh fails", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		mockSessionID := uuid.New()
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		mockRepository.On("UpdateValidity", mock.Anything, mockSessionID, 24*time.Hour).Return(errors.New("internal-error"))

		err := svc.Refresh(context.Background(), mockSessionID)

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "internal-error")
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("should delete a session successfully", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		mockSessionID := uuid.New()
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		mockRepository.On("Delete", mock.Anything, mockSessionID).Return(nil)

		err := svc.Delete(context.Background(), mockSessionID)

		assert.Nil(t, err)
	})

	t.Run("should return an error if deletion fails", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		mockSessionID := uuid.New()
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		mockRepository.On("Delete", mock.Anything, mockSessionID).Return(errors.New("internal-error"))

		err := svc.Delete(context.Background(), mockSessionID)

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "internal-error")
	})
}

func TestService_ExtractFromContext(t *testing.T) {
	t.Run("should be able to extract session from context if it is present", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		mockSessionID := uuid.New()
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		md := metadata.New(map[string]string{consts.SessionIDGatewayKey: mockSessionID.String(), "key2": "val2"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		mockRepository.On("Get", ctx, mockSessionID).Return(&session.Session{
			ID: mockSessionID,
		}, nil)

		sess, err := svc.ExtractFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, sess.ID, mockSessionID)
	})

	t.Run("should return an error if session is not present in context metadata", func(t *testing.T) {
		mockRepository := mocks.NewRepository(t)
		svc := session.NewService(log.NewLogrus(), mockRepository, 24*time.Hour)

		_, err := svc.ExtractFromContext(context.Background())
		assert.NotNil(t, err)
		assert.Equal(t, err, session.ErrNoSession)
	})
}
