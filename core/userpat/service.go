package userpat

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"slices"
	"time"

	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	pkgUtils "github.com/raystack/frontier/pkg/utils"
	"log/slog"
	"github.com/raystack/salt/rql"
	"golang.org/x/crypto/sha3"
)

type OrganizationService interface {
	GetRaw(ctx context.Context, id string) (organization.Organization, error)
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
}

type PolicyService interface {
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
}

type ProjectService interface {
	ListByUser(ctx context.Context, principal authenticate.Principal, flt project.Filter) ([]project.Project, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	repo                  Repository
	config                Config
	logger                *slog.Logger
	orgService            OrganizationService
	roleService           RoleService
	policyService         PolicyService
	projectService        ProjectService
	auditRecordRepository AuditRecordRepository
	deniedPerms           map[string]struct{}
}

func NewService(logger *slog.Logger, repo Repository, config Config, orgService OrganizationService,
	roleService RoleService, policyService PolicyService, projectService ProjectService, auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		repo:                  repo,
		config:                config,
		logger:                logger,
		orgService:            orgService,
		roleService:           roleService,
		policyService:         policyService,
		projectService:        projectService,
		auditRecordRepository: auditRecordRepository,
		deniedPerms:           config.DeniedPermissionsSet(),
	}
}

type CreateRequest struct {
	UserID    string
	OrgID     string
	Title     string
	Scopes    []patmodels.PATScope
	ExpiresAt time.Time
	Metadata  map[string]any
}

// ValidateExpiry checks that the given expiry time is in the future and within
// the configured maximum PAT lifetime.
func (s *Service) ValidateExpiry(expiresAt time.Time) error {
	if !expiresAt.After(time.Now()) {
		return paterrors.ErrExpiryInPast
	}
	if expiresAt.After(time.Now().Add(s.config.MaxExpiry())) {
		return paterrors.ErrExpiryExceeded
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id string) (patmodels.PAT, error) {
	return s.repo.GetByID(ctx, id)
}

// IsTitleAvailable checks if a PAT title is available for the given user and org.
func (s *Service) IsTitleAvailable(ctx context.Context, userID, orgID, title string) (bool, error) {
	if !s.config.Enabled {
		return false, paterrors.ErrDisabled
	}
	return s.repo.IsTitleAvailable(ctx, userID, orgID, title)
}

// Get retrieves a PAT by ID, verifying it belongs to the given user.
// Returns ErrDisabled if PATs are not enabled, ErrNotFound if the PAT
// does not exist or belongs to a different user.
func (s *Service) Get(ctx context.Context, userID, id string) (patmodels.PAT, error) {
	if !s.config.Enabled {
		return patmodels.PAT{}, paterrors.ErrDisabled
	}
	pat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return patmodels.PAT{}, err
	}
	if pat.UserID != userID {
		return patmodels.PAT{}, paterrors.ErrNotFound
	}
	if err := s.enrichWithScope(ctx, &pat); err != nil {
		return patmodels.PAT{}, fmt.Errorf("enriching PAT scope: %w", err)
	}
	return pat, nil
}

// Delete soft-deletes the PAT first, then removes its SpiceDB policies.
// Soft-delete before policy cleanup prevents concurrent Update from re-creating
// policies for a deleted PAT (TOCTOU mitigation).
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	if !s.config.Enabled {
		return paterrors.ErrDisabled
	}
	pat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if pat.UserID != userID {
		return paterrors.ErrNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("soft deleting PAT: %w", err)
	}

	if err := s.deletePolicies(ctx, id); err != nil {
		return fmt.Errorf("deleting policies: %w", err)
	}

	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATRevokedEvent, pat, time.Now().UTC(), nil); err != nil {
		s.logger.Error("failed to create audit record for PAT revocation", "pat_id", id, "error", err)
	}

	return nil
}

