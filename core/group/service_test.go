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
	"github.com/raystack/frontier/core/user"
	pat "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	t.Run("should create group and delegate hierarchy + owner wiring to membership", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockMembershipSvc := mocks.NewMembershipService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		svc.SetMembershipService(mockMembershipSvc)

		mockUserID := uuid.New().String()
		mockAuthnSvc.On("GetPrincipal", mock.Anything).Return(authenticate.Principal{
			ID:   mockUserID,
			Type: schema.UserPrincipal,
			User: &user.User{ID: mockUserID},
		}, nil)

		groupParam := group.Group{
			Name:           "test-group",
			Title:          "Test Group",
			OrganizationID: uuid.New().String(),
		}
		groupInRepo := groupParam
		groupInRepo.ID = uuid.New().String()
		mockRepo.On("Create", mock.Anything, groupParam).Return(groupInRepo, nil)

		mockMembershipSvc.EXPECT().OnGroupCreated(mock.Anything, groupInRepo.ID, groupInRepo.OrganizationID, mockUserID, schema.UserPrincipal).Return(nil)

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

	t.Run("should propagate error from membership.OnGroupCreated", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockMembershipSvc := mocks.NewMembershipService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		svc.SetMembershipService(mockMembershipSvc)

		mockUserID := uuid.New().String()
		mockAuthnSvc.On("GetPrincipal", mock.Anything).Return(authenticate.Principal{
			ID:   mockUserID,
			Type: schema.UserPrincipal,
			User: &user.User{ID: mockUserID},
		}, nil)

		groupParam := group.Group{Name: "g", OrganizationID: uuid.New().String()}
		groupInRepo := groupParam
		groupInRepo.ID = uuid.New().String()
		mockRepo.On("Create", mock.Anything, groupParam).Return(groupInRepo, nil)
		mockMembershipSvc.EXPECT().OnGroupCreated(mock.Anything, groupInRepo.ID, groupInRepo.OrganizationID, mockUserID, schema.UserPrincipal).Return(errors.New("spicedb down"))

		_, err := svc.Create(context.Background(), groupParam)
		assert.ErrorContains(t, err, "spicedb down")
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

func TestService_List_PrincipalFilter(t *testing.T) {
	ctx := context.Background()

	t.Run("user principal — narrows GroupIDs via membership shim", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockMembershipSvc := mocks.NewMembershipService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		svc.SetMembershipService(mockMembershipSvc)

		principal := authenticate.Principal{ID: "user-123", Type: schema.UserPrincipal}
		mockMembershipSvc.EXPECT().ListGroupsByPrincipal(ctx, principal, "").
			Return([]string{"group-1", "group-2"}, nil).Once()
		mockRepo.On("List", ctx, group.Filter{
			Principal: &principal,
			GroupIDs:  []string{"group-1", "group-2"},
		}).Return([]group.Group{
			{ID: "group-1", Name: "group-one"},
			{ID: "group-2", Name: "group-two"},
		}, nil).Once()

		result, err := svc.List(ctx, group.Filter{Principal: &principal})
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Principal + OrganizationID — forwards orgID to shim and repo", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockMembershipSvc := mocks.NewMembershipService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		svc.SetMembershipService(mockMembershipSvc)

		principal := authenticate.Principal{ID: "user-123", Type: schema.UserPrincipal}
		mockMembershipSvc.EXPECT().ListGroupsByPrincipal(ctx, principal, "org-1").
			Return([]string{"group-1"}, nil).Once()
		mockRepo.On("List", ctx, group.Filter{
			Principal:      &principal,
			OrganizationID: "org-1",
			GroupIDs:       []string{"group-1"},
		}).Return([]group.Group{{ID: "group-1", Name: "group-one"}}, nil).Once()

		result, err := svc.List(ctx, group.Filter{Principal: &principal, OrganizationID: "org-1"})
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("PAT principal — shim handles PAT scoping, service stays oblivious", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockMembershipSvc := mocks.NewMembershipService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		svc.SetMembershipService(mockMembershipSvc)

		principal := authenticate.Principal{
			ID:   "pat-456",
			Type: schema.PATPrincipal,
			PAT:  &pat.PAT{ID: "pat-456", UserID: "user-123", OrgID: "org-1"},
		}
		mockMembershipSvc.EXPECT().ListGroupsByPrincipal(ctx, principal, "").
			Return([]string{"group-1", "group-3"}, nil).Once()
		mockRepo.On("List", ctx, group.Filter{
			Principal: &principal,
			GroupIDs:  []string{"group-1", "group-3"},
		}).Return([]group.Group{
			{ID: "group-1", Name: "group-one"},
			{ID: "group-3", Name: "group-three"},
		}, nil).Once()

		result, err := svc.List(ctx, group.Filter{Principal: &principal})
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("empty membership result — short-circuits to empty slice", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockMembershipSvc := mocks.NewMembershipService(t)

		svc := group.NewService(mockRepo, mockRelationSvc, mockAuthnSvc, mockPolicySvc)
		svc.SetMembershipService(mockMembershipSvc)

		principal := authenticate.Principal{ID: "user-123", Type: schema.UserPrincipal}
		mockMembershipSvc.EXPECT().ListGroupsByPrincipal(ctx, principal, "").
			Return(nil, nil).Once()

		result, err := svc.List(ctx, group.Filter{Principal: &principal})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}
