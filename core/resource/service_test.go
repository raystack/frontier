package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/resource/mocks"
	"github.com/raystack/frontier/core/user"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgauditrecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestService(t *testing.T) (*mocks.Repository, *mocks.RelationService, *mocks.AuthnService, *mocks.ProjectService, *mocks.OrgService, *mocks.PATService, *mocks.AuditRecordRepository, *mocks.PolicyService, *resource.Service) {
	t.Helper()
	repo := mocks.NewRepository(t)
	relationSvc := mocks.NewRelationService(t)
	authnSvc := mocks.NewAuthnService(t)
	projectSvc := mocks.NewProjectService(t)
	orgSvc := mocks.NewOrgService(t)
	patSvc := mocks.NewPATService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)
	policySvc := mocks.NewPolicyService(t)
	svc := resource.NewService(repo, relationSvc, authnSvc, projectSvc, orgSvc, patSvc, policySvc, auditRepo)
	return repo, relationSvc, authnSvc, projectSvc, orgSvc, patSvc, auditRepo, policySvc, svc
}

func TestCheckAuthz_NonPAT(t *testing.T) {
	ctx := context.Background()
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, patSvc, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
	_, relationSvc, authnSvc, _, _, _, _, _, svc := newTestService(t)

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
		repo, relationSvc, authnSvc, projectSvc, _, _, auditRepo, _, svc := newTestService(t)

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
		repo, relationSvc, authnSvc, projectSvc, _, patSvc, auditRepo, _, svc := newTestService(t)

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

	t.Run("audit record target name uses title when set", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, _, svc := newTestService(t)
		userID := uuid.New().String()

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		repo.EXPECT().Create(mock.Anything, mock.Anything).Return(resource.Resource{
			ID: uuid.New().String(), Name: "aoi-uuid-1", Title: "My AOI", NamespaceID: "resource/item",
			ProjectID: testProject.ID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.Anything).Return(relation.Relation{}, nil).Times(2)

		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r auditmodels.AuditRecord) bool {
			return r.Target != nil && r.Target.Name == "My AOI"
		})).Return(auditmodels.AuditRecord{}, nil)

		_, err := svc.Create(ctx, resource.Resource{
			Name: "aoi-uuid-1", Title: "My AOI", NamespaceID: "resource/item", ProjectID: testProject.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
	})

	t.Run("audit record target name falls back to name when title empty", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, _, svc := newTestService(t)
		userID := uuid.New().String()

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		repo.EXPECT().Create(mock.Anything, mock.Anything).Return(resource.Resource{
			ID: uuid.New().String(), Name: "res-slug", NamespaceID: "resource/item",
			ProjectID: testProject.ID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.Anything).Return(relation.Relation{}, nil).Times(2)

		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r auditmodels.AuditRecord) bool {
			return r.Target != nil && r.Target.Name == "res-slug"
		})).Return(auditmodels.AuditRecord{}, nil)

		_, err := svc.Create(ctx, resource.Resource{
			Name: "res-slug", NamespaceID: "resource/item", ProjectID: testProject.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
	})

	t.Run("explicit principal skips authn lookup", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, _, svc := newTestService(t)
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

	t.Run("project resource also stores the project's org", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, _, svc := newTestService(t)
		userID := uuid.New().String()

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r resource.Resource) bool {
			return r.ProjectID == testProject.ID && r.OrgID == testProject.Organization.ID
		})).Return(resource.Resource{
			ID: uuid.New().String(), Name: "res-org-id", NamespaceID: "resource/item",
			ProjectID: testProject.ID, OrgID: testProject.Organization.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.Anything).Return(relation.Relation{}, nil).Times(2)

		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		_, err := svc.Create(ctx, resource.Resource{
			Name: "res-org-id", NamespaceID: "resource/item", ProjectID: testProject.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
	})

	t.Run("links the resource to its project and names the project in the audit record", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, _, svc := newTestService(t)
		userID := uuid.New().String()

		projectSvc.EXPECT().Get(mock.Anything, testProject.ID).Return(testProject, nil)

		createdRes := resource.Resource{
			ID: uuid.New().String(), Name: "res-link", NamespaceID: "resource/item",
			ProjectID: testProject.ID, OrgID: testProject.Organization.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}
		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r resource.Resource) bool {
			return r.URN == "frn:"+testProject.Name+":resource/item:res-link"
		})).Return(createdRes, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, relation.Relation{
			Object:       relation.Object{ID: createdRes.ID, Namespace: "resource/item"},
			Subject:      relation.Subject{ID: testProject.ID, Namespace: schema.ProjectNamespace},
			RelationName: schema.ProjectRelationName,
		}).Return(relation.Relation{}, nil).Once()
		relationSvc.EXPECT().Create(mock.Anything, mock.MatchedBy(func(rel relation.Relation) bool {
			return rel.RelationName == schema.OwnerRelationName && rel.Subject.ID == userID
		})).Return(relation.Relation{}, nil).Once()

		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r auditmodels.AuditRecord) bool {
			return r.Resource.ID == testProject.ID && r.Resource.Type == pkgauditrecord.ProjectType &&
				r.OrgID == testProject.Organization.ID && r.OrgName == testProject.Organization.Title
		})).Return(auditmodels.AuditRecord{}, nil)

		_, err := svc.Create(ctx, resource.Resource{
			Name: "res-link", NamespaceID: "resource/item", ProjectID: testProject.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
	})
}

