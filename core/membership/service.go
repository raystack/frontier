package membership

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/log"
)

type PolicyService interface {
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	List(ctx context.Context, flt policy.Filter) ([]policy.Policy, error)
	Delete(ctx context.Context, id string) error
}

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type RoleService interface {
	Get(ctx context.Context, idOrName string) (role.Role, error)
}

type OrgService interface {
	Get(ctx context.Context, idOrName string) (organization.Organization, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord auditrecord.AuditRecord) (auditrecord.AuditRecord, error)
}

type Service struct {
	log                   log.Logger
	policyService         PolicyService
	relationService       RelationService
	roleService           RoleService
	orgService            OrgService
	userService           UserService
	auditRecordRepository AuditRecordRepository
}

func NewService(
	logger log.Logger,
	policyService PolicyService,
	relationService RelationService,
	roleService RoleService,
	orgService OrgService,
	userService UserService,
	auditRecordRepository AuditRecordRepository,
) *Service {
	return &Service{
		log:                   logger,
		policyService:         policyService,
		relationService:       relationService,
		roleService:           roleService,
		orgService:            orgService,
		userService:           userService,
		auditRecordRepository: auditRecordRepository,
	}
}

// AddOrganizationMember adds a new member to an organization with an explicit role.
// Returns ErrAlreadyMember if the principal already has a policy on this org.
// Unlike the old AddOrganizationUsers RPC, this requires an explicit role_id
// and supports all principal types (user, serviceuser).
func (s *Service) AddOrganizationMember(ctx context.Context, orgID, principalID, principalType, roleID string) error {
	// orgService.Get returns ErrDisabled for disabled orgs
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	usr, err := s.userService.GetByID(ctx, principalID)
	if err != nil {
		return err
	}
	if usr.State == user.Disabled {
		return user.ErrDisabled
	}

	fetchedRole, err := s.validateOrgRole(ctx, roleID, orgID)
	if err != nil {
		return err
	}

	// check if principal is already a member
	existing, err := s.policyService.List(ctx, policy.Filter{
		OrgID:         orgID,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("list existing policies: %w", err)
	}
	if len(existing) > 0 {
		return ErrAlreadyMember
	}

	createdPolicy, err := s.createPolicy(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, roleID)
	if err != nil {
		return err
	}

	relationName := orgRoleToRelation(fetchedRole)
	if err := s.createRelation(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, relationName); err != nil {
		// best-effort cleanup to avoid orphaned policy
		if deleteErr := s.policyService.Delete(ctx, createdPolicy.ID); deleteErr != nil {
			s.log.Warn("orphaned policy: relation creation failed and policy cleanup also failed",
				"policy_id", createdPolicy.ID,
				"org_id", orgID,
				"principal_id", principalID,
				"policy_delete_error", deleteErr,
			)
		}
		return err
	}

	// audit logging
	s.auditOrgMemberAdded(ctx, org, usr, roleID)

	return nil
}

// validateOrgRole checks that the role is valid for organization scope and returns it.
// A role is valid if it is either:
// - a platform-wide role scoped to organizations, or
// - a custom role created for this specific organization.
func (s *Service) validateOrgRole(ctx context.Context, roleID, orgID string) (role.Role, error) {
	fetchedRole, err := s.roleService.Get(ctx, roleID)
	if err != nil {
		return role.Role{}, err
	}

	// custom role belonging to this org
	if fetchedRole.OrgID == orgID {
		return fetchedRole, nil
	}

	// platform-wide role with org scope
	if utils.IsNullUUID(fetchedRole.OrgID) && slices.Contains(fetchedRole.Scopes, schema.OrganizationNamespace) {
		return fetchedRole, nil
	}

	return role.Role{}, ErrInvalidOrgRole
}

// orgRoleToRelation maps an org role to the corresponding SpiceDB relation name.
// Owner role gets "owner" relation, everything else gets "member" relation.
func orgRoleToRelation(r role.Role) string {
	if r.Name == schema.RoleOrganizationOwner {
		return schema.OwnerRelationName
	}
	return schema.MemberRelationName
}

// replacePolicy deletes all existing policies for the principal+resource and creates a new one.
func (s *Service) createPolicy(ctx context.Context, resourceID, resourceType, principalID, principalType, roleID string) (policy.Policy, error) {
	created, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        roleID,
		ResourceID:    resourceID,
		ResourceType:  resourceType,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return policy.Policy{}, fmt.Errorf("create policy: %w", err)
	}
	return created, nil
}

func (s *Service) createRelation(ctx context.Context, resourceID, resourceType, principalID, principalType, relationName string) error {
	if _, err := s.relationService.Create(ctx, relation.Relation{
		Object:       relation.Object{ID: resourceID, Namespace: resourceType},
		Subject:      relation.Subject{ID: principalID, Namespace: principalType},
		RelationName: relationName,
	}); err != nil {
		return fmt.Errorf("create relation: %w", err)
	}
	return nil
}

func (s *Service) auditOrgMemberAdded(ctx context.Context, org organization.Organization, usr user.User, roleID string) {
	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:   usr.ID,
			Type: pkgAuditRecord.UserType,
			Name: usr.Title,
			Metadata: map[string]any{
				"email":   usr.Email,
				"role_id": roleID,
			},
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})

	audit.GetAuditor(ctx, org.ID).LogWithAttrs(audit.OrgMemberCreatedEvent, audit.Target{
		ID:   usr.ID,
		Type: schema.UserPrincipal,
	}, map[string]string{
		"role_id": roleID,
	})
}
