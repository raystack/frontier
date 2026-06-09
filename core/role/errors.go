package role

import "errors"

var (
	ErrNotExist      = errors.New("role doesn't exist")
	ErrInvalidID     = errors.New("role id is invalid")
	ErrConflict      = errors.New("role name already exist")
	ErrInvalidDetail = errors.New("invalid role detail")
	ErrRoleInUse     = errors.New("role is in use by one or more policies")
)
