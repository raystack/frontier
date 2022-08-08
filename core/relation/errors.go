package relation

import "errors"

var (
	ErrNotExist    = errors.New("relation doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
	ErrInvalidID   = errors.New("relation id is invalid")
	ErrConflict    = errors.New("relation already exist")
)
