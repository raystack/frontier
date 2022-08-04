package v1beta1

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP Codes defined here:
// https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L36
var (
	internalServerError   = errors.New("internal server error")
	badRequestError       = errors.New("invalid syntax in body")
	permissionDeniedError = errors.New("permission denied")

	grpcInternalServerError = status.Errorf(codes.Internal, internalServerError.Error())
	grpcConflictError       = status.Errorf(codes.AlreadyExists, badRequestError.Error())
	grpcBadBodyError        = status.Error(codes.InvalidArgument, badRequestError.Error())
	grpcPermissionDenied    = status.Error(codes.PermissionDenied, permissionDeniedError.Error())
)