// Regenerate creates a new secret and updates the expiry for an existing PAT.
// Scope (roles + projects) and policies are preserved. Expired PATs can be
// regenerated; if reviving one, checks the active count limit.
func (s *Service) Regenerate(ctx context.Context, userID, id string, newExpiresAt time.Time) (patmodels.PAT, string, error) {
	if !s.config.Enabled {
		return patmodels.PAT{}, "", paterrors.ErrDisabled
	}

	pat, err := s.getOwnedPAT(ctx, userID, id)
	if err != nil {
		return patmodels.PAT{}, "", err
	}

	if err := s.ValidateExpiry(newExpiresAt); err != nil {
		return patmodels.PAT{}, "", err
	}

	// If PAT is currently not active, regenerating revives it — check active count limit.
	if !pat.ExpiresAt.After(time.Now()) {
		count, err := s.repo.CountActive(ctx, pat.UserID, pat.OrgID)
		if err != nil {
			return patmodels.PAT{}, "", fmt.Errorf("counting active PATs: %w", err)
		}
		if count >= s.config.MaxPerUserPerOrg {
			return patmodels.PAT{}, "", paterrors.ErrLimitExceeded
		}
	}

	patValue, secretHash, err := s.generatePAT()
	if err != nil {
		return patmodels.PAT{}, "", err
	}

	oldExpiresAt := pat.ExpiresAt
	regenerated, err := s.repo.Regenerate(ctx, id, secretHash, newExpiresAt)
	if err != nil {
		return patmodels.PAT{}, "", fmt.Errorf("regenerating PAT: %w", err)
	}

	if err := s.enrichWithScope(ctx, &regenerated); err != nil {
		return patmodels.PAT{}, "", fmt.Errorf("enriching PAT scope: %w", err)
	}

	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATRegeneratedEvent, regenerated, *regenerated.RegeneratedAt, map[string]any{
		"expires_at":     regenerated.ExpiresAt,
		"old_expires_at": oldExpiresAt,
	}); err != nil {
		s.logger.Error("failed to create audit record for PAT regeneration", "pat_id", id, "error", err)
	}

	return regenerated, patValue, nil
}

// Update replaces a PAT's title, metadata, and scope (roles + projects).
// Scope changes use revoke-all + recreate pattern with a TOCTOU guard
// against concurrent Delete.
func (s *Service) Update(ctx context.Context, toUpdate patmodels.PAT) (patmodels.PAT, error) {
	if !s.config.Enabled {
		return patmodels.PAT{}, paterrors.ErrDisabled
	}

	existing, err := s.getOwnedPAT(ctx, toUpdate.UserID, toUpdate.ID)
	if err != nil {
		return patmodels.PAT{}, err
	}

	if err := s.validateScopes(ctx, toUpdate.Scopes); err != nil {
		return patmodels.PAT{}, err
	}
	if err := s.validateProjectAccess(ctx, toUpdate.UserID, existing.OrgID, toUpdate.Scopes); err != nil {
		return patmodels.PAT{}, err
	}

	oldTitle, oldScopes, err := s.captureOldScope(ctx, &existing)
	if err != nil {
		return patmodels.PAT{}, err
	}

	updated, err := s.repo.Update(ctx, patmodels.PAT{
		ID:       toUpdate.ID,
		Title:    toUpdate.Title,
		Metadata: toUpdate.Metadata,
	})
	if err != nil {
		return patmodels.PAT{}, fmt.Errorf("updating PAT: %w", err)
	}

	if err := s.replacePolicies(ctx, toUpdate.ID, existing.OrgID, toUpdate.Scopes); err != nil {
		return patmodels.PAT{}, err
	}

	if err := s.enrichWithScope(ctx, &updated); err != nil {
		return patmodels.PAT{}, fmt.Errorf("enriching PAT scope: %w", err)
	}

	s.auditUpdate(ctx, updated, toUpdate, oldTitle, oldScopes)

	return updated, nil
}

// getOwnedPAT retrieves a PAT and verifies it belongs to the given user.
func (s *Service) getOwnedPAT(ctx context.Context, userID, id string) (patmodels.PAT, error) {
	pat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return patmodels.PAT{}, err
	}
	if pat.UserID != userID {
		return patmodels.PAT{}, paterrors.ErrNotFound
	}
	return pat, nil
}

// captureOldScope enriches the PAT with its current scope and returns old title and scopes for audit.
func (s *Service) captureOldScope(ctx context.Context, pat *patmodels.PAT) (string, []patmodels.PATScope, error) {
	oldTitle := pat.Title
	if err := s.enrichWithScope(ctx, pat); err != nil {
		return "", nil, fmt.Errorf("enriching old PAT scope: %w", err)
	}
	return oldTitle, pat.Scopes, nil
}

