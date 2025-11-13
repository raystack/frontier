package v1beta1connect

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateProspectPublic(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProspectPublicRequest]) (*connect.Response[frontierv1beta1.CreateProspectPublicResponse], error) {
	errorLogger := NewErrorLogger()
	email := request.Msg.GetEmail()

	if !isValidEmail(email) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail)
	}
	activity := strings.TrimSpace(request.Msg.GetActivity())
	if activity == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrActivityRequired)
	}
	metaDataMap, err := buildAndValidateMetadata(request.Msg.GetMetadata().AsMap(), h)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	_, err = h.prospectService.Create(ctx, prospect.Prospect{
		Name:     request.Msg.GetName(),
		Email:    strings.ToLower(email),
		Phone:    request.Msg.GetPhone(),
		Activity: strings.TrimSpace(activity),
		Status:   prospect.Subscribed, // Subscribed by default
		Verified: false,
		Source:   request.Msg.GetSource(),
		Metadata: metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateProspectPublic", err,
			zap.String("email", strings.ToLower(email)),
			zap.String("activity", activity),
			zap.String("source", request.Msg.GetSource()))

		switch {
		case errors.Is(err, prospect.ErrEmailActivityAlreadyExists):
			return connect.NewResponse(&frontierv1beta1.CreateProspectPublicResponse{}), nil
		default:
			errorLogger.LogUnexpectedError(ctx, request, "CreateProspectPublic", err,
				zap.String("email", strings.ToLower(email)),
				zap.String("activity", activity),
				zap.String("source", request.Msg.GetSource()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return connect.NewResponse(&frontierv1beta1.CreateProspectPublicResponse{}), nil
}

func (h *ConnectHandler) CreateProspect(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProspectRequest]) (*connect.Response[frontierv1beta1.CreateProspectResponse], error) {
	errorLogger := NewErrorLogger()
	email := request.Msg.GetEmail()

	if !isValidEmail(email) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail)
	}
	activity := strings.TrimSpace(request.Msg.GetActivity())
	if activity == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrActivityRequired)
	}
	reqStatus := request.Msg.GetStatus()
	if reqStatus == frontierv1beta1.Prospect_STATUS_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrStatusRequired)
	}
	subsStatus := frontierv1beta1.Prospect_Status_name[int32(reqStatus)] // convert using proto methods
	metaDataMap, err := buildAndValidateMetadata(request.Msg.GetMetadata().AsMap(), h)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newProspect, err := h.prospectService.Create(ctx, prospect.Prospect{
		Name:     request.Msg.GetName(),
		Email:    strings.ToLower(email),
		Phone:    request.Msg.GetPhone(),
		Activity: strings.TrimSpace(activity),
		Status:   prospect.StringToStatus(strings.ToLower(subsStatus)),
		Verified: request.Msg.GetVerified(),
		Source:   request.Msg.GetSource(),
		Metadata: metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateProspect", err,
			zap.String("email", strings.ToLower(email)),
			zap.String("activity", activity),
			zap.String("status", subsStatus),
			zap.String("source", request.Msg.GetSource()))

		switch {
		case errors.Is(err, prospect.ErrEmailActivityAlreadyExists):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "CreateProspect", err,
				zap.String("email", strings.ToLower(email)),
				zap.String("activity", activity),
				zap.String("status", subsStatus),
				zap.String("source", request.Msg.GetSource()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	transformedProspect, err := transformProspectToPB(newProspect)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "CreateProspect", newProspect.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.CreateProspectResponse{Prospect: transformedProspect}), nil
}

func (h *ConnectHandler) ListProspects(ctx context.Context, request *connect.Request[frontierv1beta1.ListProspectsRequest]) (*connect.Response[frontierv1beta1.ListProspectsResponse], error) {
	errorLogger := NewErrorLogger()

	requestQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), prospect.Prospect{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrRQLParse)
	}

	err = rql.ValidateQuery(requestQuery, prospect.Prospect{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("datatype or operator not supported in RQL query: %s", err.Error()))
	}

	prospects, err := h.prospectService.List(ctx, requestQuery)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProspects", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var transformedProspects []*frontierv1beta1.Prospect
	for _, val := range prospects.Prospects {
		transformedProspect, err := transformProspectToPB(val)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProspects", val.ID, err)
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInternalServerError)
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

	return connect.NewResponse(&frontierv1beta1.ListProspectsResponse{Prospects: transformedProspects, Group: transformedGroups, Pagination: pagination}), nil
}

