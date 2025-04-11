package user

import "errors"

var (
	ErrNotExist         = errors.New("user doesn't exist")
	ErrInvalidID        = errors.New("user id is invalid")
	ErrInvalidEmail     = errors.New("user email is invalid")
	ErrConflict         = errors.New("user already exist")
	ErrInvalidDetails   = errors.New("invalid user details (name|email)")
	ErrMissingName      = errors.New("user name is missing")
	ErrEmptyKey         = errors.New("empty key")
	ErrKeyAlreadyExists = errors.New("key already exist")
	ErrKeyDoesNotExists = errors.New("key does not exist")
	ErrMissingEmail     = errors.New("user email is missing")
	ErrInvalidUUID      = errors.New("invalid syntax of uuid")
	ErrDisabled         = errors.New("user is disabled")
	ErrNoUsersFound     = errors.New("no users found")
)
