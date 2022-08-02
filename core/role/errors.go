package role

import "errors"

var (
	ErrNotExist      = errors.New("role doesn't exist")
	ErrInvalidID     = errors.New("role id is invalid")
	ErrConflict      = errors.New("role name already exist")
	ErrInvalidDetail = errors.New("invalid role detail")
)
