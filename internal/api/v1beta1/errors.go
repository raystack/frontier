package v1beta1

import (
	"github.com/raystack/frontier/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP Codes defined here:
// https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L36
var (
	ErrBadRequest            = errors.New("invalid syntax in body")
	ErrInvalidMetadata       = errors.New("metadata schema validation failed")
	ErrConflictRequest       = errors.New("already exist")
	ErrRequestBodyValidation = errors.New("invalid format for field(s)")
	ErrEmptyEmailID          = errors.New("email id is empty")
	ErrEmailConflict         = errors.New("user email can't be updated")
	ErrOperationUnsupported  = errors.New("operation not supported")
	ErrInternalServerError   = errors.New("internal server error")

	grpcConflictError          = status.Errorf(codes.AlreadyExists, ErrConflictRequest.Error())
	grpcBadBodyError           = status.Error(codes.InvalidArgument, ErrBadRequest.Error())
	grpcBadBodyMetaSchemaError = status.Error(codes.InvalidArgument, ErrBadRequest.Error()+" : "+ErrInvalidMetadata.Error())
	grpcUnauthenticated        = status.Error(codes.Unauthenticated, errors.ErrUnauthenticated.Error())
	grpcPermissionDenied       = status.Error(codes.PermissionDenied, errors.ErrForbidden.Error())
	grpcOperationUnsupported   = status.Error(codes.Unavailable, ErrOperationUnsupported.Error()) //nolint:unused
	grpcInternalServerError    = status.Errorf(codes.Internal, ErrInternalServerError.Error())
)

func ErrInvalidInput(err string) error {
	return status.Errorf(codes.InvalidArgument, err)
}
