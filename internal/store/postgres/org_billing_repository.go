package postgres

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	COLUMN_ID                            = "id"
	COLUMN_TITLE                         = "title"
	COLUMN_STATE                         = "state"
	COLUMN_ENDED_AT                      = "ended_at"
	COLUMN_TRIAL_ENDS_AT                 = "trial_ends_at"
	COLUMN_CURRENT_PERIOD_END_AT         = "current_period_end_at"
	COLUMN_CUSTOMER_ID                   = "customer_id"
	COLUMN_PLAN_ID                       = "plan_id"
	COLUMN_ORG_ID                        = "org_id"
	COLUMN_ORG_TITLE                     = "org_title"
	COLUMN_PLAN_TITLE                    = "plan"
	COLUMN_SUBSCRIPTION_STATE            = "subscription_state"
	COLUMN_SUBSCRIPTION_END_TRIGGERED_AT = "subscription_end_triggered_at"
	COLUMN_UPDATED_AT                    = "updated_at"
	COLUMN_ROW_NUM                       = "row_num"
)

type OrgBilling struct {
	dbc *db.Client
}

func NewOrgBillingRepository(dbc *db.Client) *OrgBilling {
	return &OrgBilling{
		dbc: dbc,
	}
}

func (r OrgBilling) Search(ctx context.Context, rql *rql.Query) ([]orgbilling.AggregatedOrganization, error) {
	sql, _, err := prepareSQL()
	if err != nil {
		return nil, err
	}
	fmt.Println(sql)
	return nil, nil
}

func prepareSQL() (string, []interface{}, error) {
	rankedSubscriptions := goqu.From(TABLE_ORGANIZATIONS).
		Select(
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID).As(COLUMN_ORG_ID),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_TITLE).As(COLUMN_ORG_TITLE),
			goqu.I(TABLE_BILLING_PLANS+"."+COLUMN_TITLE).As(COLUMN_PLAN_TITLE),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_STATE).As(COLUMN_SUBSCRIPTION_STATE),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_ENDED_AT).As(COLUMN_SUBSCRIPTION_END_TRIGGERED_AT),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_TRIAL_ENDS_AT),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CURRENT_PERIOD_END_AT),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_UPDATED_AT),
			goqu.Literal("ROW_NUMBER() OVER (PARTITION BY ? ORDER BY ? DESC)", goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID), goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_UPDATED_AT)).As(COLUMN_ROW_NUM),
		).
		LeftJoin(
			goqu.T(TABLE_BILLING_CUSTOMERS),
			goqu.On(goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID).Eq(goqu.I(TABLE_BILLING_CUSTOMERS+"."+COLUMN_ORG_ID))),
		).
		LeftJoin(
			goqu.T(TABLE_BILLING_SUBSCRIPTIONS),
			goqu.On(goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CUSTOMER_ID).Eq(goqu.I(TABLE_BILLING_CUSTOMERS+"."+COLUMN_ID))),
		).
		LeftJoin(
			goqu.T(TABLE_BILLING_PLANS),
			goqu.On(goqu.I(TABLE_BILLING_PLANS+"."+COLUMN_ID).Eq(goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_PLAN_ID))),
		)

	finalQuery := goqu.From(rankedSubscriptions.As("ranked_subscriptions")).
		Select(
			goqu.I(COLUMN_ORG_ID),
			goqu.I(COLUMN_ORG_TITLE),
			goqu.I(COLUMN_PLAN_TITLE),
			goqu.I(COLUMN_SUBSCRIPTION_STATE),
			goqu.I(COLUMN_SUBSCRIPTION_END_TRIGGERED_AT),
			goqu.I(COLUMN_TRIAL_ENDS_AT),
			goqu.I(COLUMN_CURRENT_PERIOD_END_AT),
			goqu.I(COLUMN_UPDATED_AT),
		).
		Where(goqu.I(COLUMN_ROW_NUM).Eq(1))
	return finalQuery.ToSQL()
}
