package product

import (
	"errors"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrPriceNotFound   = errors.New("price not found")
	ErrInvalidDetail   = errors.New("invalid product detail")

	ErrInvalidFeatureDetail = errors.New("invalid feature detail")
	ErrFeatureNotFound      = errors.New("feature not found")
)
