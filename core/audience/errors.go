package audience

import "errors"

var (
	ErrInvalidEmail               = errors.New("invalid email")
	ErrEmailActivityAlreadyExists = errors.New("email and activity combination already exists")
)
