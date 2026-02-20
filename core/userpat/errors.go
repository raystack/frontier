package userpat

import "errors"

var (
	ErrNotFound       = errors.New("personal access token not found")
	ErrConflict       = errors.New("personal access token with this name already exists")
	ErrExpired        = errors.New("personal access token has expired")
	ErrInvalidToken   = errors.New("personal access token is invalid")
	ErrLimitExceeded  = errors.New("maximum number of personal access tokens reached")
	ErrDisabled       = errors.New("personal access tokens are not enabled")
	ErrExpiryExceeded = errors.New("expiry exceeds maximum allowed lifetime")
	ErrExpiryInPast   = errors.New("expiry must be in the future")
)