func TestCreateOrgResource(t *testing.T) {
	ctx := context.Background()
	testOrg := organization.Organization{
		ID:    uuid.New().String(),
		Name:  "test-org",
		Title: "Test Org",
	}

	t.Run("links the resource to the org and names the org in the audit record", func(t *testing.T) {
		repo, relationSvc, _, _, orgSvc, _, auditRepo, _, svc := newTestService(t)
		userID := uuid.New().String()

		orgSvc.EXPECT().GetRaw(mock.Anything, testOrg.ID).Return(testOrg, nil)

		createdRes := resource.Resource{
			ID: uuid.New().String(), Name: "store-1", NamespaceID: "resource/store",
			OrgID: testOrg.ID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		}
		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r resource.Resource) bool {
			return r.ProjectID == "" && r.OrgID == testOrg.ID &&
				r.URN == "frn:"+testOrg.ID+":resource/store:store-1"
		})).Return(createdRes, nil)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		relationSvc.EXPECT().Create(mock.Anything, mock.MatchedBy(func(rel relation.Relation) bool {
			return rel.RelationName == schema.OrganizationRelationName &&
				rel.Object.ID == createdRes.ID && rel.Object.Namespace == "resource/store" &&
				rel.Subject.ID == testOrg.ID && rel.Subject.Namespace == schema.OrganizationNamespace
		})).Return(relation.Relation{}, nil).Once()
		relationSvc.EXPECT().Create(mock.Anything, mock.MatchedBy(func(rel relation.Relation) bool {
			return rel.RelationName == schema.OwnerRelationName && rel.Subject.ID == userID
		})).Return(relation.Relation{}, nil).Once()

		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(r auditmodels.AuditRecord) bool {
			return r.Event == pkgauditrecord.ResourceCreatedEvent &&
				r.Resource.ID == testOrg.ID && r.Resource.Type == pkgauditrecord.OrganizationType &&
				r.Resource.Name == testOrg.Title &&
				r.OrgID == testOrg.ID && r.OrgName == testOrg.Title &&
				r.Target != nil && r.Target.ID == createdRes.ID
		})).Return(auditmodels.AuditRecord{}, nil)

		got, err := svc.Create(ctx, resource.Resource{
			Name: "store-1", NamespaceID: "resource/store", OrgID: testOrg.ID,
			PrincipalID: userID, PrincipalType: schema.UserPrincipal,
		})
		assert.NoError(t, err)
		assert.Equal(t, createdRes, got)
	})

	t.Run("returns the org lookup error without creating anything", func(t *testing.T) {
		_, _, _, _, orgSvc, _, _, _, svc := newTestService(t)

		orgSvc.EXPECT().GetRaw(mock.Anything, testOrg.ID).Return(organization.Organization{}, organization.ErrNotExist)

		_, err := svc.Create(ctx, resource.Resource{
			Name: "store-1", NamespaceID: "resource/store", OrgID: testOrg.ID,
			PrincipalID: uuid.New().String(), PrincipalType: schema.UserPrincipal,
		})
		assert.ErrorIs(t, err, organization.ErrNotExist)
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
	ctx := context.Background()
	proj := project.Project{
		ID:           uuid.New().String(),
		Title:        "Project One",
		Organization: organization.Organization{ID: uuid.New().String(), Title: "Org One"},
	}
	res := resource.Resource{ID: "r1", Name: "res-1", Title: "Resource One", NamespaceID: "resource/item", ProjectID: proj.ID}
	lookup := project.Filter{ProjectIDs: []string{proj.ID}, IncludeDisabled: true}
	resPolicies := policy.Filter{ResourceID: "r1", ResourceType: "resource/item"}

	t.Run("deletes relations then the row, and writes the audit record", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, relation.Relation{
			Object: relation.Object{ID: "r1", Namespace: "resource/item"},
		}).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(rec auditmodels.AuditRecord) bool {
			return rec.Event == pkgauditrecord.ResourceDeletedEvent &&
				rec.Target != nil && rec.Target.ID == "r1" && rec.Target.Name == "Resource One" &&
				rec.Resource.ID == proj.ID && rec.OrgID == proj.Organization.ID
		})).Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("deletes an org resource and names the org in the audit record", func(t *testing.T) {
		repo, relationSvc, _, _, orgSvc, _, auditRepo, policySvc, svc := newTestService(t)
		org := organization.Organization{ID: uuid.New().String(), Title: "Org Two"}
		orgRes := resource.Resource{ID: "r2", Name: "store-1", NamespaceID: "resource/store", OrgID: org.ID}

		repo.EXPECT().GetByID(mock.Anything, "r2").Return(orgRes, nil)
		orgSvc.EXPECT().GetRaw(mock.Anything, org.ID).Return(org, nil)
		policySvc.EXPECT().List(mock.Anything, policy.Filter{ResourceID: "r2", ResourceType: "resource/store"}).
			Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, relation.Relation{
			Object: relation.Object{ID: "r2", Namespace: "resource/store"},
		}).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r2").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(rec auditmodels.AuditRecord) bool {
			return rec.Event == pkgauditrecord.ResourceDeletedEvent &&
				rec.Resource.ID == org.ID && rec.Resource.Type == pkgauditrecord.OrganizationType &&
				rec.OrgID == org.ID && rec.OrgName == org.Title
		})).Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/store", "r2")
		assert.NoError(t, err)
	})

	t.Run("returns not found when the resource does not exist", func(t *testing.T) {
		repo, _, _, _, _, _, _, _, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(resource.Resource{}, resource.ErrNotExist)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.ErrorIs(t, err, resource.ErrNotExist)
	})

	t.Run("ignores relation not exist error", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(relation.ErrNotExist)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("returns relation delete error", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, _, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(errors.New("spicedb down"))

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.ErrorContains(t, err, "spicedb down")
	})

	t.Run("writes no audit record when the row delete fails", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, _, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(errors.New("db down"))

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("deletes a resource of a disabled project", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, policySvc, svc := newTestService(t)
		disabled := proj
		disabled.State = project.Disabled

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{disabled}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(rec auditmodels.AuditRecord) bool {
			return rec.Resource.ID == proj.ID
		})).Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("deletes nothing when the project is gone", func(t *testing.T) {
		repo, _, _, projectSvc, _, _, _, _, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{}, nil)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.ErrorIs(t, err, project.ErrNotExist)
	})

	t.Run("deletes the policies on the resource before its relations and row", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, policySvc, svc := newTestService(t)
		var steps []string

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{{ID: "pol-1"}, {ID: "pol-2"}}, nil)
		policySvc.EXPECT().Delete(mock.Anything, mock.Anything).
			Run(func(_ context.Context, id string) { steps = append(steps, "policy:"+id) }).
			Return(nil)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).
			Run(func(_ context.Context, _ relation.Relation) { steps = append(steps, "relations") }).
			Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").
			Run(func(_ context.Context, _ string) { steps = append(steps, "row") }).
			Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.NoError(t, err)
		assert.Equal(t, []string{"policy:pol-1", "policy:pol-2", "relations", "row"}, steps)
	})

	t.Run("skips a policy that is already gone", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{{ID: "pol-1"}}, nil)
		policySvc.EXPECT().Delete(mock.Anything, "pol-1").Return(policy.ErrNotExist)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("stops at the first failed policy delete and keeps the resource", func(t *testing.T) {
		repo, _, _, projectSvc, _, _, _, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{{ID: "pol-1"}, {ID: "pol-2"}}, nil)
		policySvc.EXPECT().Delete(mock.Anything, "pol-1").Return(errors.New("spicedb down"))

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.ErrorContains(t, err, "spicedb down")
		policySvc.AssertNotCalled(t, "Delete", mock.Anything, "pol-2")
	})

	t.Run("keeps the resource when listing its policies fails", func(t *testing.T) {
		repo, _, _, projectSvc, _, _, _, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return(nil, errors.New("db down"))

		err := svc.Delete(ctx, "resource/item", "r1")
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("looks up the policies with the namespace stored on the resource", func(t *testing.T) {
		repo, relationSvc, _, projectSvc, _, _, auditRepo, policySvc, svc := newTestService(t)

		repo.EXPECT().GetByID(mock.Anything, "r1").Return(res, nil)
		projectSvc.EXPECT().List(mock.Anything, lookup).Return([]project.Project{proj}, nil)
		policySvc.EXPECT().List(mock.Anything, resPolicies).Return([]policy.Policy{}, nil)
		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil)
		repo.EXPECT().Delete(mock.Anything, "r1").Return(nil)
		auditRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.AuditRecord")).
			Return(auditmodels.AuditRecord{}, nil)

		err := svc.Delete(ctx, "resource/other", "r1")
		assert.NoError(t, err)
	})
}

