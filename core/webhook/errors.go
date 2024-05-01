package webhook

import "errors"

var (
	ErrNotFound      = errors.New("webhook doesn't exist")
	ErrInvalidDetail = errors.New("invalid webhook details")
	ErrConflict      = errors.New("webhook already exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrDisabled      = errors.New("webhook is disabled")
)
