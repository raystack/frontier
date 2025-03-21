package prospect

import "errors"

var (
	ErrInvalidEmail               = errors.New("invalid email")
	ErrEmailActivityAlreadyExists = errors.New("email and activity combination already exists")
	ErrNotExist                   = errors.New("prospect does not exist")
	ErrInvalidUUID                = errors.New("invalid syntax of uuid")
)
