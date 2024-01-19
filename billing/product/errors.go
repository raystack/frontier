package product

import (
	"errors"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrPriceNotFound   = errors.New("price not found")
	ErrInvalidDetail   = errors.New("invalid product detail")

	ErrPerSeatLimitReached = errors.New("per seat limit reached")

	ErrInvalidFeatureDetail = errors.New("invalid feature detail")
	ErrFeatureNotFound      = errors.New("feature not found")
)
