package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/resource/mocks"
	"github.com/raystack/frontier/core/user"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestService(t *testing.T) (*mocks.Repository, *mocks.ConfigRepository, *mocks.RelationService, *mocks.AuthnService, *mocks.ProjectService, *mocks.OrgService, *mocks.PATService, *mocks.AuditRecordRepository, *resource.Service) {
	t.Helper()
	repo := mocks.NewRepository(t)
	configRepo := mocks.NewConfigRepository(t)
	relationSvc := mocks.NewRelationService(t)
	authnSvc := mocks.NewAuthnService(t)
	projectSvc := mocks.NewProjectService(t)
	orgSvc := mocks.NewOrgService(t)
	patSvc := mocks.NewPATService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)
	svc := resource.NewService(repo, configRepo, relationSvc, authnSvc, projectSvc, orgSvc, patSvc, auditRepo)
	return repo, configRepo, relationSvc, authnSvc, projectSvc, orgSvc, patSvc, auditRepo, svc
}

func TestCheckAuthz_NonPAT(t *testing.T) {
	ctx := context.Background()
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, patSvc, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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
	_, _, relationSvc, authnSvc, _, _, _, _, svc := newTestService(t)

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

func TestGet(t *testing.T) {
	t.Run("get by UUID calls GetByID", func(t *testing.T) {
		repo, _, _, _, _, _, _, _, svc := newTestService(t)
		id := uuid.New().String()
		expected := resource.Resource{ID: id, Name: "test"}
		repo.EXPECT().GetByID(mock.Anything, id).Return(expected, nil)

		got, err := svc.Get(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, got.ID)
	})

	t.Run("get by URN calls GetByURN", func(t *testing.T) {
		repo, _, _, _, _, _, _, _, svc := newTestService(t)
		urn := "frn:myproject:resource/item:myresource"
		expected := resource.Resource{URN: urn, Name: "myresource"}
		repo.EXPECT().GetByURN(mock.Anything, urn).Return(expected, nil)

		got, err := svc.Get(context.Background(), urn)
		assert.NoError(t, err)
		assert.Equal(t, expected.URN, got.URN)
	})
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	testProject := project.Project{
		ID:   uuid.New().String(),
		Name: "test-project",
		Organization: organization.Organization{
			ID:    uuid.New().String(),
			Title: "Test Org",
		},
	}

	t.Run("creates resource with user principal", func(t *testing.T) {
		repo, _, relationSvc, authnSvc, projectSvc, _, _, auditRepo, svc := newTestService(t)

		userID := uuid.New().String()
		authnSvc.EXPECT().GetPrincipal(mock.Anything, mock.Anything).Return(authenticate.Principal{
			ID:   userID,
			Type: schema.UserPrincipal,
			User: &user.User{ID: userID},
		}, nil)

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		createdRes := resource.Resource{
			ID: uuid.New().String(), Name: "res-1", NamespaceID: "resource/item",
			ProjectID: testProject.ID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}
		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r resource.Resource) bool {
			return r.PrincipalID == userID && r.PrincipalType == schema.UserPrincipal
		})).Return(createdRes, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.Anything).Return(relation.Relation{}, nil).Times(2) // project + owner

		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		got, err := svc.Create(ctx, resource.Resource{
			Name: "res-1", NamespaceID: "resource/item", ProjectID: testProject.ID,
		})
		assert.NoError(t, err)
		assert.Equal(t, createdRes.ID, got.ID)
	})

	t.Run("PAT principal resolves to user", func(t *testing.T) {
		repo, _, relationSvc, authnSvc, projectSvc, _, patSvc, auditRepo, svc := newTestService(t)

		patID := uuid.New().String()
		userID := uuid.New().String()
		authnSvc.EXPECT().GetPrincipal(mock.Anything, mock.Anything).Return(authenticate.Principal{
			ID:   patID,
			Type: schema.PATPrincipal,
			PAT:  &patmodels.PAT{ID: patID, UserID: userID},
			User: &user.User{ID: userID},
		}, nil)

		patSvc.EXPECT().GetByID(mock.Anything, patID).Return(patmodels.PAT{
			ID: patID, UserID: userID,
		}, nil).Maybe()

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r resource.Resource) bool {
			return r.PrincipalID == userID && r.PrincipalType == schema.UserPrincipal
		})).Return(resource.Resource{
			ID: uuid.New().String(), Name: "res-pat", NamespaceID: "resource/item",
			ProjectID: testProject.ID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.Anything).Return(relation.Relation{}, nil).Times(2)

		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		got, err := svc.Create(ctx, resource.Resource{
			Name: "res-pat", NamespaceID: "resource/item", ProjectID: testProject.ID,
		})
		assert.NoError(t, err)
		assert.Equal(t, userID, got.PrincipalID)
		assert.Equal(t, schema.UserPrincipal, got.PrincipalType)
	})

	t.Run("explicit principal skips authn lookup", func(t *testing.T) {
		repo, _, relationSvc, _, projectSvc, _, _, auditRepo, svc := newTestService(t)
		userID := uuid.New().String()

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r resource.Resource) bool {
			return r.PrincipalID == userID
		})).Return(resource.Resource{
			ID: uuid.New().String(), Name: "res-explicit", NamespaceID: "resource/item",
			ProjectID: testProject.ID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.Anything).Return(relation.Relation{}, nil).Times(2)

		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		_, err := svc.Create(ctx, resource.Resource{
			Name: "res-explicit", NamespaceID: "resource/item", ProjectID: testProject.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
	})
}

