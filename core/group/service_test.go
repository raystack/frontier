package group_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

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

func TestService_Get(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPolicySvc := mocks.NewPolicyService(t)

	t.Run("should return group if present", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		paramID := uuid.New().String()

		expectedGroup := group.Group{
			ID:    paramID,
			Name:  "test-group",
			Title: "Test Group",
		}
		mockRepo.On("GetByID", mock.Anything, paramID).Return(expectedGroup, nil).Once()
		actual, err := svc.Get(context.Background(), paramID)

		assert.Nil(t, err)
		assert.Equal(t, expectedGroup, actual)
	})

	t.Run("should return error if group is not present", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		paramID := uuid.New().String()

		mockRepo.On("GetByID", mock.Anything, paramID).Return(group.Group{}, group.ErrNotExist).Once()
		_, err := svc.Get(context.Background(), paramID)
		assert.NotNil(t, err)
		assert.Equal(t, err, group.ErrNotExist)
	})
}

func TestService_GetByIDs(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPolicySvc := mocks.NewPolicyService(t)

	t.Run("should return group if present", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		ID1 := uuid.New().String()
		ID2 := uuid.New().String()

		expectedGroup1 := group.Group{
			ID:    ID1,
			Name:  "test-group-2",
			Title: "Test Group One",
		}
		expectedGroup2 := group.Group{
			ID:    ID2,
			Name:  "test-group-2",
			Title: "Test Group Two",
		}

		expectedGroups := []group.Group{expectedGroup1, expectedGroup2}

		mockRepo.On("GetByIDs", mock.Anything, []string{ID1, ID2}, group.Filter{}).Return(expectedGroups, nil).Once()
		actualGroups, err := svc.GetByIDs(context.Background(), []string{ID1, ID2})

		assert.Nil(t, err)
		assert.Equal(t, len(actualGroups), 2)
		assert.ElementsMatch(t, expectedGroups, actualGroups)
	})

	t.Run("should return error if no groups are found", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		ID1 := uuid.New().String()
		ID2 := uuid.New().String()

		mockRepo.On("GetByIDs", mock.Anything, []string{ID1, ID2}, group.Filter{}).Return([]group.Group{}, group.ErrNotExist).Once()
		actualGroups, err := svc.GetByIDs(context.Background(), []string{ID1, ID2})

		assert.Equal(t, len(actualGroups), 0)
		assert.NotNil(t, err)
		assert.Equal(t, err, group.ErrNotExist)
	})
}

func TestService_List(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPolicySvc := mocks.NewPolicyService(t)

	t.Run("should return list of users based on filters passed", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)

		flt := group.Filter{
			OrganizationID:  "123123123",
			WithMemberCount: true,
		}
		g1 := group.Group{
			ID:   "123",
			Name: "group-1",
		}
		g2 := group.Group{
			ID:   "456",
			Name: "group-2",
		}
		mockRepo.On("List", mock.Anything, flt).Return([]group.Group{g1, g2}, nil)
		mockPolicySvc.On("GroupMemberCount", mock.Anything, []string{"123", "456"}).Return([]policy.MemberCount{{ID: "123", Count: 4}, {ID: "456", Count: 10}}, nil)

		receivedGroups, err := svc.List(context.Background(), flt)
		fmt.Println(receivedGroups)

		expectedGroup1 := g1
		expectedGroup1.MemberCount = 4

		expectedGroup2 := g2
		expectedGroup2.MemberCount = 10
		assert.Nil(t, err)
		assert.ElementsMatch(t, receivedGroups, []group.Group{expectedGroup1, expectedGroup2})
	})

	t.Run("should return an error if no org id or groupID filter is passed", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		flt := group.Filter{}

		_, err := svc.List(context.Background(), flt)
		assert.NotNil(t, err)
		assert.Equal(t, err, group.ErrInvalidID)
	})
}

func TestService_Update(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPolicySvc := mocks.NewPolicyService(t)

	t.Run("should update the group parameters as requested", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)

		groupToBeUpdated := group.Group{
			ID:    "123123",
			Name:  "test-group",
			Title: "Test Group",
		}
		mockRepo.On("UpdateByID", mock.Anything, groupToBeUpdated).Return(groupToBeUpdated, nil)
		grp, err := svc.Update(context.Background(), groupToBeUpdated)

		assert.Nil(t, err)
		assert.Equal(t, grp, groupToBeUpdated)
	})

	t.Run("should return an error if group id is empty", func(t *testing.T) {
		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		_, err := svc.Update(context.Background(), group.Group{ID: ""})
		assert.NotNil(t, err)
		assert.Equal(t, err, group.ErrInvalidID)
	})
}
