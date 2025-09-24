package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	orgMetaSchema = "organization"
)

func (h *ConnectHandler) GetOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationRequest]) (*connect.Response[frontierv1beta1.GetOrganizationResponse], error) {
	fetchedOrg, err := h.orgService.GetRaw(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.Is(err, organization.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.GetOrganizationResponse{
		Organization: orgPB,
	}), nil
}

func (h *ConnectHandler) ListOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsResponse], error) {
	var orgs []*frontierv1beta1.Organization
	paginate := pagination.NewPagination(request.Msg.GetPageNum(), request.Msg.GetPageSize())

	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:      organization.State(request.Msg.GetState()),
		UserID:     request.Msg.GetUserId(),
		Pagination: paginate,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		orgs = append(orgs, orgPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationsResponse{
		Organizations: orgs,
	}), nil
}

func (h *ConnectHandler) ListAllOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListAllOrganizationsResponse], error) {
	var orgs []*frontierv1beta1.Organization
	paginate := pagination.NewPagination(request.Msg.GetPageNum(), request.Msg.GetPageSize())

	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:      organization.State(request.Msg.GetState()),
		UserID:     request.Msg.GetUserId(),
		Pagination: paginate,
	})
	if err != nil {
		return nil, err
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, orgPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListAllOrganizationsResponse{
		Organizations: orgs,
		Count:         paginate.Count,
	}), nil
}

func (h *ConnectHandler) CreateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationResponse], error) {
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newOrg, err := h.orgService.Create(ctx, organization.Organization{
		Name:     request.Msg.GetBody().GetName(),
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	})
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) AdminCreateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.AdminCreateOrganizationRequest]) (*connect.Response[frontierv1beta1.AdminCreateOrganizationResponse], error) {
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newOrg, err := h.orgService.AdminCreate(ctx, organization.Organization{
		Name:     request.Msg.GetBody().GetName(),
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	}, request.Msg.GetBody().GetOrgOwnerEmail())
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	})
	return connect.NewResponse(&frontierv1beta1.AdminCreateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) UpdateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateOrganizationRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	var updatedOrg organization.Organization
	var err error
	if utils.IsValidUUID(request.Msg.GetId()) {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			ID:       request.Msg.GetId(),
			Name:     request.Msg.GetBody().GetName(),
			Title:    request.Msg.GetBody().GetTitle(),
			Avatar:   request.Msg.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	} else {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			Name:     request.Msg.GetBody().GetName(),
			Title:    request.Msg.GetBody().GetTitle(),
			Avatar:   request.Msg.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	}
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	orgPB, err := transformOrgToPB(updatedOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, updatedOrg.ID).Log(audit.OrgUpdatedEvent, audit.OrgTarget(updatedOrg.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateOrganizationResponse{Organization: orgPB}), nil
}

func transformOrgToPB(org organization.Organization) (*frontierv1beta1.Organization, error) {
	metaData, err := org.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		Title:     org.Title,
		Metadata:  metaData,
		State:     org.State.String(),
		Avatar:    org.Avatar,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
