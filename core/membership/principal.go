package membership

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// principalInfo holds validated principal details for audit and downstream use.
type principalInfo struct {
	ID    string
	Type  string
	Name  string
	Email string
}

// validatePrincipal checks that the principal exists, is active, and belongs to
// the target org. For users, org membership is checked separately via policies.
// For service users, org ownership is validated here since they have a fixed OrgID.
func (s *Service) validatePrincipal(ctx context.Context, orgID, principalID, principalType string) (principalInfo, error) {
	switch principalType {
	case schema.UserPrincipal:
		usr, err := s.userService.GetByID(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if usr.State == user.Disabled {
			return principalInfo{}, user.ErrDisabled
		}
		return principalInfo{
			ID:    usr.ID,
			Type:  schema.UserPrincipal,
			Name:  usr.Title,
			Email: usr.Email,
		}, nil
	case schema.ServiceUserPrincipal:
		su, err := s.serviceuserService.Get(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if su.OrgID != orgID {
			return principalInfo{}, ErrPrincipalNotInOrg
		}
		if su.State == string(serviceuser.Disabled) {
			return principalInfo{}, serviceuser.ErrDisabled
		}
		return principalInfo{
			ID:   su.ID,
			Type: schema.ServiceUserPrincipal,
			Name: su.Title,
		}, nil
	case schema.PATPrincipal:
		if s.userPATService == nil {
			return principalInfo{}, ErrInvalidPrincipal
		}
		pat, err := s.userPATService.GetByID(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if pat.OrgID != orgID {
			return principalInfo{}, ErrPrincipalNotInOrg
		}
		if !pat.ExpiresAt.After(time.Now()) {
			return principalInfo{}, ErrPrincipalExpired
		}
		return principalInfo{
			ID:   pat.ID,
			Type: schema.PATPrincipal,
			Name: pat.Title,
		}, nil
	default:
		return principalInfo{}, ErrInvalidPrincipal
	}
}

// validateOrgMembership checks that the principal exists and belongs to the given org.
// For users, org membership is verified via org-level policies.
// For service users and groups, org membership is verified via their org ID field.
func (s *Service) validateOrgMembership(ctx context.Context, orgID, principalID, principalType string) error {
	switch principalType {
	case schema.UserPrincipal:
		usr, err := s.userService.GetByID(ctx, principalID)
		if err != nil {
			return err
		}
		if usr.State == user.Disabled {
			return user.ErrDisabled
		}
		orgPolicies, err := s.policyService.List(ctx, policy.Filter{
			OrgID:         orgID,
			PrincipalID:   principalID,
			PrincipalType: principalType,
		})
		if err != nil {
			return err
		}
		if len(orgPolicies) == 0 {
			return ErrNotOrgMember
		}
	case schema.ServiceUserPrincipal:
		su, err := s.serviceuserService.Get(ctx, principalID)
		if err != nil {
			return err
		}
		if su.OrgID != orgID {
			return ErrNotOrgMember
		}
	case schema.GroupPrincipal:
		grp, err := s.groupService.Get(ctx, principalID)
		if err != nil {
			return err
		}
		if grp.OrganizationID != orgID {
			return ErrNotOrgMember
		}
	case schema.PATPrincipal:
		if s.userPATService == nil {
			return ErrInvalidPrincipal
		}
		pat, err := s.userPATService.GetByID(ctx, principalID)
		if err != nil {
			return err
		}
		if pat.OrgID != orgID {
			return ErrNotOrgMember
		}
		if !pat.ExpiresAt.After(time.Now()) {
			return ErrPrincipalExpired
		}
	default:
		return ErrInvalidPrincipalType
	}
	return nil
}

// validateGroupPrincipal fetches and validates the principal for group operations.
// Currently only app/user is supported; the switch is structured so future principal
// types (e.g. serviceuser) can be enabled here without touching call sites.
func (s *Service) validateGroupPrincipal(ctx context.Context, principalID, principalType string) (principalInfo, error) {
	switch principalType {
	case schema.UserPrincipal:
		usr, err := s.userService.GetByID(ctx, principalID)
		if err != nil {
			return principalInfo{}, err
		}
		if usr.State == user.Disabled {
			return principalInfo{}, user.ErrDisabled
		}
		return principalInfo{
			ID:    usr.ID,
			Type:  schema.UserPrincipal,
			Name:  usr.Title,
			Email: usr.Email,
		}, nil
	default:
		return principalInfo{}, ErrInvalidPrincipalType
	}
}
