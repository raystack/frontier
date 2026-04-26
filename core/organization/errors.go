package organization

import "errors"

var (
	ErrNotExist          = errors.New("org doesn't exist")
	ErrInvalidUUID       = errors.New("invalid syntax of uuid")
	ErrInvalidID         = errors.New("org id is invalid")
	ErrConflict          = errors.New("org already exist")
	ErrInvalidDetail     = errors.New("invalid org detail")
	ErrDisabled          = errors.New("org is disabled")
	ErrLastOwnerRole     = errors.New("cannot remove the last owner role")
	ErrNotMember         = errors.New("user is not a member of the organization")
	ErrInvalidOrgRole    = errors.New("role is not valid for organization scope")
	ErrUserPrincipalOnly = errors.New("only user principals can create organizations")
)