// replacePolicies deletes existing policies and creates new ones from scopes.
// Re-checks PAT existence after delete to guard against concurrent soft-delete.
func (s *Service) replacePolicies(ctx context.Context, patID, orgID string, scopes []patmodels.PATScope) error {
	if err := s.deletePolicies(ctx, patID); err != nil {
		return fmt.Errorf("deleting old policies: %w", err)
	}

	// TOCTOU guard: ensure PAT wasn't soft-deleted by concurrent Delete.
	if _, err := s.repo.GetByID(ctx, patID); err != nil {
		return fmt.Errorf("PAT deleted concurrently: %w", err)
	}

	if err := s.createPolicies(ctx, patID, orgID, scopes); err != nil {
		return fmt.Errorf("creating new policies: %w", err)
	}
	return nil
}

// auditUpdate creates an audit record for the PAT update. Errors are logged, not returned.
func (s *Service) auditUpdate(ctx context.Context, updated patmodels.PAT, toUpdate patmodels.PAT, oldTitle string, oldScopes []patmodels.PATScope) {
	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATUpdatedEvent, updated, time.Now().UTC(), map[string]any{
		"scopes":     updated.Scopes,
		"old_title":  oldTitle,
		"old_scopes": oldScopes,
	}); err != nil {
		s.logger.Error("failed to create audit record for PAT update", "pat_id", toUpdate.ID, "error", err)
	}
}

// deletePolicies removes all SpiceDB policies associated with a PAT.
// Each policy.Delete call removes SpiceDB relations first, then hard-deletes the Postgres policy row.
func (s *Service) deletePolicies(ctx context.Context, patID string) error {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   patID,
		PrincipalType: schema.PATPrincipal,
	})
	if err != nil {
		return fmt.Errorf("listing policies for PAT %s: %w", patID, err)
	}
	for _, pol := range policies {
		if err := s.policyService.Delete(ctx, pol.ID); err != nil {
			return fmt.Errorf("deleting policy %s: %w", pol.ID, err)
		}
	}
	return nil
}

// Create generates a new PAT and returns it with the plaintext value.
// The plaintext value is only available at creation time.
func (s *Service) Create(ctx context.Context, req CreateRequest) (patmodels.PAT, string, error) {
	if !s.config.Enabled {
		return patmodels.PAT{}, "", paterrors.ErrDisabled
	}

	// NOTE: CountActive + Create is not atomic (TOCTOU race). Two concurrent requests
	// could both read count=49 (assuming max limit is 50), pass the check, and create PATs exceeding the limit.
	// Acceptable for now given low concurrency on this endpoint. If this becomes an issue,
	// use an atomic INSERT ... SELECT with a count subquery in the WHERE clause.
	count, err := s.repo.CountActive(ctx, req.UserID, req.OrgID)
	if err != nil {
		return patmodels.PAT{}, "", fmt.Errorf("counting active PATs: %w", err)
	}
	if count >= s.config.MaxPerUserPerOrg {
		return patmodels.PAT{}, "", paterrors.ErrLimitExceeded
	}

	if err := s.validateScopes(ctx, req.Scopes); err != nil {
		return patmodels.PAT{}, "", err
	}
	if err := s.validateProjectAccess(ctx, req.UserID, req.OrgID, req.Scopes); err != nil {
		return patmodels.PAT{}, "", err
	}

	patValue, secretHash, err := s.generatePAT()
	if err != nil {
		return patmodels.PAT{}, "", err
	}

	pat := patmodels.PAT{
		UserID:     req.UserID,
		OrgID:      req.OrgID,
		Title:      req.Title,
		SecretHash: secretHash,
		Metadata:   req.Metadata,
		ExpiresAt:  req.ExpiresAt,
	}

	created, err := s.repo.Create(ctx, pat)
	if err != nil {
		return patmodels.PAT{}, "", err
	}

	if err := s.createPolicies(ctx, created.ID, req.OrgID, req.Scopes); err != nil {
		return patmodels.PAT{}, "", fmt.Errorf("creating policies: %w", err)
	}

	if err := s.enrichWithScope(ctx, &created); err != nil {
		return patmodels.PAT{}, "", fmt.Errorf("enriching PAT scope: %w", err)
	}

	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATCreatedEvent, created, created.CreatedAt, map[string]any{
		"scopes": created.Scopes,
	}); err != nil {
		s.logger.Error("failed to create audit record for PAT", "pat_id", created.ID, "error", err)
	}

	return created, patValue, nil
}

