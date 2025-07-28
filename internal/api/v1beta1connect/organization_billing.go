package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgBillingService interface {
	Search(ctx context.Context, query *rql.Query) (orgbilling.OrgBilling, error)
	Export(ctx context.Context) ([]byte, string, error)
}

func (h *ConnectHandler) SearchOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationsResponse], error) {
	var orgs []*frontierv1beta1.SearchOrganizationsResponse_OrganizationResult

	rqlQuery, err := transformProtoToRQL(request.Msg.GetQuery())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgbilling.AggregatedOrganization{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	orgBillingData, err := h.orgBillingService.Search(ctx, rqlQuery)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	for _, v := range orgBillingData.Organizations {
		orgs = append(orgs, transformAggregatedOrgToPB(v))
	}

	groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
	for _, groupItem := range orgBillingData.Group.Data {
		groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
			Name:  groupItem.Name,
			Count: uint32(groupItem.Count),
		})
	}
	return connect.NewResponse(&frontierv1beta1.SearchOrganizationsResponse{
		Organizations: orgs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgBillingData.Pagination.Offset),
			Limit:  uint32(orgBillingData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgBillingData.Group.Name,
			Data: groupResponse,
		},
	}), nil
}

func (h *ConnectHandler) ExportOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ExportOrganizationsRequest], stream *connect.ServerStream[httpbody.HttpBody]) error {
	orgBillingDataBytes, contentType, err := h.orgBillingService.Export(ctx)
	if err != nil {
		return nil
	}

	chunkSize := 100 * 200 // 200KB

	for i := 0; i < len(orgBillingDataBytes); i += chunkSize {
		end := min(i+chunkSize, len(orgBillingDataBytes))

		chunk := orgBillingDataBytes[i:end]
		msg := &httpbody.HttpBody{
			ContentType: contentType,
			Data:        chunk,
		}

		err := stream.Send(msg)
		if err != nil {
			return connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return nil
}

func transformProtoToRQL(q *frontierv1beta1.RQLRequest) (*rql.Query, error) {
	filters := make([]rql.Filter, 0)
	for _, filter := range q.GetFilters() {
		datatype, err := rql.GetDataTypeOfField(filter.GetName(), orgbilling.AggregatedOrganization{})
		if err != nil {
			return nil, err
		}
		switch datatype {
		case "string":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetStringValue(),
			})
		case "number":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetNumberValue(),
			})
		case "bool":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetBoolValue(),
			})
		case "datetime":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetStringValue(),
			})
		}
	}

	sortItems := make([]rql.Sort, 0)
	for _, sortItem := range q.GetSort() {
		sortItems = append(sortItems, rql.Sort{Name: sortItem.GetName(), Order: sortItem.GetOrder()})
	}

	return &rql.Query{
		Search:  q.GetSearch(),
		Offset:  int(q.GetOffset()),
		Limit:   int(q.GetLimit()),
		Filters: filters,
		Sort:    sortItems,
		GroupBy: q.GetGroupBy(),
	}, nil
}

func transformAggregatedOrgToPB(v orgbilling.AggregatedOrganization) *frontierv1beta1.SearchOrganizationsResponse_OrganizationResult {
	return &frontierv1beta1.SearchOrganizationsResponse_OrganizationResult{
		Id:                     v.ID,
		Name:                   v.Name,
		Title:                  v.Title,
		Avatar:                 v.Avatar,
		CreatedAt:              timestamppb.New(v.CreatedAt),
		UpdatedAt:              timestamppb.New(v.UpdatedAt),
		CreatedBy:              v.CreatedBy,
		State:                  string(v.State),
		Country:                v.Country,
		PaymentMode:            v.PaymentMode,
		PlanName:               v.PlanName,
		PlanId:                 v.PlanID,
		SubscriptionState:      v.SubscriptionState,
		PlanInterval:           v.PlanInterval,
		SubscriptionCycleEndAt: timestamppb.New(v.SubscriptionCycleEndAt),
	}
}
