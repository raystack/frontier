package serviceuser

import "errors"

var (
	ErrNotExist     = errors.New("service user doesn't exist")
	ErrCredNotExist = errors.New("service user credential doesn't exist")
	ErrInvalidID    = errors.New("service user id is invalid")
	ErrConflict     = errors.New("service user already exist")
	ErrEmptyKey     = errors.New("empty key")
	ErrDisabled     = errors.New("service user is disabled")
)