// createAuditRecord logs a PAT lifecycle event with org context and PAT metadata.
func (s *Service) createAuditRecord(ctx context.Context, event pkgAuditRecord.Event, pat patmodels.PAT, occurredAt time.Time, targetMetadata map[string]any) error {
	orgName := ""
	if org, err := s.orgService.GetRaw(ctx, pat.OrgID); err == nil {
		orgName = org.Title
	}

	metadata := make(map[string]any, len(targetMetadata)+1)
	maps.Copy(metadata, targetMetadata)
	metadata["user_id"] = pat.UserID

	if _, err := s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: event,
		Resource: models.Resource{
			ID:   pat.OrgID,
			Type: pkgAuditRecord.OrganizationType,
			Name: orgName,
		},
		Target: &models.Target{
			ID:       pat.ID,
			Type:     pkgAuditRecord.PATType,
			Name:     pat.Title,
			Metadata: metadata,
		},
		OrgID:      pat.OrgID,
		OccurredAt: occurredAt,
	}); err != nil {
		return fmt.Errorf("creating audit record: %w", err)
	}
	return nil
}

// resolveAndValidateRoles fetches roles by IDs and validates they exist and have no denied permissions.
func (s *Service) resolveAndValidateRoles(ctx context.Context, roleIDs []string) ([]role.Role, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	roles, err := s.roleService.List(ctx, role.Filter{IDs: roleIDs})
	if err != nil {
		return nil, fmt.Errorf("fetching roles: %w", err)
	}
	if len(roles) != len(roleIDs) {
		var missing []string
		for _, id := range roleIDs {
			if !slices.ContainsFunc(roles, func(r role.Role) bool { return r.ID == id }) {
				missing = append(missing, id)
			}
		}
		return nil, fmt.Errorf("role IDs not found: %v: %w", missing, paterrors.ErrRoleNotFound)
	}

	if err := s.validateRolePermissions(roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// validateScopes resolves roles, validates permissions, and checks scope-role compatibility.
func (s *Service) validateScopes(ctx context.Context, scopes []patmodels.PATScope) error {
	// Deduplicate role IDs
	roleIDs := make([]string, 0, len(scopes))
	for _, sc := range scopes {
		roleIDs = append(roleIDs, sc.RoleID)
	}
	roleIDs = pkgUtils.Deduplicate(roleIDs)

	roles, err := s.resolveAndValidateRoles(ctx, roleIDs)
	if err != nil {
		return err
	}

	roleMap := make(map[string]role.Role, len(roles))
	for _, r := range roles {
		roleMap[r.ID] = r
	}

	supportedResourceTypes := []string{schema.OrganizationNamespace, schema.ProjectNamespace}

	for _, sc := range scopes {
		if !slices.Contains(supportedResourceTypes, sc.ResourceType) {
			return fmt.Errorf("resource type %s: %w", sc.ResourceType, paterrors.ErrUnsupportedScope)
		}
		r := roleMap[sc.RoleID]
		if !slices.Contains(r.Scopes, sc.ResourceType) {
			return fmt.Errorf("role %s does not support resource type %s: %w", sc.RoleID, sc.ResourceType, paterrors.ErrScopeMismatch)
		}
	}
	return nil
}

// validateProjectAccess checks that the user has access to all project resource IDs in the scopes.
func (s *Service) validateProjectAccess(ctx context.Context, userID, orgID string, scopes []patmodels.PATScope) error {
	var projectIDs []string
	for _, sc := range scopes {
		if sc.ResourceType == schema.ProjectNamespace {
			projectIDs = append(projectIDs, sc.ResourceIDs...)
		}
	}
	if len(projectIDs) == 0 {
		return nil
	}

	principal := authenticate.Principal{
		ID:   userID,
		Type: schema.UserPrincipal,
	}
	userProjects, err := s.projectService.ListByUser(ctx, principal, project.Filter{OrgID: orgID})
	if err != nil {
		return fmt.Errorf("listing user projects: %w", err)
	}

	userProjectSet := make(map[string]bool, len(userProjects))
	for _, p := range userProjects {
		userProjectSet[p.ID] = true
	}

	var forbidden []string
	for _, id := range projectIDs {
		if !userProjectSet[id] {
			forbidden = append(forbidden, id)
		}
	}
	if len(forbidden) > 0 {
		s.logger.Error("user does not have access to projects", "project_ids", forbidden)
		return paterrors.ErrProjectForbidden
	}
	return nil
}

// createPolicies creates SpiceDB policies from pre-validated scopes.
func (s *Service) createPolicies(ctx context.Context, patID, orgID string, scopes []patmodels.PATScope) error {
	for _, sc := range scopes {
		switch sc.ResourceType {
		case schema.OrganizationNamespace:
			if err := s.createOrgScopedPolicy(ctx, patID, orgID, sc.RoleID); err != nil {
				return err
			}
		case schema.ProjectNamespace:
			if err := s.createProjectScopedPolicies(ctx, patID, orgID, sc.RoleID, sc.ResourceIDs); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported resource type %s: %w", sc.ResourceType, paterrors.ErrUnsupportedScope)
		}
	}
	return nil
}

// ListAllowedRoles returns predefined roles that are valid for PAT assignment.
// It lists platform roles filtered by scopes and removes any role containing
// a denied permission. If scopes is empty, defaults to org + project scopes.
// Accepts short aliases (e.g. "project", "org") which are normalized to full
// namespace form (e.g. "app/project", "app/organization").
func (s *Service) ListAllowedRoles(ctx context.Context, scopes []string) ([]role.Role, error) {
	if !s.config.Enabled {
		return nil, paterrors.ErrDisabled
	}

	if len(scopes) == 0 {
		scopes = []string{schema.OrganizationNamespace, schema.ProjectNamespace}
	} else {
		for i, scope := range scopes {
			scopes[i] = schema.ParseNamespaceAliasIfRequired(scope)
		}
		allowedScopes := []string{schema.OrganizationNamespace, schema.ProjectNamespace}
		scopes = pkgUtils.Deduplicate(scopes)
		for _, scope := range scopes {
			if !slices.Contains(allowedScopes, scope) {
				return nil, fmt.Errorf("scope %q: %w", scope, paterrors.ErrUnsupportedScope)
			}
		}
	}

	roles, err := s.roleService.List(ctx, role.Filter{
		OrgID:  schema.PlatformOrgID.String(),
		Scopes: scopes,
	})
	if err != nil {
		return nil, fmt.Errorf("listing roles: %w", err)
	}

	allowed := make([]role.Role, 0, len(roles))
	for _, r := range roles {
		if !s.hasAnyDeniedPermission(r) {
			allowed = append(allowed, r)
		}
	}
	return allowed, nil
}

// hasAnyDeniedPermission returns true if the role contains at least one denied permission.
func (s *Service) hasAnyDeniedPermission(r role.Role) bool {
	for _, perm := range r.Permissions {
		if _, denied := s.deniedPerms[perm]; denied {
			return true
		}
	}
	return false
}

// validateRolePermissions checks that none of the roles contain denied permissions.
func (s *Service) validateRolePermissions(roles []role.Role) error {
	for _, r := range roles {
		if s.hasAnyDeniedPermission(r) {
			return fmt.Errorf("role %s contains a denied permission: %w", r.Name, paterrors.ErrDeniedRole)
		}
	}
	return nil
}

// createOrgScopedPolicy creates a policy on the org with the default "granted" relation.
func (s *Service) createOrgScopedPolicy(ctx context.Context, patID, orgID, roleID string) error {
	if _, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        roleID,
		ResourceID:    orgID,
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   patID,
		PrincipalType: schema.PATPrincipal,
	}); err != nil {
		return fmt.Errorf("creating org policy for role %s: %w", roleID, err)
	}
	return nil
}

