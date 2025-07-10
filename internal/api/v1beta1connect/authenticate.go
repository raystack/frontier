package v1beta1connect

import (
	"context"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var grpcUserNotFoundError = status.Errorf(codes.NotFound, "user doesn't exist")
var grpcUnauthenticated = status.Error(codes.Unauthenticated, errors.ErrUnauthenticated.Error())

func (h ConnectHandler) GetLoggedInPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error) {
	principal, err := h.authnService.GetPrincipal(ctx, via...)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidEmail):
			return principal, grpcUserNotFoundError
		case errors.Is(err, errors.ErrUnauthenticated):
			return principal, grpcUnauthenticated
		default:
			return principal, err
		}
	}
	return principal, nil
}