func TestList(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo, _, _, _, _, _, _, _, svc := newTestService(t)
		flt := resource.Filter{ProjectID: "proj-1"}
		expected := []resource.Resource{{ID: "r1"}, {ID: "r2"}}
		repo.EXPECT().List(mock.Anything, flt).Return(expected, nil)

		got, err := svc.List(context.Background(), flt)
		assert.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("returns error from repository", func(t *testing.T) {
		repo, _, _, _, _, _, _, _, svc := newTestService(t)
		repo.EXPECT().List(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

		_, err := svc.List(context.Background(), resource.Filter{})
		assert.ErrorContains(t, err, "db error")
	})
}

func TestUpdate(t *testing.T) {
	t.Run("delegates to repository", func(t *testing.T) {
		repo, _, _, _, _, _, _, _, svc := newTestService(t)
		res := resource.Resource{ID: "r1", Title: "updated"}
		repo.EXPECT().Update(mock.Anything, res).Return(res, nil)

		got, err := svc.Update(context.Background(), res)
		assert.NoError(t, err)
		assert.Equal(t, "updated", got.Title)
	})
}

func TestDelete(t *testing.T) {
	t.Run("deletes relations then resource", func(t *testing.T) {
		repo, _, relationSvc, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Delete(mock.Anything, relation.Relation{
			Object: relation.Object{ID: "r1", Namespace: "resource/item"},
		}).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)

		err := svc.Delete(context.Background(), "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("ignores relation not exist error", func(t *testing.T) {
		repo, _, relationSvc, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(relation.ErrNotExist)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)

		err := svc.Delete(context.Background(), "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("returns relation delete error", func(t *testing.T) {
		_, _, relationSvc, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(errors.New("spicedb down"))

		err := svc.Delete(context.Background(), "resource/item", "r1")
		assert.ErrorContains(t, err, "spicedb down")
	})
}

func TestAddProjectToResource(t *testing.T) {
	t.Run("creates project relation", func(t *testing.T) {
		_, _, relationSvc, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Create(mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "r1", Namespace: "resource/item"},
			Subject:      relation.Subject{ID: "proj-1", Namespace: schema.ProjectNamespace},
			RelationName: schema.ProjectRelationName,
		}).Return(relation.Relation{}, nil)

		err := svc.AddProjectToResource(context.Background(), "proj-1", resource.Resource{
			ID: "r1", NamespaceID: "resource/item",
		})
		assert.NoError(t, err)
	})
}

func TestAddResourceOwner(t *testing.T) {
	t.Run("creates owner relation with user principal", func(t *testing.T) {
		_, _, relationSvc, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Create(mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "r1", Namespace: "resource/item"},
			Subject:      relation.Subject{ID: "user-1", Namespace: schema.UserPrincipal},
			RelationName: schema.OwnerRelationName,
		}).Return(relation.Relation{}, nil)

		err := svc.AddResourceOwner(context.Background(), resource.Resource{
			ID: "r1", NamespaceID: "resource/item",
			PrincipalID: "user-1", PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
	})
}
