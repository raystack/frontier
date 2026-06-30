package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
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

	t.Run("removes only the specified relation when relation is set", func(t *testing.T) {
		userSvc := mocks.NewUserService(t)
		// only the admin relation is stripped; an UnSudo for member would be an
		// unexpected call and fail the mock.
		userSvc.EXPECT().UnSudo(mock.Anything, "u1", schema.AdminRelationName).Return(nil)

		h := &ConnectHandler{userService: userSvc}
		resp, err := h.RemovePlatformUser(context.Background(), connect.NewRequest(&frontierv1beta1.RemovePlatformUserRequest{
			UserId:   "u1",
			Relation: schema.AdminRelationName,
		}))
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("rejects an invalid relation", func(t *testing.T) {
		h := &ConnectHandler{}
		resp, err := h.RemovePlatformUser(context.Background(), connect.NewRequest(&frontierv1beta1.RemovePlatformUserRequest{
			UserId:   "u1",
			Relation: "owner",
		}))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("rejects a request with neither id", func(t *testing.T) {
		h := &ConnectHandler{}
		resp, err := h.RemovePlatformUser(context.Background(), connect.NewRequest(&frontierv1beta1.RemovePlatformUserRequest{}))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_ListPlatformUsers(t *testing.T) {
	t.Run("reports every relation a principal holds, deduped", func(t *testing.T) {
		relationSvc := mocks.NewRelationService(t)
		userSvc := mocks.NewUserService(t)

		// u1 holds both admin and member on the platform.
		relationSvc.EXPECT().List(mock.Anything, mock.Anything).Return([]relation.Relation{
			{Subject: relation.Subject{ID: "u1", Namespace: schema.UserPrincipal}, RelationName: schema.AdminRelationName},
			{Subject: relation.Subject{ID: "u1", Namespace: schema.UserPrincipal}, RelationName: schema.MemberRelationName},
		}, nil)
		// a single, deduped id is fetched even though two tuples were listed.
		userSvc.EXPECT().GetByIDs(mock.Anything, []string{"u1"}).Return([]user.User{{ID: "u1"}}, nil)

		h := &ConnectHandler{relationService: relationSvc, userService: userSvc}
		resp, err := h.ListPlatformUsers(context.Background(), connect.NewRequest(&frontierv1beta1.ListPlatformUsersRequest{}))
		assert.NoError(t, err)
		if assert.Len(t, resp.Msg.GetUsers(), 1) {
			fields := resp.Msg.GetUsers()[0].GetMetadata().GetFields()
			got := []string{}
			for _, v := range fields["relations"].GetListValue().GetValues() {
				got = append(got, v.GetStringValue())
			}
			assert.ElementsMatch(t, []string{schema.AdminRelationName, schema.MemberRelationName}, got)
			// "relation" stays populated for backward compatibility.
			assert.NotEmpty(t, fields["relation"].GetStringValue())
		}
	})
}
