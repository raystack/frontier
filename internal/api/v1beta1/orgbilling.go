package v1beta1

import (
	"context"
	"fmt"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgAggregationService interface {
	Search(ctx context.Context, query *rql.Query) (orgbilling.OrgBilling, error)
}

func (h Handler) SearchOrganizations(ctx context.Context, request *frontierv1beta1.SearchOrganizationsRequest) (*frontierv1beta1.SearchOrganizationsResponse, error) {
	var orgs []*frontierv1beta1.SearchOrganizationsResponse_OrganizationResult

	rqlQuery, err := transformProtoToRQL(request.Query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgbilling.AggregatedOrganization{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	orgBillingData, err := h.orgAggregationService.Search(ctx, rqlQuery)
	if err != nil {
		return nil, err
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
	return &frontierv1beta1.SearchOrganizationsResponse{
		Organizations: orgs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgBillingData.Pagination.Offset),
			Limit:  uint32(orgBillingData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgBillingData.Group.Name,
			Data: groupResponse,
		},
	}, nil
}

func transformProtoToRQL(q *frontierv1beta1.RQLRequest) (*rql.Query, error) {
	filters := make([]rql.Filter, 0)
	for _, filter := range q.Filters {
		datatype, err := rql.GetDataTypeOfField(filter.Name, orgbilling.AggregatedOrganization{})
		if err != nil {
			return nil, err
		}
		switch datatype {
		case "string":
			filters = append(filters, rql.Filter{
				Name:     filter.Name,
				Operator: filter.Operator,
				Value:    filter.GetStringValue(),
			})
		case "number":
			filters = append(filters, rql.Filter{
				Name:     filter.Name,
				Operator: filter.Operator,
				Value:    filter.GetNumberValue(),
			})
		case "bool":
			filters = append(filters, rql.Filter{
				Name:     filter.Name,
				Operator: filter.Operator,
				Value:    filter.GetBoolValue(),
			})
		case "datetime":
			filters = append(filters, rql.Filter{
				Name:     filter.Name,
				Operator: filter.Operator,
				Value:    filter.GetStringValue(),
			})
		}

	}

	sortItems := make([]rql.Sort, 0)
	for _, sortItem := range q.Sort {
		sortItems = append(sortItems, rql.Sort{Name: sortItem.Name, Order: sortItem.Order})
	}

	return &rql.Query{
		Search:  q.Search,
		Offset:  int(q.Offset),
		Limit:   int(q.Limit),
		Filters: filters,
		Sort:    sortItems,
		GroupBy: q.GroupBy,
	}, nil
}

func transformAggregatedOrgToPB(v orgbilling.AggregatedOrganization) *frontierv1beta1.SearchOrganizationsResponse_OrganizationResult {
	return &frontierv1beta1.SearchOrganizationsResponse_OrganizationResult{
		Id:                v.ID,
		Name:              v.Name,
		Title:             v.Title,
		Avatar:            v.Avatar,
		CreatedAt:         timestamppb.New(v.CreatedAt),
		UpdatedAt:         timestamppb.New(v.UpdatedAt),
		CreatedBy:         v.CreatedBy,
		State:             string(v.State),
		Country:           v.Country,
		PaymentMode:       v.PaymentMode,
		Plan:              v.Plan,
		PlanId:            v.PlanID,
		SubscriptionState: v.SubscriptionState,
		PlanInterval:      v.PlanInterval,
		CycleEndAt:        timestamppb.New(v.CycleEndAt),
	}
}
