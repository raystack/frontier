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
	grpcAudienceNotFoundErr = status.Errorf(codes.NotFound, "audience not found for the given details")
)

type AudienceService interface {
	Create(ctx context.Context, audience audience.Audience) (audience.Audience, error)
	List(ctx context.Context, filter audience.Filter) ([]audience.Audience, error)
}

func (h Handler) CreateEnrollmentForCurrentUser(ctx context.Context, request *frontierv1beta1.CreateEnrollmentForCurrentUserRequest) (*frontierv1beta1.CreateEnrollmentForCurrentUserResponse, error) {
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
	name := principal.User.Title
	subsStatus := frontierv1beta1.Audience_Status_name[int32(request.GetStatus())] // convert using proto methods
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
			return &frontierv1beta1.CreateEnrollmentForCurrentUserResponse{}, grpcConflictError
		default:
			return &frontierv1beta1.CreateEnrollmentForCurrentUserResponse{}, grpcInternalServerError
		}
	}

	transformedAudience, err := transformAudienceToPB(newAudience)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateEnrollmentForCurrentUserResponse{Audience: transformedAudience}, nil
}

func (h Handler) ListEnrollmentForCurrentUser(ctx context.Context, request *frontierv1beta1.ListEnrollmentForCurrentUserRequest) (*frontierv1beta1.ListEnrollmentForCurrentUserResponse, error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	email := principal.User.Email
	activity := request.GetActivity()

	filters := audience.Filter{Activity: activity, Email: email}

	enrollments, err := h.audienceService.List(ctx, filters)
	if err != nil {
		switch {
		case errors.Is(err, audience.ErrNotExist):
			return nil, grpcAudienceNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	var transformedAudiences []*frontierv1beta1.Audience
	for _, enrollment := range enrollments {
		transformedAudience, err := transformAudienceToPB(enrollment)
		if err != nil {
			return nil, err
		}
		transformedAudiences = append(transformedAudiences, transformedAudience)
	}
	return &frontierv1beta1.ListEnrollmentForCurrentUserResponse{Audience: transformedAudiences}, nil
}

func convertStatusToPBFormat(status audience.Status) frontierv1beta1.Audience_Status {
	switch status {
	case audience.Unsubscribed:
		return frontierv1beta1.Audience_STATUS_UNSUBSCRIBED
	case audience.Subscribed:
		return frontierv1beta1.Audience_STATUS_SUBSCRIBED
	default:
		return frontierv1beta1.Audience_STATUS_UNSUBSCRIBED
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
