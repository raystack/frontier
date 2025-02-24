package v1beta1

import (
	"context"
	"errors"
	"strings"

	"github.com/raystack/frontier/core/audience"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUserTypeNotSupported = status.Errorf(codes.InvalidArgument, "user type not supported")
	ErrActivityRequired     = status.Errorf(codes.InvalidArgument, "activity is required")
	ErrSourceRequired       = status.Errorf(codes.InvalidArgument, "source is required")
)

type AudienceService interface {
	Create(ctx context.Context, audience audience.Audience) (audience.Audience, error)
}

func (h Handler) CreateAudience(ctx context.Context, request *frontierv1beta1.CreateAudienceRequest) (*frontierv1beta1.CreateAudienceResponse, error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Type != schema.UserPrincipal {
		return nil, ErrUserTypeNotSupported
	}

	activity := strings.TrimSpace(request.GetActivity())
	if activity == "" {
		return nil, ErrActivityRequired
	}
	source := request.GetSource()
	if source == "" {
		return nil, ErrSourceRequired
	}

	email := principal.User.Email
	name := principal.User.Name
	subsStatus := frontierv1beta1.SUBSCRIPTION_STATUS_name[int32(request.GetStatus())] // convert using proto methods
	metaDataMap := metadata.Build(request.GetMetadata().AsMap())

	newAudience, err := h.audienceService.Create(ctx, audience.Audience{
		Name:     name,
		Email:    email,
		Activity: activity,
		Status:   audience.StringToStatus(strings.ToLower(subsStatus)),
		Source:   source,
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, audience.ErrEmailActivityAlreadyExists):
			return &frontierv1beta1.CreateAudienceResponse{}, grpcConflictError
		default:
			return &frontierv1beta1.CreateAudienceResponse{}, grpcInternalServerError
		}
	}

	transformedAudience, err := transformAudienceToPB(newAudience)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateAudienceResponse{Audience: transformedAudience}, nil
}

func convertStatusToPBFormat(status audience.Status) frontierv1beta1.SUBSCRIPTION_STATUS {
	switch status {
	case audience.Unsubscribed:
		return frontierv1beta1.SUBSCRIPTION_STATUS_UNSUBSCRIBED
	case audience.Subscribed:
		return frontierv1beta1.SUBSCRIPTION_STATUS_SUBSCRIBED
	default:
		return frontierv1beta1.SUBSCRIPTION_STATUS_UNSUBSCRIBED
	}
}

func transformAudienceToPB(audience audience.Audience) (*frontierv1beta1.Audience, error) {
	metaData, err := audience.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Audience{}, err
	}
	return &frontierv1beta1.Audience{
		Id:        audience.ID,
		Name:      audience.Name,
		Email:     audience.Email,
		Phone:     audience.Phone,
		Activity:  audience.Activity,
		Status:    convertStatusToPBFormat(audience.Status),
		ChangedAt: timestamppb.New(audience.ChangedAt),
		Source:    audience.Source,
		Verified:  audience.Verified,
		CreatedAt: timestamppb.New(audience.CreatedAt),
		UpdatedAt: timestamppb.New(audience.UpdatedAt),
		Metadata:  metaData,
	}, nil
}
