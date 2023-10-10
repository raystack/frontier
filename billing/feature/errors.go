package feature

import (
	"errors"
)

var (
	ErrFeatureNotFound = errors.New("feature not found")
	ErrPriceNotFound   = errors.New("price not found")
	ErrInvalidDetail   = errors.New("invalid plan detail")
)
