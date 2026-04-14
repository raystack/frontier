package membership

import "errors"

var (
	ErrAlreadyMember  = errors.New("principal is already a member of this resource")
	ErrNotMember      = errors.New("principal is not a member of this resource")
	ErrInvalidOrgRole = errors.New("role is not valid for organization scope")
	ErrLastOwnerRole  = errors.New("cannot change role: this is the last owner of the organization")
)
