package user

import "errors"

var (
	ErrNotExist         = errors.New("user doesn't exist")
	ErrInvalidID        = errors.New("user id is invalid")
	ErrInvalidEmail     = errors.New("user email is invalid")
	ErrConflict         = errors.New("user already exist")
	ErrEmptyKey         = errors.New("empty key")
	ErrKeyAlreadyExists = errors.New("key already exist")
	ErrKeyDoesNotExists = errors.New("key does not exist")
	ErrMissingEmail     = errors.New("user email is missing")
	ErrInvalidUUID      = errors.New("invalid syntax of uuid")
)