func (h *ConnectHandler) GetProspect(ctx context.Context, request *connect.Request[frontierv1beta1.GetProspectRequest]) (*connect.Response[frontierv1beta1.GetProspectResponse], error) {
	errorLogger := NewErrorLogger()
	prospectId := request.Msg.GetId()

	if prospectId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrProspectIdRequired)
	}
	prspct, err := h.prospectService.Get(ctx, prospectId)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetProspect", err,
			zap.String("prospect_id", prospectId))

		switch {
		case errors.Is(err, prospect.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrProspectNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "GetProspect", err,
				zap.String("prospect_id", prospectId))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	transformedProspect, err := transformProspectToPB(prspct)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetProspect", prspct.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GetProspectResponse{Prospect: transformedProspect}), nil
}

func (h *ConnectHandler) UpdateProspect(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProspectRequest]) (*connect.Response[frontierv1beta1.UpdateProspectResponse], error) {
	errorLogger := NewErrorLogger()
	prospectId := request.Msg.GetId()

	if prospectId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrProspectIdRequired)
	}
	email := request.Msg.GetEmail()
	if !isValidEmail(email) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail)
	}
	activity := strings.TrimSpace(request.Msg.GetActivity())
	if activity == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrActivityRequired)
	}
	reqStatus := request.Msg.GetStatus()
	if reqStatus == frontierv1beta1.Prospect_STATUS_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrStatusRequired)
	}
	subsStatus := frontierv1beta1.Prospect_Status_name[int32(reqStatus)] // convert using proto methods
	metaDataMap, err := buildAndValidateMetadata(request.Msg.GetMetadata().AsMap(), h)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	updatedProspect, err := h.prospectService.Update(ctx, prospect.Prospect{
		ID:       prospectId,
		Name:     request.Msg.GetName(),
		Email:    strings.ToLower(email),
		Verified: request.Msg.GetVerified(),
		Phone:    request.Msg.GetPhone(),
		Activity: strings.TrimSpace(activity),
		Status:   prospect.StringToStatus(strings.ToLower(subsStatus)),
		Source:   request.Msg.GetSource(),
		Metadata: metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdateProspect", err,
			zap.String("prospect_id", prospectId),
			zap.String("email", strings.ToLower(email)),
			zap.String("activity", activity),
			zap.String("status", subsStatus),
			zap.String("source", request.Msg.GetSource()))

		switch {
		case errors.Is(err, prospect.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrProspectNotFound)
		case errors.Is(err, prospect.ErrEmailActivityAlreadyExists):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrConflictRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "UpdateProspect", err,
				zap.String("prospect_id", prospectId),
				zap.String("email", strings.ToLower(email)),
				zap.String("activity", activity),
				zap.String("status", subsStatus),
				zap.String("source", request.Msg.GetSource()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	transformedProspect, err := transformProspectToPB(updatedProspect)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "UpdateProspect", updatedProspect.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.UpdateProspectResponse{Prospect: transformedProspect}), nil
}

func (h *ConnectHandler) DeleteProspect(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteProspectRequest]) (*connect.Response[frontierv1beta1.DeleteProspectResponse], error) {
	errorLogger := NewErrorLogger()
	prospectId := request.Msg.GetId()

	if prospectId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrProspectIdRequired)
	}
	err := h.prospectService.Delete(ctx, prospectId)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteProspect", err,
			zap.String("prospect_id", prospectId))

		switch {
		case errors.Is(err, prospect.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrProspectNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "DeleteProspect", err,
				zap.String("prospect_id", prospectId))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return connect.NewResponse(&frontierv1beta1.DeleteProspectResponse{}), nil
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

func buildAndValidateMetadata(m map[string]any, h *ConnectHandler) (metadata.Metadata, error) {
	var metaDataMap metadata.Metadata

	metaDataMap = metadata.Build(m)
	if err := h.metaSchemaService.Validate(metaDataMap, prospectMetaSchema); err != nil {
		return nil, ErrBadBodyMetaSchemaError
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
