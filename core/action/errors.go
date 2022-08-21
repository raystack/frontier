package action

import "errors"

var (
	ErrInvalidID     = errors.New("action id is invalid")
	ErrNotExist      = errors.New("action doesn't exist")
	ErrInvalidDetail = errors.New("invalid action detail")
)
