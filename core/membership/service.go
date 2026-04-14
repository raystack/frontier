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
	policyService         PolicyService
	relationService       RelationService
	roleService           RoleService
	orgService            OrgService
	userService           UserService
	auditRecordRepository AuditRecordRepository
}

func NewService(
	policyService PolicyService,
	relationService RelationService,
	roleService RoleService,
	orgService OrgService,
	userService UserService,
	auditRecordRepository AuditRecordRepository,
) *Service {
	return &Service{
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
	org, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
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

	if err := s.createPolicy(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, roleID); err != nil {
		return err
	}

	relationName := orgRoleToRelation(fetchedRole)
	if err := s.createRelation(ctx, orgID, schema.OrganizationNamespace, principalID, principalType, relationName); err != nil {
		return err
	}

	// audit logging
	s.auditOrgMemberAdded(ctx, org, principalID, principalType, roleID)

	return nil
}

// validateOrgRole checks that the role is valid for organization scope and returns it.
func (s *Service) validateOrgRole(ctx context.Context, roleID, orgID string) (role.Role, error) {
	fetchedRole, err := s.roleService.Get(ctx, roleID)
	if err != nil {
		return role.Role{}, err
	}

	isGlobalRole := utils.IsNullUUID(fetchedRole.OrgID)
	isGlobalOrgRole := isGlobalRole && slices.Contains(fetchedRole.Scopes, schema.OrganizationNamespace)
	isOrgSpecificRole := fetchedRole.OrgID == orgID
	if !isGlobalOrgRole && !isOrgSpecificRole {
		return role.Role{}, ErrInvalidOrgRole
	}

	return fetchedRole, nil
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
func (s *Service) createPolicy(ctx context.Context, resourceID, resourceType, principalID, principalType, roleID string) error {
	_, err := s.policyService.Create(ctx, policy.Policy{
		RoleID:        roleID,
		ResourceID:    resourceID,
		ResourceType:  resourceType,
		PrincipalID:   principalID,
		PrincipalType: principalType,
	})
	if err != nil {
		return fmt.Errorf("create policy: %w", err)
	}
	return nil
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

func (s *Service) auditOrgMemberAdded(ctx context.Context, org organization.Organization, principalID, principalType, roleID string) {
	var targetName string
	targetMetadata := map[string]any{
		"role_id": roleID,
	}

	if principalType == schema.UserPrincipal {
		if usr, err := s.userService.GetByID(ctx, principalID); err == nil {
			targetName = usr.Title
			targetMetadata["email"] = usr.Email
		}
	}

	var auditEntityType pkgAuditRecord.EntityType
	switch principalType {
	case schema.ServiceUserPrincipal:
		auditEntityType = pkgAuditRecord.ServiceUserType
	default:
		auditEntityType = pkgAuditRecord.UserType
	}

	s.auditRecordRepository.Create(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:       principalID,
			Type:     auditEntityType,
			Name:     targetName,
			Metadata: targetMetadata,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})

	audit.GetAuditor(ctx, org.ID).LogWithAttrs(audit.OrgMemberCreatedEvent, audit.Target{
		ID:   principalID,
		Type: principalType,
	}, map[string]string{
		"role_id": roleID,
	})
}
