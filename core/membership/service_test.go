package membership_test

import (
	"context"
	"errors"
	"testing"

	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/membership/mocks"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
)

func TestService_RemoveAllPATPolicies(t *testing.T) {
	ctx := context.Background()
	patID := uuid.New().String()

	t.Run("should delete every policy held by the PAT", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)

		mockPolicySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: patID, PrincipalType: schema.PATPrincipal}).
			Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}, {ID: "p3"}}, nil)
		mockPolicySvc.EXPECT().Delete(ctx, "p1").Return(nil)
		mockPolicySvc.EXPECT().Delete(ctx, "p2").Return(nil)
		mockPolicySvc.EXPECT().Delete(ctx, "p3").Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		err := svc.RemoveAllPATPolicies(ctx, patID)
		assert.NoError(t, err)
	})

	t.Run("should be a no-op when the PAT has no policies", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)

		mockPolicySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: patID, PrincipalType: schema.PATPrincipal}).
			Return([]policy.Policy{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		err := svc.RemoveAllPATPolicies(ctx, patID)
		assert.NoError(t, err)
	})

	t.Run("should surface policy list errors", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)

		mockPolicySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: patID, PrincipalType: schema.PATPrincipal}).
			Return(nil, errors.New("db down"))

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		err := svc.RemoveAllPATPolicies(ctx, patID)
		assert.Error(t, err)
	})

	t.Run("should surface policy delete errors", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)

		mockPolicySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: patID, PrincipalType: schema.PATPrincipal}).
			Return([]policy.Policy{{ID: "p1"}}, nil)
		mockPolicySvc.EXPECT().Delete(ctx, "p1").Return(errors.New("spicedb unavailable"))

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		err := svc.RemoveAllPATPolicies(ctx, patID)
		assert.Error(t, err)
	})
}