// createProjectScopedPolicies creates policies for a project-scoped role.
// If resourceIDs is empty, it creates a single policy on the org with "pat_granted" relation
// (cascades to all projects). Otherwise, it creates one policy per project with default "granted".
func (s *Service) createProjectScopedPolicies(ctx context.Context, patID, orgID, roleID string, resourceIDs []string) error {
	if len(resourceIDs) == 0 {
		if _, err := s.policyService.Create(ctx, policy.Policy{
			RoleID:        roleID,
			ResourceID:    orgID,
			ResourceType:  schema.OrganizationNamespace,
			PrincipalID:   patID,
			PrincipalType: schema.PATPrincipal,
			GrantRelation: schema.PATGrantRelationName,
		}); err != nil {
			return fmt.Errorf("creating all-projects policy for role %s: %w", roleID, err)
		}
		return nil
	}

	for _, resourceID := range resourceIDs {
		if _, err := s.policyService.Create(ctx, policy.Policy{
			RoleID:        roleID,
			ResourceID:    resourceID,
			ResourceType:  schema.ProjectNamespace,
			PrincipalID:   patID,
			PrincipalType: schema.PATPrincipal,
		}); err != nil {
			return fmt.Errorf("creating project policy for role %s on %s: %w", roleID, resourceID, err)
		}
	}
	return nil
}

