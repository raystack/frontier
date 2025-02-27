package postgres

import (
	"context"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
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

func (r OrgAggregationRepository) Search(ctx context.Context, id string) (orgbilling.AggregatedOrganization, error) {
	return orgbilling.AggregatedOrganization{}, nil
}
