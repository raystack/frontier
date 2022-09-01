package policy

import "errors"

var (
	ErrNotExist      = errors.New("policies doesn't exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("policy id is invalid")
	ErrConflict      = errors.New("policy already exist")
	ErrInvalidDetail = errors.New("invalid policy detail")
)
