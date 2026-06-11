package role

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/auditrecord/models"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type PermissionService interface {
	Get(ctx context.Context, id string) (permission.Permission, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	repository            Repository
	relationService       RelationService
	permissionService     PermissionService
	auditRecordRepository AuditRecordRepository
	patDeniedPerms        map[string]struct{}
}

func NewService(repository Repository, relationService RelationService, permissionService PermissionService,
	auditRecordRepository AuditRecordRepository, patDeniedPerms map[string]struct{}) *Service {
	return &Service{
		repository:            repository,
		relationService:       relationService,
		permissionService:     permissionService,
		auditRecordRepository: auditRecordRepository,
		patDeniedPerms:        patDeniedPerms,
	}
}

// extractOrgNameFromMetadata extracts organization name from role metadata with platform fallback
func extractOrgNameFromMetadata(orgID string, metadata map[string]any) string {
	orgName := "platform"
	if orgID != schema.PlatformOrgID.String() {
		if name, ok := metadata["org_name"].(string); ok && name != "" {
			orgName = name
		}
	}
	return orgName
}

func (s Service) Upsert(ctx context.Context, toCreate Role) (Role, error) {
	for idx, permName := range toCreate.Permissions {
		// verify if perm exists
		if perm, err := s.permissionService.Get(ctx, permName); err != nil {
			return Role{}, fmt.Errorf("%s: %w", permName, err)
		} else {
			toCreate.Permissions[idx] = perm.GenerateSlug()
		}
	}

	createdRole, err := s.repository.Upsert(ctx, toCreate)
	if err != nil {
		return Role{}, err
	}

	// create relation between role and permissions
	if err := s.createRolePermissionRelation(ctx, createdRole.ID, createdRole.Permissions); err != nil {
		return Role{}, err
	}

	// Create audit record - Actor will be auto-enriched from context by repository
	_, err = s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: pkgAuditRecord.RoleCreatedEvent,
		Resource: models.Resource{
			ID:   createdRole.OrgID,
			Type: pkgAuditRecord.OrganizationType,
			Name: extractOrgNameFromMetadata(createdRole.OrgID, toCreate.Metadata),
		},
		Target: &models.Target{
			ID:   createdRole.ID,
			Type: pkgAuditRecord.RoleType,
			Name: createdRole.Name,
			Metadata: map[string]any{
				"permissions": createdRole.Permissions,
			},
		},
		OrgID:      createdRole.OrgID,
		OccurredAt: createdRole.UpdatedAt,
	})
	if err != nil {
		return Role{}, err
	}

	return createdRole, nil
}

func (s Service) createRolePermissionRelation(ctx context.Context, roleID string, permissions []string) error {
	// create relation between role and permissions
	// for example for each permission:
	// app/role:org_owner#organization_delete@app/user:*
	// app/role:org_owner#organization_update@app/user:*
	// this needs to be created for each type of principles
	for _, perm := range permissions {
		_, err := s.relationService.Create(ctx, relation.Relation{
			Object: relation.Object{
				ID:        roleID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principals who have role will have access
				Namespace: schema.UserPrincipal,
			},
			RelationName: perm,
		})
		if err != nil {
			return err
		}
		// do the same with service user
		_, err = s.relationService.Create(ctx, relation.Relation{
			Object: relation.Object{
				ID:        roleID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principals who have role will have access
				Namespace: schema.ServiceUserPrincipal,
			},
			RelationName: perm,
		})
		if err != nil {
			return err
		}
		// do the same with PAT (skip denied permissions)
		if _, denied := s.patDeniedPerms[perm]; !denied {
			_, err = s.relationService.Create(ctx, relation.Relation{
				Object: relation.Object{
					ID:        roleID,
					Namespace: schema.RoleNamespace,
				},
				Subject: relation.Subject{
					ID:        "*", // all principals who have role will have access
					Namespace: schema.PATPrincipal,
				},
				RelationName: perm,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s Service) deleteRolePermissionRelations(ctx context.Context, roleID string) error {
	// delete relation between role and permissions
	// for example for each permission:
	// app/role:org_owner#organization_delete@app/user:*
	// app/role:org_owner#organization_update@app/user:*
	// this needs to be created for each type of principles
	err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        roleID,
			Namespace: schema.RoleNamespace,
		},
		Subject: relation.Subject{
			ID:        "*", // all principals who have role will have access
			Namespace: schema.UserPrincipal,
		},
	})
	if err != nil {
		return err
	}
	// do the same with service user
	err = s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        roleID,
			Namespace: schema.RoleNamespace,
		},
		Subject: relation.Subject{
			ID:        "*", // all principals who have role will have access
			Namespace: schema.ServiceUserPrincipal,
		},
	})
	if err != nil {
		return err
	}
	// do the same with PAT
	err = s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        roleID,
			Namespace: schema.RoleNamespace,
		},
		Subject: relation.Subject{
			ID:        "*",
			Namespace: schema.PATPrincipal,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s Service) Get(ctx context.Context, id string) (Role, error) {
	if utils.IsValidUUID(id) {
		return s.repository.Get(ctx, id)
	}
	// passing empty orgID will return roles created by system
	return s.repository.GetByName(ctx, "", id)
}

