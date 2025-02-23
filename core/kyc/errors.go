package kyc

import "errors"

var (
	ErrNotExist       = errors.New("org kyc doesn't exist")
	ErrKycLinkNotSet  = errors.New("link cannot be empty")
	ErrInvalidUUID    = errors.New("invalid syntax of uuid")
	ErrOrgDoesntExist = errors.New("org doesn't exist")
)
