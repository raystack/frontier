package postgres

import (
	"context"
	orgaggregation "github.com/raystack/frontier/core/aggregates/organization"
	"github.com/raystack/frontier/pkg/db"
)

type OrgAggregationRepository struct {
	dbc *db.Client
}

func NewOrgAggregationRepository(dbc *db.Client) *OrgAggregationRepository {
	return &OrgAggregationRepository{
		dbc: dbc,
	}
}

func (r OrgAggregationRepository) Search(ctx context.Context, id string) (orgaggregation.AggregatedOrganization, error) {
	return orgaggregation.AggregatedOrganization{}, nil
}
