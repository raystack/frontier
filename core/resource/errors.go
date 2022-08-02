package resource

import "errors"

var (
	ErrNotExist      = errors.New("resource doesn't exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("resource id is invalid")
	ErrInvalidURN    = errors.New("resource urn is invalid")
	ErrConflict      = errors.New("resource already exist")
	ErrInvalidDetail = errors.New("invalid resource detail")
)
