package resource_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/resource/mocks"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestService(t *testing.T) (*mocks.Repository, *mocks.ConfigRepository, *mocks.RelationService, *mocks.AuthnService, *mocks.ProjectService, *mocks.OrgService, *mocks.PATService, *resource.Service) {
	t.Helper()
	repo := mocks.NewRepository(t)
	configRepo := mocks.NewConfigRepository(t)
	relationSvc := mocks.NewRelationService(t)
	authnSvc := mocks.NewAuthnService(t)
	projectSvc := mocks.NewProjectService(t)
	orgSvc := mocks.NewOrgService(t)
	patSvc := mocks.NewPATService(t)
	svc := resource.NewService(repo, configRepo, relationSvc, authnSvc, projectSvc, orgSvc, patSvc)
	return repo, configRepo, relationSvc, authnSvc, projectSvc, orgSvc, patSvc, svc
}

func TestCheckAuthz_NonPAT(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	userID := uuid.New().String()
	orgID := uuid.New().String()

	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   userID,
		Type: schema.UserPrincipal,
	}, nil).Maybe()

	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.GetPermission,
	}).Return(true, nil)

	result, err := svc.CheckAuthz(ctx, resource.Check{
		Object:     relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Permission: schema.GetPermission,
	})
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestCheckAuthz_PATScopeAllowed(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	patID := uuid.New().String()
	userID := uuid.New().String()
	orgID := uuid.New().String()

	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   patID,
		Type: schema.PATPrincipal,
		PAT:  &patmodels.PAT{ID: patID, UserID: userID},
	}, nil).Maybe()

	// PAT scope check — allowed
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.GetPermission,
	}).Return(true, nil)

	// User permission check — allowed
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.GetPermission,
	}).Return(true, nil)

	result, err := svc.CheckAuthz(ctx, resource.Check{
		Object:     relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Permission: schema.GetPermission,
	})
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestCheckAuthz_PATScopeDenied(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	patID := uuid.New().String()
	userID := uuid.New().String()
	orgID := uuid.New().String()

	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   patID,
		Type: schema.PATPrincipal,
		PAT:  &patmodels.PAT{ID: patID, UserID: userID},
	}, nil).Maybe()

	// PAT scope check — denied
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.UpdatePermission,
	}).Return(false, nil)

	// User check should NOT be called (early exit)
	result, err := svc.CheckAuthz(ctx, resource.Check{
		Object:     relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Permission: schema.UpdatePermission,
	})
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestCheckAuthz_PATScopeAllowed_UserDenied(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	patID := uuid.New().String()
	userID := uuid.New().String()
	orgID := uuid.New().String()

	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   patID,
		Type: schema.PATPrincipal,
		PAT:  &patmodels.PAT{ID: patID, UserID: userID},
	}, nil).Maybe()

	// PAT scope check — allowed
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.DeletePermission,
	}).Return(true, nil)

	// User permission check — denied
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.DeletePermission,
	}).Return(false, nil)

	result, err := svc.CheckAuthz(ctx, resource.Check{
		Object:     relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Permission: schema.DeletePermission,
	})
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestCheckAuthz_ExplicitPATSubject_ScopeAllowed(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, patSvc, svc := newTestService(t)

	patID := uuid.New().String()
	userID := uuid.New().String()
	orgID := uuid.New().String()

	// Principal is NOT a PAT (e.g., superuser making federated check with PAT subject)
	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   uuid.New().String(),
		Type: schema.UserPrincipal,
	}, nil).Maybe()

	// PAT scope check for explicit subject — allowed
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.GetPermission,
	}).Return(true, nil)

	// Federated check passes explicit app/pat subject — needs DB lookup
	patSvc.EXPECT().GetByID(ctx, patID).Return(patmodels.PAT{
		ID:     patID,
		UserID: userID,
	}, nil)

	// User permission check (resolved from PAT)
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.GetPermission,
	}).Return(true, nil)

	result, err := svc.CheckAuthz(ctx, resource.Check{
		Object:     relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Subject:    relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Permission: schema.GetPermission,
	})
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestCheckAuthz_ExplicitPATSubject_ScopeDenied(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	patID := uuid.New().String()
	orgID := uuid.New().String()

	// Principal is NOT a PAT (e.g., superuser making federated check with PAT subject)
	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   uuid.New().String(),
		Type: schema.UserPrincipal,
	}, nil).Maybe()

	// PAT scope check for explicit subject — denied
	relationSvc.EXPECT().CheckPermission(ctx, relation.Relation{
		Subject:      relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.UpdatePermission,
	}).Return(false, nil)

	// User check should NOT be called — PAT scope denied
	result, err := svc.CheckAuthz(ctx, resource.Check{
		Object:     relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Subject:    relation.Subject{ID: patID, Namespace: schema.PATPrincipal},
		Permission: schema.UpdatePermission,
	})
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestBatchCheck_PATScopeAllowed(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	patID := uuid.New().String()
	userID := uuid.New().String()
	orgID := uuid.New().String()
	projID := uuid.New().String()

	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   patID,
		Type: schema.PATPrincipal,
		PAT:  &patmodels.PAT{ID: patID, UserID: userID},
	}, nil).Maybe()

	checks := []resource.Check{
		{Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace}, Permission: schema.GetPermission},
		{Object: relation.Object{ID: projID, Namespace: schema.ProjectNamespace}, Permission: schema.GetPermission},
	}

	// PAT scope batch — all allowed
	relationSvc.EXPECT().BatchCheckPermission(ctx, []relation.Relation{
		{Subject: relation.Subject{ID: patID, Namespace: schema.PATPrincipal}, Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace}, RelationName: schema.GetPermission},
		{Subject: relation.Subject{ID: patID, Namespace: schema.PATPrincipal}, Object: relation.Object{ID: projID, Namespace: schema.ProjectNamespace}, RelationName: schema.GetPermission},
	}).Return([]relation.CheckPair{
		{Status: true},
		{Status: true},
	}, nil)

	// User check batch
	relationSvc.EXPECT().BatchCheckPermission(ctx, []relation.Relation{
		{Subject: relation.Subject{ID: userID, Namespace: schema.UserPrincipal}, Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace}, RelationName: schema.GetPermission},
		{Subject: relation.Subject{ID: userID, Namespace: schema.UserPrincipal}, Object: relation.Object{ID: projID, Namespace: schema.ProjectNamespace}, RelationName: schema.GetPermission},
	}).Return([]relation.CheckPair{
		{Status: true},
		{Status: true},
	}, nil)

	results, err := svc.BatchCheck(ctx, checks)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Status)
	assert.True(t, results[1].Status)
}

func TestBatchCheck_PATScopeDenied(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, svc := newTestService(t)

	patID := uuid.New().String()
	userID := uuid.New().String()
	orgID := uuid.New().String()

	authnSvc.EXPECT().GetPrincipal(ctx, mock.Anything).Return(authenticate.Principal{
		ID:   patID,
		Type: schema.PATPrincipal,
		PAT:  &patmodels.PAT{ID: patID, UserID: userID},
	}, nil).Maybe()

	checks := []resource.Check{
		{Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace}, Permission: schema.UpdatePermission},
	}

	// PAT scope batch — denied
	relationSvc.EXPECT().BatchCheckPermission(ctx, []relation.Relation{
		{Subject: relation.Subject{ID: patID, Namespace: schema.PATPrincipal}, Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace}, RelationName: schema.UpdatePermission},
	}).Return([]relation.CheckPair{
		{Status: false},
	}, nil)

	// User check should NOT be called — scope-denied items return false directly
	results, err := svc.BatchCheck(ctx, checks)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
}
