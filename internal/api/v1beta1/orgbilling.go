package v1beta1

import (
	"context"
	"fmt"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	rqlUtils "github.com/raystack/frontier/pkg/rql"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const TAG = "rql"

type OrgAggregationService interface {
	Search(ctx context.Context, query *rql.Query) ([]orgbilling.AggregatedOrganization, error)
}

func (h Handler) SearchOrganizations(ctx context.Context, request *frontierv1beta1.SearchOrganizationsRequest) (*frontierv1beta1.SearchOrganizationsResponse, error) {
	var orgs []*frontierv1beta1.SearchOrganizationsResponse_OrganizationResult
	//TODO: validated request with rql struct tag defined in domain struct
	rqlQuery := transformProtoToRQL(request.Query)
	aggregatedOrgList, err := h.orgAggregationService.Search(ctx, rqlQuery)
	if err != nil {
		return nil, err
	}

	for _, v := range aggregatedOrgList {
		orgs = append(orgs, transformAggregatedOrgToPB(v))
	}

	return &frontierv1beta1.SearchOrganizationsResponse{
		Organizations: orgs,
	}, nil
}

func transformProtoToRQL(q *frontierv1beta1.RQLRequest) *rql.Query {
	filters := make([]rql.Filter, 0)
	for _, filter := range q.Filters {
		datatype, err := rqlUtils.GetDataTypeOfField(filter.Name, orgbilling.AggregatedOrganization{})
		if err != nil {
			//TODO: handle this error
			// also validate using rql
			fmt.Println(err.Error())
			return nil
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
	}
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
		PlanName:          v.Plan,
		PlanId:            v.PlanID,
		SubscriptionState: v.SubscriptionState,
		PlanInterval:      v.PlanInterval,
		CycleEndAt:        timestamppb.New(v.CycleEndAt),
	}
}
