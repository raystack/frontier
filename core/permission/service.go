package permission

import (
	"context"
	"errors"
	"log/slog"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
)

// RelationService is used to delete the SpiceDB tuples that grant a permission,
// so deleting a permission doesn't leave them behind.
type RelationService interface {
	Delete(ctx context.Context, rel relation.Relation) error
}

// RoleService is implemented by role.Service and injected via SetRoleService
// (not the constructor) because the role and permission services depend on each
// other.
type RoleService interface {
	// RemovePermissionFromRoles removes a deleted permission from each role's list.
	RemovePermissionFromRoles(ctx context.Context, slug string) error
}

type Service struct {
	logger          *slog.Logger
	repository      Repository
	relationService RelationService
	roleService     RoleService
}

func NewService(logger *slog.Logger, repository Repository, relationService RelationService) *Service {
	return &Service{
		logger:          logger,
		repository:      repository,
		relationService: relationService,
	}
}

// SetRoleService wires in the role service used to remove a deleted permission
// from role lists. Set after construction because the permission and role
// services depend on each other.
func (s *Service) SetRoleService(roleService RoleService) {
	s.roleService = roleService
}

func (s Service) Get(ctx context.Context, id string) (Permission, error) {
	if utils.IsValidUUID(id) {
		return s.repository.Get(ctx, id)
	}
	return s.repository.GetBySlug(ctx, ParsePermissionName(id))
}

func (s Service) Upsert(ctx context.Context, perm Permission) (Permission, error) {
	if perm.Slug == "" {
		perm.Slug = perm.GenerateSlug()
	}
	return s.repository.Upsert(ctx, perm)
}

func (s Service) List(ctx context.Context, flt Filter) ([]Permission, error) {
	return s.repository.List(ctx, flt)
}

func (s Service) Update(ctx context.Context, perm Permission) (Permission, error) {
	if perm.Slug == "" {
		perm.Slug = perm.GenerateSlug()
	}
	return s.repository.Update(ctx, perm)
}

// Delete removes a permission and everything that points to it:
//   - the SpiceDB tuples that let roles grant it
//     (app/role:<role>#<slug>@<principal>:*, one per principal type),
//   - the permission from every role's list, and
//   - the permission row itself.
//
// These steps span SpiceDB and two DB writes with no shared transaction, so a
// failure partway leaves a partial state. The order puts the grant-tuple
// removal (the one that actually revokes access) first, and each later step
// logs what was already done if it fails, so the leftover can be cleaned up.
func (s Service) Delete(ctx context.Context, id string) error {
	perm, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	slug := perm.Slug
	if slug == "" {
		slug = perm.GenerateSlug()
	}
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: schema.RoleNamespace,
		},
		RelationName: slug,
	}); err != nil && !errors.Is(err, relation.ErrNotExist) {
		// nothing has been removed yet
		return err
	}

	// from here the grant tuples are gone — a later failure is a partial delete
	if err := s.roleService.RemovePermissionFromRoles(ctx, slug); err != nil {
		s.logger.ErrorContext(ctx, "permission delete partially done: grant tuples removed, but failed to remove the permission from role lists",
			"permission_id", perm.ID, "slug", slug, "error", err)
		return err
	}

	if err := s.repository.Delete(ctx, perm.ID); err != nil {
		s.logger.ErrorContext(ctx, "permission delete partially done: grant tuples and role lists cleaned, but failed to delete the permission row",
			"permission_id", perm.ID, "slug", slug, "error", err)
		return err
	}

	return nil
}