func TestPurge(t *testing.T) {
	t.Run("removes relations then the row", func(t *testing.T) {
		repo, relationSvc, _, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Delete(mock.Anything, relation.Relation{
			Object: relation.Object{ID: "r1", Namespace: "resource/item"},
		}).Return(nil)
		repo.EXPECT().Purge(mock.Anything, "r1").Return(nil)

		err := svc.Purge(context.Background(), "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("ignores relation not exist error", func(t *testing.T) {
		repo, relationSvc, _, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(relation.ErrNotExist)
		repo.EXPECT().Purge(mock.Anything, "r1").Return(nil)

		err := svc.Purge(context.Background(), "resource/item", "r1")
		assert.NoError(t, err)
	})

	t.Run("returns relation delete error", func(t *testing.T) {
		_, relationSvc, _, _, _, _, _, _, svc := newTestService(t)

		relationSvc.EXPECT().Delete(mock.Anything, mock.Anything).Return(errors.New("spicedb down"))

		err := svc.Purge(context.Background(), "resource/item", "r1")
		assert.ErrorContains(t, err, "spicedb down")
	})
}

func TestAddResourceOwner(t *testing.T) {
	t.Run("creates owner relation with user principal", func(t *testing.T) {
		_, relationSvc, _, _, _, _, _, _, svc := newTestService(t)

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

func TestRemovePrincipalAccess(t *testing.T) {
	const userID = "user-1"

	t.Run("no org and no projects: no-op", func(t *testing.T) {
		_, _, _, _, _, _, _, _, svc := newTestService(t)

		err := svc.RemovePrincipalAccess(context.Background(), userID, schema.UserPrincipal, "", nil)
		assert.NoError(t, err)
	})

	t.Run("deletes only policies on the org's own resources and on resources owned by the given projects", func(t *testing.T) {
		repo, _, _, _, _, _, _, policySvc, svc := newTestService(t)

		policySvc.EXPECT().List(mock.Anything, policy.Filter{
			PrincipalID:   userID,
			PrincipalType: schema.UserPrincipal,
		}).Return([]policy.Policy{
			{ID: "res-p1", ResourceType: "compute/instance", ResourceID: "res-1"},
			{ID: "res-p2", ResourceType: "compute/instance", ResourceID: "res-2"},
			{ID: "store-p1", ResourceType: "resource/store", ResourceID: "store-1"},
			{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: "org-1"},
		}, nil)

		// the org owns store-1 and proj-1 owns res-1. res-2 lives elsewhere.
		repo.EXPECT().List(mock.Anything, resource.Filter{OrgID: "org-1"}).
			Return([]resource.Resource{{ID: "store-1"}}, nil)
		repo.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{{ID: "res-1"}}, nil)

		// res-2 and the org policy are untouched.
		policySvc.EXPECT().Delete(mock.Anything, "res-p1").Return(nil)
		policySvc.EXPECT().Delete(mock.Anything, "store-p1").Return(nil)

		err := svc.RemovePrincipalAccess(context.Background(), userID, schema.UserPrincipal, "org-1", []string{"proj-1"})
		assert.NoError(t, err)
	})

	t.Run("an org without projects still clears policies on its own resources", func(t *testing.T) {
		repo, _, _, _, _, _, _, policySvc, svc := newTestService(t)

		policySvc.EXPECT().List(mock.Anything, mock.Anything).Return([]policy.Policy{
			{ID: "store-p1", ResourceType: "resource/store", ResourceID: "store-1"},
		}, nil)
		repo.EXPECT().List(mock.Anything, resource.Filter{OrgID: "org-1"}).
			Return([]resource.Resource{{ID: "store-1"}}, nil)
		policySvc.EXPECT().Delete(mock.Anything, "store-p1").Return(nil)

		err := svc.RemovePrincipalAccess(context.Background(), userID, schema.UserPrincipal, "org-1", nil)
		assert.NoError(t, err)
	})

	t.Run("no policies: skips resource lookup", func(t *testing.T) {
		_, _, _, _, _, _, _, policySvc, svc := newTestService(t)

		policySvc.EXPECT().List(mock.Anything, mock.Anything).Return([]policy.Policy{}, nil)

		err := svc.RemovePrincipalAccess(context.Background(), userID, schema.UserPrincipal, "org-1", []string{"proj-1"})
		assert.NoError(t, err)
	})

	t.Run("aggregates delete errors and keeps going", func(t *testing.T) {
		repo, _, _, _, _, _, _, policySvc, svc := newTestService(t)

		policySvc.EXPECT().List(mock.Anything, mock.Anything).Return([]policy.Policy{
			{ID: "res-p1", ResourceType: "compute/instance", ResourceID: "res-1"},
			{ID: "res-p2", ResourceType: "compute/instance", ResourceID: "res-2"},
		}, nil)
		repo.EXPECT().List(mock.Anything, resource.Filter{OrgID: "org-1"}).
			Return([]resource.Resource{}, nil)
		repo.EXPECT().List(mock.Anything, resource.Filter{ProjectID: "proj-1"}).
			Return([]resource.Resource{{ID: "res-1"}, {ID: "res-2"}}, nil)

		policySvc.EXPECT().Delete(mock.Anything, "res-p1").Return(errors.New("delete boom"))
		policySvc.EXPECT().Delete(mock.Anything, "res-p2").Return(nil)

		err := svc.RemovePrincipalAccess(context.Background(), userID, schema.UserPrincipal, "org-1", []string{"proj-1"})
		assert.ErrorContains(t, err, "delete boom")
	})

	t.Run("returns error when policy list fails", func(t *testing.T) {
		_, _, _, _, _, _, _, policySvc, svc := newTestService(t)

		policySvc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, errors.New("list boom"))

		err := svc.RemovePrincipalAccess(context.Background(), userID, schema.UserPrincipal, "org-1", []string{"proj-1"})
		assert.ErrorContains(t, err, "list boom")
	})
}
