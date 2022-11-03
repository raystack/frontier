package relation

import "errors"

var (
	ErrNotExist                      = errors.New("relation doesn't exist")
	ErrInvalidUUID                   = errors.New("invalid syntax of uuid")
	ErrInvalidID                     = errors.New("relation id is invalid")
	ErrConflict                      = errors.New("relation already exist")
	ErrInvalidDetail                 = errors.New("invalid relation detail")
	ErrCreatingRelationInStore       = errors.New("error while creating relation")
	ErrCreatingRelationInAuthzEngine = errors.New("error while creating relation")
	ErrFetchingUser                  = errors.New("error while fetching user")
)
