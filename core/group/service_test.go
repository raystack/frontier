package group_test

import (
	"context"
	"errors"
	"testing"
	"strings"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/group/mocks"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	t.Run("should create group successfully by adding member to org, adding relation between group and org, and making current user owner", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockPolicySvc := mocks.NewPolicyService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)

		mockUserID := uuid.New()
		mockAuthnSvc.On("GetPrincipal", mock.Anything).Return(authenticate.Principal{
			ID:   mockUserID.String(),
			Type: "user",
			User: &user.User{
				ID: mockUserID.String(),
			},
		}, nil)

		groupParam := group.Group{
			Name:           "test-group",
			Title:          "Test Group",
			OrganizationID: uuid.New().String(),
		}

		groupInRepo := groupParam
		groupInRepo.ID = uuid.New().String()
		mockRepo.On("Create", mock.Anything, groupParam).Return(groupInRepo, nil)

		// when adding group as org member
		mockRelationSvc.On("Create", mock.Anything, mock.AnythingOfType("relation.Relation")).Run(func(args mock.Arguments) {
			arg := args.Get(1)
			r := arg.(relation.Relation)
			assert.Equal(t, r.Object.ID, groupInRepo.OrganizationID)
			assert.Equal(t, r.Subject.ID, groupInRepo.ID)
			assert.Equal(t, r.RelationName, schema.MemberRelationName)
		}).Return(relation.Relation{}, nil).Once()

		// when adding group to org
		mockRelationSvc.On("Create", mock.Anything, mock.AnythingOfType("relation.Relation")).Run(func(args mock.Arguments) {
			arg := args.Get(1)
			r := arg.(relation.Relation)
			assert.Equal(t, r.Object.ID, groupInRepo.ID)
			assert.Equal(t, r.Subject.ID, groupInRepo.OrganizationID)
			assert.Equal(t, r.RelationName, schema.OrganizationRelationName)
		}).Return(relation.Relation{}, nil).Once()

		// when adding current user as group owner
		mockPolicySvc.On("Create", mock.Anything, mock.AnythingOfType("policy.Policy")).Run(func(args mock.Arguments) {
			arg := args.Get(1)
			r := arg.(policy.Policy)
			assert.Equal(t, r.RoleID, schema.GroupOwnerRole)
			assert.Equal(t, r.ResourceID, groupInRepo.ID)
			assert.Equal(t, r.ResourceType, schema.GroupNamespace)
			assert.Equal(t, r.PrincipalID, mockUserID.String())
			assert.Equal(t, r.PrincipalType, "user")
		}).Return(policy.Policy{}, nil).Once()

		// adding relation between group and user
		mockRelationSvc.On("Create", mock.Anything, mock.AnythingOfType("relation.Relation")).Run(func(args mock.Arguments) {
			arg := args.Get(1)
			r := arg.(relation.Relation)
			assert.Equal(t, r.Object.ID, groupInRepo.ID)
			assert.Equal(t, r.Object.Namespace, schema.GroupNamespace)
			assert.Equal(t, r.Subject.ID, mockUserID.String())
			assert.Equal(t, r.Subject.Namespace, "user")
			assert.Equal(t, r.RelationName, schema.OwnerRelationName)
		}).Return(relation.Relation{}, nil).Once()

		grp, err := svc.Create(context.Background(), groupParam)

		assert.Nil(t, err)
		assert.Equal(t, grp.Name, groupParam.Name)
	})

	t.Run("should return an error if principal is not found", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockPolicySvc := mocks.NewPolicyService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)

		mockAuthnSvc.On("GetPrincipal", mock.Anything).Return(authenticate.Principal{}, errors.New("internal-error"))

		_, err := svc.Create(context.Background(), group.Group{})
		assert.NotNil(t, err)
		assert.Equal(t, strings.Contains(err.Error(), authenticate.ErrInvalidID.Error()), true)
	})
}
