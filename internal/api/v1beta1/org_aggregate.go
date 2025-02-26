package v1beta1

import (
	"context"
	orgaggregation "github.com/raystack/frontier/core/aggregates/organization"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Repository interface {
	Search(ctx context.Context, query *rql.Query)
}

type OrgAggregationService interface {
	Search(ctx context.Context, query *rql.Query) ([]orgaggregation.AggregatedOrganization, error)
}

func (h Handler) SearchOrganizations(ctx context.Context, request *frontierv1beta1.SearchOrganizationsRequest) (*frontierv1beta1.SearchOrganizationsResponse, error) {
	var orgs []*frontierv1beta1.SearchOrganizationsResponse_OrganizationResult
	aggregatedOrgList, err := h.orgAggregationService.Search(ctx, transformProtoToRQL(request.Query))
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
	return &rql.Query{
		Search: q.Search,
		Offset: int(q.Offset),
		Limit:  int(q.Limit),
	}
}

func transformAggregatedOrgToPB(v orgaggregation.AggregatedOrganization) *frontierv1beta1.SearchOrganizationsResponse_OrganizationResult {
	return &frontierv1beta1.SearchOrganizationsResponse_OrganizationResult{
		Id:                 v.ID,
		Name:               v.Name,
		Title:              v.Title,
		Avatar:             v.Avatar,
		CreatedAt:          timestamppb.New(v.CreatedAt),
		UpdatedAt:          timestamppb.New(v.UpdatedAt),
		CreatedBy:          v.CreatedBy,
		State:              string(v.State),
		BillingCycleEndsAt: timestamppb.New(v.CycleEndOn),
		BillingPlanName:    v.Plan,
		Country:            v.Country,
		PaymentMode:        v.PaymentMode,
	}
}
