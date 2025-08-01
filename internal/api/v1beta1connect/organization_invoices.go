package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orginvoices"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgInvoicesService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orginvoices.OrganizationInvoices, error)
}

func (h *ConnectHandler) SearchOrganizationInvoices(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationInvoicesRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationInvoicesResponse], error) {
	var orgInvoices []*frontierv1beta1.SearchOrganizationInvoicesResponse_OrganizationInvoice

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), orginvoices.AggregatedInvoice{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orginvoices.AggregatedInvoice{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	invoicesData, err := h.orgInvoicesService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInternalServerError)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range invoicesData.Invoices {
		orgInvoices = append(orgInvoices, transformOrganizationInvoiceToPB(v))
	}

	var groupResponse *frontierv1beta1.RQLQueryGroupResponse
	if invoicesData.Group.Name != "" {
		groupData := make([]*frontierv1beta1.RQLQueryGroupData, 0)
		for _, d := range invoicesData.Group.Data {
			groupData = append(groupData, &frontierv1beta1.RQLQueryGroupData{
				Name:  d.Name,
				Count: uint32(d.Count),
			})
		}
		groupResponse = &frontierv1beta1.RQLQueryGroupResponse{
			Name: invoicesData.Group.Name,
			Data: groupData,
		}
	}

	return connect.NewResponse(&frontierv1beta1.SearchOrganizationInvoicesResponse{
		OrganizationInvoices: orgInvoices,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(invoicesData.Pagination.Offset),
			Limit:  uint32(invoicesData.Pagination.Limit),
		},
		Group: groupResponse,
	}), nil
}

func transformOrganizationInvoiceToPB(v orginvoices.AggregatedInvoice) *frontierv1beta1.SearchOrganizationInvoicesResponse_OrganizationInvoice {
	return &frontierv1beta1.SearchOrganizationInvoicesResponse_OrganizationInvoice{
		Id:          v.ID,
		Amount:      v.Amount,
		Currency:    v.Currency,
		State:       v.State,
		InvoiceLink: v.InvoiceLink,
		CreatedAt:   timestamppb.New(v.CreatedAt),
		OrgId:       v.OrgID,
	}
}