// enrichWithScope derives scopes from the PAT's policies.
// Groups policies by role ID + resource type to reconstruct PATScope entries.
func (s *Service) enrichWithScope(ctx context.Context, pat *patmodels.PAT) error {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   pat.ID,
		PrincipalType: schema.PATPrincipal,
	})
	if err != nil {
		return fmt.Errorf("listing policies for PAT %s: %w", pat.ID, err)
	}

	type scopeKey struct {
		roleID       string
		resourceType string
	}
	scopeMap := make(map[scopeKey]*patmodels.PATScope)
	allProjects := make(map[scopeKey]bool)

	for _, pol := range policies {
		var key scopeKey
		var isAllProjects bool

		switch {
		case pol.ResourceType == schema.ProjectNamespace:
			key = scopeKey{pol.RoleID, schema.ProjectNamespace}
		case pol.GrantRelation == schema.PATGrantRelationName:
			key = scopeKey{pol.RoleID, schema.ProjectNamespace}
			isAllProjects = true
		case pol.ResourceType == schema.OrganizationNamespace:
			key = scopeKey{pol.RoleID, schema.OrganizationNamespace}
		default:
			// This should never match — createPolicies and validateScopes
			// only allow app/organization and app/project resource types.
			s.logger.Warn("skipping policy with unsupported resource type during PAT scope enrichment",
				"pat_id", pat.ID, "policy_id", pol.ID, "resource_type", pol.ResourceType)
			continue
		}

		sc, ok := scopeMap[key]
		if !ok {
			sc = &patmodels.PATScope{
				RoleID:       key.roleID,
				ResourceType: key.resourceType,
			}
			scopeMap[key] = sc
		}

		if isAllProjects {
			allProjects[key] = true
			sc.ResourceIDs = nil
		} else if !allProjects[key] {
			sc.ResourceIDs = append(sc.ResourceIDs, pol.ResourceID)
		}
	}

	scopes := make([]patmodels.PATScope, 0, len(scopeMap))
	for _, sc := range scopeMap {
		sc.ResourceIDs = pkgUtils.Deduplicate(sc.ResourceIDs)
		scopes = append(scopes, *sc)
	}
	pat.Scopes = scopes
	return nil
}

// List retrieves all PATs for a user in an org and enriches each with scope fields.
func (s *Service) List(ctx context.Context, userID, orgID string, query *rql.Query) (patmodels.PATList, error) {
	if !s.config.Enabled {
		return patmodels.PATList{}, paterrors.ErrDisabled
	}
	result, err := s.repo.List(ctx, userID, orgID, query)
	if err != nil {
		return patmodels.PATList{}, err
	}
	for i := range result.PATs {
		if err := s.enrichWithScope(ctx, &result.PATs[i]); err != nil {
			return patmodels.PATList{}, fmt.Errorf("enriching PAT scope: %w", err)
		}
	}
	return result, nil
}

// generatePAT creates a random PAT string with the configured prefix and returns
// the plaintext value along with its SHA3-256 hash for storage.
// The hash is computed over the raw secret bytes (not the formatted PAT string)
// to avoid coupling the stored hash to the prefix configuration.
func (s *Service) generatePAT() (patValue, secretHash string, err error) {
	secretBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secretBytes); err != nil {
		return "", "", fmt.Errorf("generating secret: %w", err)
	}

	patValue = s.config.Prefix + "_" + base64.RawURLEncoding.EncodeToString(secretBytes)

	hash := sha3.Sum256(secretBytes)
	secretHash = hex.EncodeToString(hash[:])

	return patValue, secretHash, nil
}
