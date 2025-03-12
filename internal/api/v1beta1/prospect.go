package v1beta1

import (
	"context"
	"errors"
	"strings"

	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	grpcUserTypeNotSupportedErr = status.Errorf(codes.InvalidArgument, "user type not supported") //nolint:unused
	grpcActivityRequiredError   = status.Errorf(codes.InvalidArgument, "activity is required")
	grpcEmailInvalidError       = status.Errorf(codes.InvalidArgument, "email is invalid")
	grpcStatusRequiredError     = status.Errorf(codes.InvalidArgument, "status is required")
	grpcProspectIdRequiredError = status.Errorf(codes.InvalidArgument, "prospect ID is required")
	grpcProspectNotFoundError   = status.Errorf(codes.NotFound, "record not found for the given input")
	grpcRQLParseError           = status.Errorf(codes.NotFound, "error parsing RQL query")
	grpcRQLInvalidError         = status.Errorf(codes.NotFound, "datatype or operator not supported in RQL query")
)

type ProspectService interface {
	Create(ctx context.Context, prospect prospect.Prospect) (prospect.Prospect, error)
	List(ctx context.Context, query *rql.Query) (prospect.ListProspects, error)
	Get(ctx context.Context, prospectId string) (prospect.Prospect, error)
	Update(ctx context.Context, prospect prospect.Prospect) (prospect.Prospect, error)
	Delete(ctx context.Context, prospectId string) error
}

func (h Handler) CreateProspectPublic(ctx context.Context, request *frontierv1beta1.CreateProspectPublicRequest) (*frontierv1beta1.CreateProspectPublicResponse, error) {
	email := request.GetEmail()
	if !isValidEmail(email) {
		return nil, grpcEmailInvalidError
	}
	activity := strings.TrimSpace(request.GetActivity())
	if activity == "" {
		return nil, grpcActivityRequiredError
	}
	metaDataMap, err := buildAndValidateMetadata(request.GetMetadata().AsMap(), h)
	if err != nil {
		return nil, err
	}

	_, err = h.prospectService.Create(ctx, prospect.Prospect{
		Name:     request.GetName(),
		Email:    strings.ToLower(email),
		Phone:    request.GetPhone(),
		Activity: strings.TrimSpace(activity),
		Status:   prospect.Subscribed, // Subscribed by default
		Verified: false,
		Source:   request.GetSource(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, prospect.ErrEmailActivityAlreadyExists):
			return &frontierv1beta1.CreateProspectPublicResponse{}, nil
		default:
			return &frontierv1beta1.CreateProspectPublicResponse{}, grpcInternalServerError
		}
	}
	return &frontierv1beta1.CreateProspectPublicResponse{}, nil
}

func (h Handler) CreateProspect(ctx context.Context, request *frontierv1beta1.CreateProspectRequest) (*frontierv1beta1.CreateProspectResponse, error) {
	email := request.GetEmail()
	if !isValidEmail(email) {
		return nil, grpcEmailInvalidError
	}
	activity := strings.TrimSpace(request.GetActivity())
	if activity == "" {
		return nil, grpcActivityRequiredError
	}
	reqStatus := request.GetStatus()
	if reqStatus == frontierv1beta1.Prospect_STATUS_UNSPECIFIED {
		return nil, grpcStatusRequiredError
	}
	subsStatus := frontierv1beta1.Prospect_Status_name[int32(reqStatus)] // convert using proto methods
	metaDataMap, err := buildAndValidateMetadata(request.GetMetadata().AsMap(), h)
	if err != nil {
		return nil, err
	}

	newProspect, err := h.prospectService.Create(ctx, prospect.Prospect{
		Name:     request.GetName(),
		Email:    strings.ToLower(email),
		Phone:    request.GetPhone(),
		Activity: strings.TrimSpace(activity),
		Status:   prospect.StringToStatus(strings.ToLower(subsStatus)),
		Verified: request.GetVerified(),
		Source:   request.GetSource(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, prospect.ErrEmailActivityAlreadyExists):
			return &frontierv1beta1.CreateProspectResponse{}, grpcConflictError
		default:
			return &frontierv1beta1.CreateProspectResponse{}, grpcInternalServerError
		}
	}

	transformedProspect, err := transformProspectToPB(newProspect)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateProspectResponse{Prospect: transformedProspect}, nil
}

func (h Handler) ListProspects(ctx context.Context, request *frontierv1beta1.ListProspectsRequest) (*frontierv1beta1.ListProspectsResponse, error) {
	requestQuery, err := utils.TransformProtoToRQL(request.GetQuery(), prospect.Prospect{})
	if err != nil {
		return nil, grpcRQLParseError
	}

	err = rql.ValidateQuery(requestQuery, prospect.Prospect{})
	if err != nil {
		// return nil, err
		return nil, status.Errorf(codes.InvalidArgument, "%v: %v", grpcRQLInvalidError, err)
	}

	prospects, err := h.prospectService.List(ctx, requestQuery)
	if err != nil {
		return nil, grpcInternalServerError
	}

	var transformedProspects []*frontierv1beta1.Prospect
	for _, val := range prospects.Prospects {
		transformedProspect, err := transformProspectToPB(val)
		if err != nil {
			return nil, err
		}
		transformedProspects = append(transformedProspects, transformedProspect)
	}

	var transformedGroups *frontierv1beta1.RQLQueryGroupResponse

	if len(prospects.Group.Data) > 0 {
		groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
		for _, groupItem := range prospects.Group.Data {
			groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
				Name:  groupItem.Name,
				Count: uint32(groupItem.Count),
			})
		}
		transformedGroups = &frontierv1beta1.RQLQueryGroupResponse{
			Name: prospects.Group.Name,
			Data: groupResponse,
		}
	} else {
		transformedGroups = nil
	}

	pagination := &frontierv1beta1.RQLQueryPaginationResponse{
		Offset:     uint32(prospects.Page.Offset),
		Limit:      uint32(prospects.Page.Limit),
		TotalCount: uint32(prospects.Page.TotalCount),
	}

	return &frontierv1beta1.ListProspectsResponse{Prospects: transformedProspects, Group: transformedGroups, Pagination: pagination}, nil
}

