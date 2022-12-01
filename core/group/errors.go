package group

import "errors"

var (
	ErrNotExist              = errors.New("group doesn't exist")
	ErrInvalidUUID           = errors.New("invalid syntax of uuid")
	ErrInvalidID             = errors.New("group id is invalid")
	ErrConflict              = errors.New("group already exist")
	ErrInvalidDetail         = errors.New("invalid group detail")
	ErrListingGroupRelations = errors.New("error while listing relations")
	ErrFetchingUsers         = errors.New("error while fetching users")
	ErrFetchingGroups        = errors.New("error while fetching groups")
)
