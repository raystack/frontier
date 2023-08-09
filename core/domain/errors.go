package domain

import "errors"

var (
	ErrNotExist          = errors.New("org domain request does not exist")
	ErrInvalidDomain     = errors.New("invalid domain. No such host found")
	ErrTXTrecordNotFound = errors.New("required TXT record not found for domain verification")
	ErrDomainsMisMatch   = errors.New("user domain does not match the organization domain")
	ErrInvalidId         = errors.New("invalid domain id")
)
