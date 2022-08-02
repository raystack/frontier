package v1beta1

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP Codes defined here:
// https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L36
var (
	ErrInternalServer   = errors.New("internal server error")
	ErrBadRequest       = errors.New("invalid syntax in body")
	ErrConflictRequest  = errors.New("already exist")
	ErrPermissionDenied = errors.New("permission denied")

	grpcInternalServerError = status.Errorf(codes.Internal, ErrInternalServer.Error())
	grpcConflictError       = status.Errorf(codes.AlreadyExists, ErrConflictRequest.Error())
	grpcBadBodyError        = status.Error(codes.InvalidArgument, ErrBadRequest.Error())
	grpcPermissionDenied    = status.Error(codes.PermissionDenied, ErrPermissionDenied.Error())

	ErrEmptyEmailID = errors.New("email id is empty")
)