func (h Handler) GetProspect(ctx context.Context, request *frontierv1beta1.GetProspectRequest) (*frontierv1beta1.GetProspectResponse, error) {
	prospectId := request.GetId()
	if prospectId == "" {
		return nil, grpcProspectIdRequiredError
	}
	prspct, err := h.prospectService.Get(ctx, prospectId)
	if err != nil {
		switch {
		case errors.Is(err, prospect.ErrNotExist):
			return nil, grpcProspectNotFoundError
		default:
			return nil, grpcInternalServerError
		}
	}
	transformedProspect, err := transformProspectToPB(prspct)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetProspectResponse{Prospect: transformedProspect}, nil
}

func (h Handler) UpdateProspect(ctx context.Context, request *frontierv1beta1.UpdateProspectRequest) (*frontierv1beta1.UpdateProspectResponse, error) {
	prospectId := request.GetId()
	if prospectId == "" {
		return nil, grpcProspectIdRequiredError
	}
	email := request.GetEmail()
	if !isValidEmail(email) {
		return nil, grpcEmailInvalidError
	}
	activity := strings.TrimSpace(request.GetActivity())
	if activity == "" {
		return nil, grpcActivityRequiredError
	}
	reqStatus := request.GetStatus()
	if reqStatus == frontierv1beta1.Prospect_STATUS_UNSPECIFIED {
		return nil, grpcStatusRequiredError
	}
	subsStatus := frontierv1beta1.Prospect_Status_name[int32(reqStatus)] // convert using proto methods
	metaDataMap, err := buildAndValidateMetadata(request.GetMetadata().AsMap(), h)
	if err != nil {
		return nil, err
	}
	updatedProspect, err := h.prospectService.Update(ctx, prospect.Prospect{
		ID:       prospectId,
		Name:     request.GetName(),
		Email:    strings.ToLower(email),
		Verified: request.GetVerified(),
		Phone:    request.GetPhone(),
		Activity: strings.TrimSpace(activity),
		Status:   prospect.StringToStatus(strings.ToLower(subsStatus)),
		Source:   request.GetSource(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, prospect.ErrNotExist):
			return nil, grpcProspectNotFoundError
		case errors.Is(err, prospect.ErrEmailActivityAlreadyExists):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}
	transformedProspect, err := transformProspectToPB(updatedProspect)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.UpdateProspectResponse{Prospect: transformedProspect}, nil
}

func (h Handler) DeleteProspect(ctx context.Context, request *frontierv1beta1.DeleteProspectRequest) (*frontierv1beta1.DeleteProspectResponse, error) {
	prospectId := request.GetId()
	if prospectId == "" {
		return nil, grpcProspectIdRequiredError
	}
	err := h.prospectService.Delete(ctx, prospectId)
	if err != nil {
		switch {
		case errors.Is(err, prospect.ErrNotExist):
			return nil, grpcProspectNotFoundError
		default:
			return nil, grpcInternalServerError
		}
	}
	return &frontierv1beta1.DeleteProspectResponse{}, nil
}

func convertStatusToPBFormat(status prospect.Status) frontierv1beta1.Prospect_Status {
	switch status {
	case prospect.Unsubscribed:
		return frontierv1beta1.Prospect_STATUS_UNSUBSCRIBED
	case prospect.Subscribed:
		return frontierv1beta1.Prospect_STATUS_SUBSCRIBED
	default:
		return frontierv1beta1.Prospect_STATUS_UNSUBSCRIBED
	}
}

func buildAndValidateMetadata(m map[string]any, h Handler) (metadata.Metadata, error) {
	var metaDataMap metadata.Metadata

	metaDataMap = metadata.Build(m)
	if err := h.metaSchemaService.Validate(metaDataMap, prospectMetaSchema); err != nil {
		return nil, grpcBadBodyMetaSchemaError
	}

	return metaDataMap, nil
}

func transformProspectToPB(prospect prospect.Prospect) (*frontierv1beta1.Prospect, error) {
	metaData, err := prospect.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Prospect{}, err
	}
	return &frontierv1beta1.Prospect{
		Id:        prospect.ID,
		Name:      prospect.Name,
		Email:     prospect.Email,
		Phone:     prospect.Phone,
		Activity:  prospect.Activity,
		Status:    convertStatusToPBFormat(prospect.Status),
		ChangedAt: timestamppb.New(prospect.ChangedAt),
		Source:    prospect.Source,
		Verified:  prospect.Verified,
		CreatedAt: timestamppb.New(prospect.CreatedAt),
		UpdatedAt: timestamppb.New(prospect.UpdatedAt),
		Metadata:  metaData,
	}, nil
}
