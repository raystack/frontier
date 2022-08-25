package user

import "errors"

var (
	ErrNotExist         = errors.New("user doesn't exist")
	ErrInvalidID        = errors.New("user id is invalid")
	ErrInvalidEmail     = errors.New("user email is invalid")
	ErrNotUUID          = errors.New("invalid syntax of uuid")
	ErrConflict         = errors.New("user already exist")
	ErrEmptyKey         = errors.New("empty key")
	ErrKeyAlreadyExists = errors.New("key already exist")
)
