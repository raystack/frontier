package organization

import "errors"

var (
	ErrNotExist      = errors.New("org doesn't exist")
	ErrNoAdminsExist = errors.New("no admins exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("org id is invalid")
)
