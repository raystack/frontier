package v1beta1

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/resource"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h Handler) CheckResourcePermission(ctx context.Context, req *shieldv1beta1.ResourceActionAuthzRequest) (*shieldv1beta1.ResourceActionAuthzResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := req.ValidateAll(); err != nil {
		formattedErr := getValidationErrorMessage(err)
		logger.Error(formattedErr.Error())
		return nil, status.Errorf(codes.NotFound, formattedErr.Error())
	}

	result, err := h.resourceService.CheckAuthz(ctx, resource.Resource{
		Name:        req.GetResourceId(),
		NamespaceID: req.GetNamespaceId(),
	}, action.Action{ID: req.GetActionId()})
	if err != nil {
		formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
		logger.Error(formattedErr.Error())
		return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
	}

	if !result {
		return nil, status.Errorf(codes.Unauthenticated, "user not allowed to make request")
	}

	return &shieldv1beta1.ResourceActionAuthzResponse{Status: "OK"}, nil
}

func getValidationErrorMessage(err error) error {
	consolidateInvalidFields := ""
	for _, validationErr := range err.(shieldv1beta1.ResourceActionAuthzRequestMultiError) {
		consolidateInvalidFields += validationErr.(shieldv1beta1.ResourceActionAuthzRequestValidationError).Field()
	}

	formattedErr := fmt.Errorf("%w: %s", ErrRequestBodyValidation, consolidateInvalidFields)
	return formattedErr
}
