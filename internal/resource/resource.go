package resource

import "errors"

var (
	ResourceDoesntExist = errors.New("resource doesn't exist")
	InvalidUUID         = errors.New("invalid syntax of uuid")
)
