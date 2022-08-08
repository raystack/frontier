package role

import "errors"

var (
	ErrNotExist    = errors.New("role doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
	ErrInvalidID   = errors.New("role id is invalid")
	ErrConflict    = errors.New("role name already exist")
)
