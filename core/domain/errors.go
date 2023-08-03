package domain

import "errors"

var (
	ErrNotExist        = errors.New("org domain request does not exist")
	ErrDomainsMisMatch = errors.New("user domain does not match the organization domain")
	ErrInvalidId       = errors.New("invalid domain id")
)
