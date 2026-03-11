package errors

import "errors"

var (
	ErrNotFound         = errors.New("personal access token not found")
	ErrConflict         = errors.New("personal access token with this name already exists")
	ErrExpired          = errors.New("personal access token has expired")
	ErrInvalidPAT       = errors.New("not a personal access token")
	ErrMalformedPAT     = errors.New("personal access token is malformed")
	ErrLimitExceeded    = errors.New("maximum number of personal access tokens reached")
	ErrDisabled         = errors.New("personal access tokens are not enabled")
	ErrExpiryExceeded   = errors.New("expiry exceeds maximum allowed lifetime")
	ErrExpiryInPast     = errors.New("expiry must be in the future")
	ErrDeniedRole       = errors.New("one or more requested roles not permissible for personal access tokens")
	ErrUnsupportedScope = errors.New("role scope is not supported for personal access tokens")
	ErrRoleNotFound     = errors.New("one or more requested roles do not exist")
)
