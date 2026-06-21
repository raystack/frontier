package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_RemovePlatformUser(t *testing.T) {
	t.Run("removes both admin and member relations for a user", func(t *testing.T) {
		userSvc := mocks.NewUserService(t)
		// both platform relations are stripped; each UnSudo is a no-op for a relation
		// the user doesn't hold.
		userSvc.EXPECT().UnSudo(mock.Anything, "u1", schema.AdminRelationName).Return(nil)
		userSvc.EXPECT().UnSudo(mock.Anything, "u1", schema.MemberRelationName).Return(nil)

		h := &ConnectHandler{userService: userSvc}
		resp, err := h.RemovePlatformUser(context.Background(), connect.NewRequest(&frontierv1beta1.RemovePlatformUserRequest{
			UserId: "u1",
		}))
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("removes both admin and member relations for a service user", func(t *testing.T) {
		serviceUserSvc := mocks.NewServiceUserService(t)
		serviceUserSvc.EXPECT().UnSudo(mock.Anything, "s1", schema.AdminRelationName).Return(nil)
		serviceUserSvc.EXPECT().UnSudo(mock.Anything, "s1", schema.MemberRelationName).Return(nil)

		h := &ConnectHandler{serviceUserService: serviceUserSvc}
		resp, err := h.RemovePlatformUser(context.Background(), connect.NewRequest(&frontierv1beta1.RemovePlatformUserRequest{
			ServiceuserId: "s1",
		}))
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("rejects a request with neither id", func(t *testing.T) {
		h := &ConnectHandler{}
		resp, err := h.RemovePlatformUser(context.Background(), connect.NewRequest(&frontierv1beta1.RemovePlatformUserRequest{}))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
