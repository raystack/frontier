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
	ErrAudienceIdRequired   = status.Errorf(codes.InvalidArgument, "audience ID is required")
	grpcAudienceNotFoundErr = status.Errorf(codes.NotFound, "record not found for the given input")
)

type AudienceService interface {
	Create(ctx context.Context, audience audience.Audience) (audience.Audience, error)
	List(ctx context.Context, filter audience.Filter) ([]audience.Audience, error)
	Update(ctx context.Context, audience audience.Audience) (audience.Audience, error)
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
		Verified: true, // if user is logged in on platform them we already would have already verified the email
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

func (h Handler) ListEnrollmentsForCurrentUser(ctx context.Context, request *frontierv1beta1.ListEnrollmentsForCurrentUserRequest) (*frontierv1beta1.ListEnrollmentsForCurrentUserResponse, error) {
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
	return &frontierv1beta1.ListEnrollmentsForCurrentUserResponse{Audience: transformedAudiences}, nil
}

func (h Handler) UpdateEnrollmentForCurrentUser(ctx context.Context, request *frontierv1beta1.UpdateEnrollmentForCurrentUserRequest) (*frontierv1beta1.UpdateEnrollmentForCurrentUserResponse, error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Type != schema.UserPrincipal {
		return nil, ErrUserTypeNotSupported
	}

	audienceId := request.GetId()
	if audienceId == "" {
		return nil, ErrAudienceIdRequired
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

	updatedAudience, err := h.audienceService.Update(ctx, audience.Audience{
		ID:       audienceId,
		Name:     name,
		Email:    email,
		Verified: true,
		Activity: activity,
		Status:   audience.StringToStatus(strings.ToLower(subsStatus)),
		Source:   source,
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, audience.ErrNotExist):
			return &frontierv1beta1.UpdateEnrollmentForCurrentUserResponse{}, grpcAudienceNotFoundErr
		case errors.Is(err, audience.ErrEmailActivityAlreadyExists):
			return &frontierv1beta1.UpdateEnrollmentForCurrentUserResponse{}, grpcConflictError
		default:
			return &frontierv1beta1.UpdateEnrollmentForCurrentUserResponse{}, grpcInternalServerError
		}
	}
	transformedAudience, err := transformAudienceToPB(updatedAudience)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.UpdateEnrollmentForCurrentUserResponse{Audience: transformedAudience}, nil
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
