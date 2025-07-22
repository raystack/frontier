package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/pagination"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
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
