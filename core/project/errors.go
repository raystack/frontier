package project

import "errors"

var (
	ErrNotExist             = errors.New("project or its relations doesn't exist")
	ErrInvalidUUID          = errors.New("invalid syntax of uuid")
	ErrInvalidID            = errors.New("project id is invalid")
	ErrConflict             = errors.New("project already exist")
	ErrInvalidDetail        = errors.New("invalid project detail")
	ErrInvalidProjectRole   = errors.New("role is not valid for project scope")
	ErrNotOrgMember         = errors.New("user is not a member of the organization")
	ErrInvalidPrincipalType = errors.New("invalid principal type")
)
