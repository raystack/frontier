package v1beta1connect

import (
	"github.com/raystack/frontier/pkg/errors"
)

var (
	ErrBadRequest             = errors.New("invalid syntax in body")
	ErrInvalidMetadata        = errors.New("metadata schema validation failed")
	ErrOperationUnsupported   = errors.New("operation not supported")
	ErrInternalServerError    = errors.New("internal server error")
	ErrUnauthenticated        = errors.New("not authenticated")
	ErrUnauthorized           = errors.New("not authorized")
	ErrNotFound               = errors.New("not found")
	ErrInvalidEmail           = errors.New("Invalid email")
	ErrUserNotExist           = errors.New("user doesn't exist")
	ErrInvalidNamesapceOrID   = errors.New("namespace and ID cannot be empty")
	ErrConflictRequest        = errors.New("already exist")
	ErrBadBodyMetaSchemaError = errors.New(ErrBadRequest.Error() + " : " + ErrInvalidMetadata.Error())
	ErrInvalidActorType       = errors.New("invalid actor type")
	ErrActivityRequired       = errors.New("activity is required")
	ErrStatusRequired         = errors.New("status is required")
	ErrProspectIdRequired     = errors.New("prospect ID is required")
	ErrProspectNotFound       = errors.New("record not found for the given input")
	ErrRQLParse               = errors.New("error parsing RQL query")
	ErrOrgDisabled            = errors.New("org is disabled. Please contact your administrator to enable it")
	ErrRoleFilter             = errors.New("cannot use role filters and with_roles together")
)
