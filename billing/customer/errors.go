package customer

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("customer not found")
	ErrInvalidUUID    = errors.New("invalid syntax of uuid")
	ErrInvalidID      = errors.New("billing customer id is invalid")
	ErrConflict       = errors.New("customer already exist")
	ErrActiveConflict = errors.New("an active account already exists for the organization")
	ErrInvalidDetail  = errors.New("invalid billing customer detail")
	ErrDisabled       = errors.New("billing customer is disabled")
)
