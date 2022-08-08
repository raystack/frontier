package project

import "errors"

var (
	ErrNotExist    = errors.New("project doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
	ErrInvalidID   = errors.New("project id is invalid")
	ErrConflict    = errors.New("project already exist")
)
