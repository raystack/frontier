package v1beta1connect

import (
	"github.com/raystack/frontier/pkg/errors"
)

var (
	ErrBadRequest           = errors.New("invalid syntax in body")
	ErrOperationUnsupported = errors.New("operation not supported")
	ErrInternalServerError  = errors.New("internal server error")
	ErrUnauthenticated      = errors.New("not authenticated")
	ErrUnauthorized         = errors.New("not authorized")
	ErrNotFound             = errors.New("not found")
	ErrInvalidEmail         = errors.New("Invalid email")
	ErrUserNotExist         = errors.New("user doesn't exist")
	ErrInvalidNamesapceOrID = errors.New("namespace and ID cannot be empty")
)
