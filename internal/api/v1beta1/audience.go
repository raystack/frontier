package v1beta1

import (
	"context"
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
	UserTypeNotSupported = status.Errorf(codes.InvalidArgument, "user type not supported")
	ActivityRequired     = status.Errorf(codes.InvalidArgument, "activity is required")
	SourceRequired       = status.Errorf(codes.InvalidArgument, "source is required")
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
		return nil, UserTypeNotSupported
	}

	activity := strings.TrimSpace(request.GetActivity())
	if activity == "" {
		return nil, ActivityRequired
	}
	source := request.GetSource()
	if source == "" {
		return nil, SourceRequired
	}

	email := principal.User.Email // validate email here??
	name := principal.User.Name
	subsStatus := request.GetStatus()
	metaDataMap := metadata.Build(request.GetMetadata().AsMap())

	newAudience, err := h.audienceService.Create(ctx, audience.Audience{
		Name:     name,
		Email:    email,
		Activity: activity,
		Status:   audience.Status(subsStatus),
		Source:   source,
		Metadata: metaDataMap,
	})

	if err != nil {
		return nil, err
	}

	transformedAudience, err := transformAudienceToPB(newAudience)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateAudienceResponse{Audience: transformedAudience}, nil
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
		Status:    0,
		ChangedAt: timestamppb.New(*audience.ChangedAt),
		Source:    audience.Source,
		Verified:  false,
		CreatedAt: timestamppb.New(audience.CreatedAt),
		UpdatedAt: timestamppb.New(audience.UpdatedAt),
		Metadata:  metaData,
	}, nil
}
