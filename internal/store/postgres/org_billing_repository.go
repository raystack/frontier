package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	COLUMN_ID                      = "id"
	COLUMN_TITLE                   = "title"
	COLUMN_NAME                    = "name"
	COLUMN_STATE                   = "state"
	COLUMN_CREATED_AT              = "created_at"
	COLUMN_POC                     = "poc"
	COLUMN_AVATAR                  = "avatar"
	COLUMN_COUNTRY                 = "country"
	COLUMN_INTERVAL                = "interval"
	COLUMN_TRIAL_ENDS_AT           = "trial_ends_at"
	COLUMN_CURRENT_PERIOD_END_AT   = "current_period_end_at"
	COLUMN_CUSTOMER_ID             = "customer_id"
	COLUMN_PLAN_ID                 = "plan_id"
	COLUMN_ORG_ID                  = "org_id"
	COLUMN_ORG_AVATAR              = "avatar"
	COLUMN_ORG_TITLE               = "org_title"
	COLUMN_ORG_NAME                = "org_name"
	COLUMN_ORG_CREATED_AT          = "org_created_at"
	COLUMN_ORG_STATE               = "org_state"
	COLUMN_ORG_CREATED_BY          = "org_created_by"
	COLUMN_ORG_UPDATED_AT          = "org_updated_at"
	COLUMN_PLAN_NAME               = "plan"
	COLUMN_SUBSCRIPTION_STATE      = "subscription_state"
	COLUMN_UPDATED_AT              = "updated_at"
	COLUMN_ROW_NUM                 = "row_num"
	COLUMN_SUBSCRIPTION_CREATED_AT = "subscription_created_at"
	COLUMN_PLAN_INTERVAL           = "plan_interval"
)

type OrgBillingRepository struct {
	dbc *db.Client
}

type OrgBilling struct {
	OrgID                   string         `db:"org_id"`
	OrgTitle                string         `db:"org_title"`
	OrgName                 string         `db:"org_name"`
	OrgState                string         `db:"org_state"`
	OrgAvatar               string         `db:"avatar"`
	Plan                    sql.NullString `db:"plan"`
	OrgCreatedAt            sql.NullTime   `db:"org_created_at"`
	OrgCreatedBy            sql.NullString `db:"org_created_by"`
	OrgUpdatedAt            sql.NullTime   `db:"org_updated_at"`
	SubscriptionCreatedAt   sql.NullTime   `db:"subscription_created_at"`
	TrialEndsAt             sql.NullTime   `db:"trial_ends_at"`
	SubscriptionPeriodEndAt sql.NullTime   `db:"current_period_end_at"`
	SubscriptionState       sql.NullString `db:"subscription_state"`
	PlanInterval            sql.NullString `db:"plan_interval"`
	Country                 sql.NullString `db:"country"`
	PaymentMode             string         `db:"payment_mode"`
	PlanID                  sql.NullString `db:"plan_id"`
}

func (o *OrgBilling) transformToAggregatedOrganization() orgbilling.AggregatedOrganization {
	return orgbilling.AggregatedOrganization{
		ID:                 o.OrgID,
		Name:               o.OrgName,
		Title:              o.OrgTitle,
		CreatedBy:          o.OrgCreatedBy.String,
		Country:            o.Country.String,
		Avatar:             o.OrgAvatar,
		State:              organization.State(o.OrgState),
		CreatedAt:          o.OrgCreatedAt.Time,
		UpdatedAt:          o.OrgUpdatedAt.Time,
		Plan:               o.Plan.String,
		PlanInterval:       o.PlanInterval.String,
		SubscriptionStatus: o.SubscriptionState.String,
		PaymentMode:        o.PaymentMode,
		CycleEndOn:         o.SubscriptionPeriodEndAt.Time,
		PlanID:             o.PlanID.String,
	}
}

func NewOrgBillingRepository(dbc *db.Client) *OrgBillingRepository {
	return &OrgBillingRepository{
		dbc: dbc,
	}
}

func (r OrgBillingRepository) Search(ctx context.Context, rql *rql.Query) ([]orgbilling.AggregatedOrganization, error) {
	query, params, err := prepareSQL()
	if err != nil {
		return nil, err
	}

	var orgBilling []OrgBilling
	err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetOrgBilling", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &orgBilling, query, params...)
	})
	if err != nil {
		return nil, err
	}

	res := make([]orgbilling.AggregatedOrganization, 0)
	for _, org := range orgBilling {
		res = append(res, org.transformToAggregatedOrganization())
	}
	return res, nil
}

