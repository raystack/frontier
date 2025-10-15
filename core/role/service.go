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
}

func NewService(repository Repository, relationService RelationService, permissionService PermissionService, auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		repository:            repository,
		relationService:       relationService,
		permissionService:     permissionService,
		auditRecordRepository: auditRecordRepository,
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
		Event: pkgAuditRecord.RoleCreatedEvent.String(),
		Resource: models.Resource{
			ID:   createdRole.OrgID,
			Type: pkgAuditRecord.OrganizationType.String(),
			Name: extractOrgNameFromMetadata(createdRole.OrgID, toCreate.Metadata),
		},
		Target: &models.Target{
			ID:   createdRole.ID,
			Type: pkgAuditRecord.RoleType.String(),
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
				ID:        "*", // all principles who have role will have access
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
				ID:        "*", // all principles who have role will have access
				Namespace: schema.ServiceUserPrincipal,
			},
			RelationName: perm,
		})
		if err != nil {
			return err
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
			ID:        "*", // all principles who have role will have access
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
			ID:        "*", // all principles who have role will have access
			Namespace: schema.ServiceUserPrincipal,
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
		Event: pkgAuditRecord.RoleUpdatedEvent.String(),
		Resource: models.Resource{
			ID:   updatedRole.OrgID,
			Type: pkgAuditRecord.OrganizationType.String(),
			Name: extractOrgNameFromMetadata(updatedRole.OrgID, toUpdate.Metadata),
		},
		Target: &models.Target{
			ID:   updatedRole.ID,
			Type: pkgAuditRecord.RoleType.String(),
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

func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{Object: relation.Object{
		ID:        id,
		Namespace: schema.RoleNamespace,
	}}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
