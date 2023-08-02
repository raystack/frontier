package organization

import "errors"

var (
	ErrNotExist        = errors.New("org doesn't exist")
	ErrInvalidUUID     = errors.New("invalid syntax of uuid")
	ErrInvalidID       = errors.New("org id is invalid")
	ErrConflict        = errors.New("org already exist")
	ErrInvalidDetail   = errors.New("invalid org detail")
	ErrDomainsNotMatch = errors.New("user domain does not match the organization domain")
)