// for each organization, fetch the last created billing_subscription entry
func prepareSQL() (string, []interface{}, error) {
	//prepare a subquery by left joining organizations and billing subscriptions tables
	//and sort by descending order of billing_subscriptions.created_at column
	rankedSubscriptions := goqu.From(TABLE_ORGANIZATIONS).
		Select(
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID).As(COLUMN_ORG_ID),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_TITLE).As(COLUMN_ORG_TITLE),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_NAME).As(COLUMN_ORG_NAME),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_AVATAR).As(COLUMN_AVATAR),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_CREATED_AT).As(COLUMN_ORG_CREATED_AT),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_UPDATED_AT).As(COLUMN_ORG_UPDATED_AT),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_STATE).As(COLUMN_ORG_STATE),
			goqu.L(fmt.Sprintf("%s.metadata->'%s'", TABLE_ORGANIZATIONS, COLUMN_COUNTRY)).As(COLUMN_COUNTRY),
			goqu.L(fmt.Sprintf("%s.metadata->'%s'", TABLE_ORGANIZATIONS, COLUMN_POC)).As(COLUMN_ORG_CREATED_BY),
			goqu.I(TABLE_BILLING_PLANS+"."+COLUMN_ID).As(COLUMN_PLAN_ID),
			goqu.I(TABLE_BILLING_PLANS+"."+COLUMN_NAME).As(COLUMN_PLAN_NAME),
			goqu.I(TABLE_BILLING_PLANS+"."+COLUMN_INTERVAL).As(COLUMN_PLAN_INTERVAL),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_STATE).As(COLUMN_SUBSCRIPTION_STATE),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_TRIAL_ENDS_AT),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CURRENT_PERIOD_END_AT),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CREATED_AT).As(COLUMN_SUBSCRIPTION_CREATED_AT),
			goqu.Literal("ROW_NUMBER() OVER (PARTITION BY ? ORDER BY ? DESC)", goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID), goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CREATED_AT)).As(COLUMN_ROW_NUM),
		).
		LeftJoin(
			goqu.T(TABLE_BILLING_CUSTOMERS),
			goqu.On(goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID).Eq(goqu.I(TABLE_BILLING_CUSTOMERS+"."+COLUMN_ORG_ID))),
		).
		LeftJoin(
			goqu.T(TABLE_BILLING_SUBSCRIPTIONS),
			goqu.On(
				goqu.And(
					goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CUSTOMER_ID).Eq(goqu.I(TABLE_BILLING_CUSTOMERS+"."+COLUMN_ID)),
					goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_STATE).Neq("canceled"),
				),
			)).
		LeftJoin(
			goqu.T(TABLE_BILLING_PLANS),
			goqu.On(goqu.I(TABLE_BILLING_PLANS+"."+COLUMN_ID).Eq(goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_PLAN_ID))),
		)

	// pick the first entry from the above subquery result
	finalQuery := goqu.From(rankedSubscriptions.As("ranked_subscriptions")).
		Select(
			goqu.I(COLUMN_ORG_ID),
			goqu.I(COLUMN_ORG_TITLE),
			goqu.I(COLUMN_ORG_NAME),
			goqu.I(COLUMN_ORG_STATE),
			goqu.I(COLUMN_AVATAR),
			goqu.I(COLUMN_ORG_UPDATED_AT),
			goqu.I(COLUMN_ORG_CREATED_AT),
			goqu.I(COLUMN_ORG_CREATED_BY),
			goqu.I(COLUMN_PLAN_NAME),
			goqu.I(COLUMN_PLAN_ID),
			goqu.I(COLUMN_SUBSCRIPTION_STATE),
			goqu.I(COLUMN_TRIAL_ENDS_AT),
			goqu.I(COLUMN_SUBSCRIPTION_CREATED_AT),
			goqu.I(COLUMN_CURRENT_PERIOD_END_AT),
			goqu.I(COLUMN_PLAN_INTERVAL),
			goqu.I(COLUMN_COUNTRY),
		).
		Where(goqu.I(COLUMN_ROW_NUM).Eq(1))
	return finalQuery.ToSQL()
}
