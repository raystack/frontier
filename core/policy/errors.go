package policy

import "errors"

var (
	ErrNotExist      = errors.New("policies doesn't exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("policy id is invalid")
	ErrConflict      = errors.New("policy already exist")
	ErrInvalidDetail = errors.New("invalid policy detail")
	ErrLastRoleGuard = errors.New("cannot delete: this is the last policy with the guarded role for this resource")
	ErrInvalidFilter = errors.New("invalid policy filter: PrincipalType set without PrincipalID or PrincipalIDs")
)