func (s Service) List(ctx context.Context, f Filter) ([]Role, error) {
	return s.repository.List(ctx, f)
}

func (s Service) Update(ctx context.Context, toUpdate Role) (Role, error) {
	for idx, permName := range toUpdate.Permissions {
		// verify if perm exists
		if perm, err := s.permissionService.Get(ctx, permName); err != nil {
			return Role{}, fmt.Errorf("%s: %w", permName, err)
		} else {
			toUpdate.Permissions[idx] = perm.Slug
		}
	}

	// fetch existing role
	existingRole, err := s.Get(ctx, toUpdate.ID)
	if err != nil {
		return Role{}, err
	}

	// delete all existing relation between role and permissions
	if err := s.deleteRolePermissionRelations(ctx, existingRole.ID); err != nil {
		return Role{}, err
	}

	// create relation between role and permissions
	if err := s.createRolePermissionRelation(ctx, existingRole.ID, toUpdate.Permissions); err != nil {
		return Role{}, err
	}

	// update in db
	updatedRole, err := s.repository.Update(ctx, toUpdate)
	if err != nil {
		return Role{}, err
	}

	// Create audit record - Actor will be auto-enriched from context by repository
	_, err = s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: pkgAuditRecord.RoleUpdatedEvent,
		Resource: models.Resource{
			ID:   updatedRole.OrgID,
			Type: pkgAuditRecord.OrganizationType,
			Name: extractOrgNameFromMetadata(updatedRole.OrgID, toUpdate.Metadata),
		},
		Target: &models.Target{
			ID:   updatedRole.ID,
			Type: pkgAuditRecord.RoleType,
			Name: updatedRole.Name,
			Metadata: map[string]any{
				"permissions": updatedRole.Permissions,
			},
		},
		OrgID:      updatedRole.OrgID,
		OccurredAt: updatedRole.UpdatedAt,
	})
	if err != nil {
		return Role{}, err
	}

	return updatedRole, nil
}

// RemovePermissionFromRoles removes a permission from every role's list. Called
// when a permission is deleted so no role keeps a permission that no longer
// exists.
//
// It deliberately does not go through Update. Update deletes and recreates a
// role's entire set of permission tuples in SpiceDB (and writes an audit
// record), so every other permission the role keeps would get its tuples
// rewritten for no reason. All we need here is to drop one name from the list —
// a single small DB update that touches no SpiceDB tuples.
func (s Service) RemovePermissionFromRoles(ctx context.Context, slug string) error {
	return s.repository.RemovePermissionFromRoles(ctx, slug)
}

func (s Service) Delete(ctx context.Context, id string) error {
	// resolve name->ID so the row delete and tuple cleanup target the same role
	roleToDelete, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete the row first. The policies.role_id foreign key rejects this while
	// any policy still references the role, surfaced as ErrRoleInUse — so a role that
	// is still in use is left fully intact (no SpiceDB tuples are touched below).
	// This is intentionally a guard, not a cascade: removing a role would revoke
	// access for everyone granted it, so the caller must drop those policies first.
	if err := s.repository.Delete(ctx, roleToDelete.ID); err != nil {
		return err
	}

	// row is gone → remove the role's permission tuples (app/role:<id>#<perm>@...)
	return s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        roleToDelete.ID,
		Namespace: schema.RoleNamespace,
	}})
}
