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
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	pkgUtils "github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/log"
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

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	repo                  Repository
	config                Config
	logger                log.Logger
	orgService            OrganizationService
	roleService           RoleService
	policyService         PolicyService
	auditRecordRepository AuditRecordRepository
	deniedPerms           map[string]struct{}
}

func NewService(logger log.Logger, repo Repository, config Config, orgService OrganizationService,
	roleService RoleService, policyService PolicyService, auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		repo:                  repo,
		config:                config,
		logger:                logger,
		orgService:            orgService,
		roleService:           roleService,
		policyService:         policyService,
		auditRecordRepository: auditRecordRepository,
		deniedPerms:           config.DeniedPermissionsSet(),
	}
}

type CreateRequest struct {
	UserID     string
	OrgID      string
	Title      string
	RoleIDs    []string
	ProjectIDs []string
	ExpiresAt  time.Time
	Metadata   map[string]any
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

	roles, err := s.resolveAndValidateRoles(ctx, toUpdate.RoleIDs)
	if err != nil {
		return patmodels.PAT{}, err
	}

	oldTitle, oldRoleIDs, oldProjectIDs, err := s.captureOldScope(ctx, &existing)
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

	if err := s.replacePolicies(ctx, toUpdate.ID, existing.OrgID, roles, toUpdate.ProjectIDs); err != nil {
		return patmodels.PAT{}, err
	}

	if err := s.enrichWithScope(ctx, &updated); err != nil {
		return patmodels.PAT{}, fmt.Errorf("enriching PAT scope: %w", err)
	}

	s.auditUpdate(ctx, updated, toUpdate, oldTitle, oldRoleIDs, oldProjectIDs)

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

// captureOldScope enriches the PAT with its current scope and returns old title, role IDs, and project IDs for audit.
func (s *Service) captureOldScope(ctx context.Context, pat *patmodels.PAT) (string, []string, []string, error) {
	oldTitle := pat.Title
	if err := s.enrichWithScope(ctx, pat); err != nil {
		return "", nil, nil, fmt.Errorf("enriching old PAT scope: %w", err)
	}
	return oldTitle, pat.RoleIDs, pat.ProjectIDs, nil
}

// replacePolicies deletes existing policies and creates new ones.
// Re-checks PAT existence after delete to guard against concurrent soft-delete.
func (s *Service) replacePolicies(ctx context.Context, patID, orgID string, roles []role.Role, projectIDs []string) error {
	if err := s.deletePolicies(ctx, patID); err != nil {
		return fmt.Errorf("deleting old policies: %w", err)
	}

	// TOCTOU guard: ensure PAT wasn't soft-deleted by concurrent Delete.
	if _, err := s.repo.GetByID(ctx, patID); err != nil {
		return fmt.Errorf("PAT deleted concurrently: %w", err)
	}

	if err := s.createPolicies(ctx, patID, orgID, roles, projectIDs); err != nil {
		return fmt.Errorf("creating new policies: %w", err)
	}
	return nil
}

// auditUpdate creates an audit record for the PAT update. Errors are logged, not returned.
func (s *Service) auditUpdate(ctx context.Context, updated patmodels.PAT, toUpdate patmodels.PAT, oldTitle string, oldRoleIDs []string, oldProjectIDs []string) {
	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATUpdatedEvent, updated, time.Now().UTC(), map[string]any{
		"role_ids":        toUpdate.RoleIDs,
		"project_ids":     toUpdate.ProjectIDs,
		"old_title":       oldTitle,
		"old_role_ids":    oldRoleIDs,
		"old_project_ids": oldProjectIDs,
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

	roles, err := s.resolveAndValidateRoles(ctx, req.RoleIDs)
	if err != nil {
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

	if err := s.createPolicies(ctx, created.ID, req.OrgID, roles, req.ProjectIDs); err != nil {
		return patmodels.PAT{}, "", fmt.Errorf("creating policies: %w", err)
	}

	// TODO: move audit record creation into the same transaction as PAT creation to avoid partial state where PAT exists but audit record doesn't.
	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATCreatedEvent, created, created.CreatedAt, map[string]any{
		"role_ids":    req.RoleIDs,
		"project_ids": req.ProjectIDs,
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

// resolveAndValidateRoles fetches the requested roles and validates they are allowed for PATs.
// All validation (existence, permissions, scopes) happens here so callers can fail fast
// before persisting any state.
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

	for _, r := range roles {
		if len(r.Scopes) == 0 {
			return nil, fmt.Errorf("role %s has no scopes defined: %w", r.Name, paterrors.ErrUnsupportedScope)
		}
		for _, scope := range r.Scopes {
			if scope != schema.ProjectNamespace && scope != schema.OrganizationNamespace {
				return nil, fmt.Errorf("role %s has scopes %v: %w", r.Name, r.Scopes, paterrors.ErrUnsupportedScope)
			}
		}
	}

	return roles, nil
}

// createPolicies creates SpiceDB policies for the PAT based on the already-validated roles and project scope.
// Each role is categorized by its Scopes field:
//   - Org-scoped role -> policy on the org with default "granted" relation
//   - Project-scoped role, all projects (projectIDs empty) -> policy on org with "pat_granted" relation
//   - Project-scoped role, specific projects -> one policy per project with default "granted" relation
func (s *Service) createPolicies(ctx context.Context, patID, orgID string, roles []role.Role, projectIDs []string) error {
	for _, r := range roles {
		var err error
		switch {
		case slices.Contains(r.Scopes, schema.ProjectNamespace):
			err = s.createProjectScopedPolicies(ctx, patID, orgID, r, projectIDs)
		case slices.Contains(r.Scopes, schema.OrganizationNamespace):
			err = s.createOrgScopedPolicy(ctx, patID, orgID, r)
		default:
			err = fmt.Errorf("role %s has scopes %v: %w", r.Name, r.Scopes, paterrors.ErrUnsupportedScope)
		}
		if err != nil {
			return err
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
func (s *Service) createOrgScopedPolicy(ctx context.Context, patID, orgID string, r role.Role) error {
	if _, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        r.ID,
		ResourceID:    orgID,
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   patID,
		PrincipalType: schema.PATPrincipal,
	}); err != nil {
		return fmt.Errorf("creating org policy for role %s: %w", r.Name, err)
	}
	return nil
}

// createProjectScopedPolicies creates policies for a project-scoped role.
// If projectIDs is empty, it creates a single policy on the org with "pat_granted" relation
// (cascades to all projects). Otherwise, it creates one policy per project with default "granted".
func (s *Service) createProjectScopedPolicies(ctx context.Context, patID, orgID string, r role.Role, projectIDs []string) error {
	if len(projectIDs) == 0 {
		// all projects -> policy on org with "pat_granted"
		if _, err := s.policyService.Create(ctx, policy.Policy{
			RoleID:        r.ID,
			ResourceID:    orgID,
			ResourceType:  schema.OrganizationNamespace,
			PrincipalID:   patID,
			PrincipalType: schema.PATPrincipal,
			GrantRelation: schema.PATGrantRelationName,
		}); err != nil {
			return fmt.Errorf("creating org pat_granted policy for role %s: %w", r.Name, err)
		}
		return nil
	}

	// specific projects -> one policy per project
	for _, projectID := range projectIDs {
		if _, err := s.policyService.Create(ctx, policy.Policy{
			RoleID:        r.ID,
			ResourceID:    projectID,
			ResourceType:  schema.ProjectNamespace,
			PrincipalID:   patID,
			PrincipalType: schema.PATPrincipal,
		}); err != nil {
			return fmt.Errorf("creating project policy for role %s on project %s: %w", r.Name, projectID, err)
		}
	}
	return nil
}

// enrichWithScope derives role_ids and project_ids from the PAT's SpiceDB policies.
func (s *Service) enrichWithScope(ctx context.Context, pat *patmodels.PAT) error {
	policies, err := s.policyService.List(ctx, policy.Filter{
		PrincipalID:   pat.ID,
		PrincipalType: schema.PATPrincipal,
	})
	if err != nil {
		return fmt.Errorf("listing policies for PAT %s: %w", pat.ID, err)
	}

	var roleIDs []string
	allProjects := false
	var projectIDs []string
	for _, pol := range policies {
		roleIDs = append(roleIDs, pol.RoleID)
		if pol.ResourceType == schema.ProjectNamespace {
			projectIDs = append(projectIDs, pol.ResourceID)
		}
		if pol.GrantRelation == schema.PATGrantRelationName {
			allProjects = true
		}
	}

	pat.RoleIDs = pkgUtils.Deduplicate(roleIDs)
	if !allProjects {
		pat.ProjectIDs = pkgUtils.Deduplicate(projectIDs)
	}
	// allProjects → pat.ProjectIDs stays nil (empty = all projects, matching create semantics)
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
