package membership

import "errors"

var (
	ErrAlreadyMember        = errors.New("principal is already a member of this resource")
	ErrNotMember            = errors.New("principal is not a member of this resource")
	ErrInvalidOrgRole       = errors.New("role is not valid for organization scope")
	ErrLastOwnerRole        = errors.New("cannot change role: this is the last owner of the organization")
	ErrInvalidPrincipal     = errors.New("invalid principal")
	ErrPrincipalNotInOrg    = errors.New("principal does not belong to this organization")
	ErrInvalidPrincipalType = errors.New("unsupported principal type")
	ErrNotOrgMember         = errors.New("principal is not a member of the organization")
	ErrInvalidProjectRole   = errors.New("role is not valid for project scope")
)
