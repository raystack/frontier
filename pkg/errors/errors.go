package errors

import "errors"

// These aliased values are added to avoid conflicting imports of standard `errors`
// package and this `errors` package where these functions are needed.
var (
	Is  = errors.Is
	As  = errors.As
	New = errors.New
)

var (
	Unauthorized   = errors.New("you are not authorized")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)
