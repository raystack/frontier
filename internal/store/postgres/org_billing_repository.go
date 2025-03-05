package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/db"
	rqlUtils "github.com/raystack/frontier/pkg/rql"
	"github.com/raystack/salt/rql"
	"golang.org/x/exp/slices"
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
	COLUMN_ORG_STATE               = "org_state"
	COLUMN_CREATED_BY              = "created_by"
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
	OrgID                 string         `db:"id"`
	OrgTitle              string         `db:"title"`
	OrgName               string         `db:"name"`
	OrgState              string         `db:"state"`
	OrgAvatar             string         `db:"avatar"`
	Plan                  sql.NullString `db:"plan"`
	OrgCreatedAt          sql.NullTime   `db:"created_at"`
	OrgCreatedBy          sql.NullString `db:"created_by"`
	OrgUpdatedAt          sql.NullTime   `db:"updated_at"`
	SubscriptionCreatedAt sql.NullTime   `db:"subscription_created_at"`
	TrialEndsAt           sql.NullTime   `db:"trial_ends_at"`
	CycleEndAt            sql.NullTime   `db:"current_period_end_at"`
	SubscriptionState     sql.NullString `db:"subscription_state"`
	PlanInterval          sql.NullString `db:"plan_interval"`
	Country               sql.NullString `db:"country"`
	PaymentMode           string         `db:"payment_mode"`
	PlanID                sql.NullString `db:"plan_id"`
}

func (o *OrgBilling) transformToAggregatedOrganization() orgbilling.AggregatedOrganization {
	return orgbilling.AggregatedOrganization{
		ID:                o.OrgID,
		Name:              o.OrgName,
		Title:             o.OrgTitle,
		CreatedBy:         o.OrgCreatedBy.String,
		Country:           o.Country.String,
		Avatar:            o.OrgAvatar,
		State:             organization.State(o.OrgState),
		CreatedAt:         o.OrgCreatedAt.Time,
		UpdatedAt:         o.OrgUpdatedAt.Time,
		Plan:              o.Plan.String,
		PlanInterval:      o.PlanInterval.String,
		SubscriptionState: o.SubscriptionState.String,
		PaymentMode:       o.PaymentMode,
		CycleEndAt:        o.CycleEndAt.Time,
		PlanID:            o.PlanID.String,
	}
}

func NewOrgBillingRepository(dbc *db.Client) *OrgBillingRepository {
	return &OrgBillingRepository{
		dbc: dbc,
	}
}

func (r OrgBillingRepository) Search(ctx context.Context, rql *rql.Query) ([]orgbilling.AggregatedOrganization, error) {
	query, params, err := prepareSQL(rql)
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
func prepareSQL(rql *rql.Query) (string, []interface{}, error) {
	//prepare a subquery by left joining organizations and billing subscriptions tables
	//and sort by descending order of billing_subscriptions.created_at column

	subquerySelects := []interface{}{
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_ID).As(COLUMN_ID),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_TITLE).As(COLUMN_TITLE),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_NAME).As(COLUMN_NAME),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_AVATAR).As(COLUMN_AVATAR),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_CREATED_AT).As(COLUMN_CREATED_AT),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_UPDATED_AT).As(COLUMN_UPDATED_AT),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_STATE).As(COLUMN_STATE),
		goqu.L(fmt.Sprintf("%s.metadata->'%s'", TABLE_ORGANIZATIONS, COLUMN_COUNTRY)).As(COLUMN_COUNTRY),
		goqu.L(fmt.Sprintf("%s.metadata->'%s'", TABLE_ORGANIZATIONS, COLUMN_POC)).As(COLUMN_CREATED_BY),
		goqu.I(TABLE_BILLING_PLANS + "." + COLUMN_ID).As(COLUMN_PLAN_ID),
		goqu.I(TABLE_BILLING_PLANS + "." + COLUMN_NAME).As(COLUMN_PLAN_NAME),
		goqu.I(TABLE_BILLING_PLANS + "." + COLUMN_INTERVAL).As(COLUMN_PLAN_INTERVAL),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_STATE).As(COLUMN_SUBSCRIPTION_STATE),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_TRIAL_ENDS_AT),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_CURRENT_PERIOD_END_AT),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_CREATED_AT).As(COLUMN_SUBSCRIPTION_CREATED_AT),
		goqu.Literal("ROW_NUMBER() OVER (PARTITION BY ? ORDER BY ? DESC)", goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CREATED_AT)).As(COLUMN_ROW_NUM),
	}

	finalQuerySelects := []interface{}{
		goqu.I(COLUMN_ID),
		goqu.I(COLUMN_TITLE),
		goqu.I(COLUMN_NAME),
		goqu.I(COLUMN_STATE),
		goqu.I(COLUMN_AVATAR),
		goqu.I(COLUMN_UPDATED_AT),
		goqu.I(COLUMN_CREATED_AT),
		goqu.I(COLUMN_CREATED_BY),
		goqu.I(COLUMN_PLAN_NAME),
		goqu.I(COLUMN_PLAN_ID),
		goqu.I(COLUMN_SUBSCRIPTION_STATE),
		goqu.I(COLUMN_TRIAL_ENDS_AT),
		goqu.I(COLUMN_SUBSCRIPTION_CREATED_AT),
		goqu.I(COLUMN_CURRENT_PERIOD_END_AT),
		goqu.I(COLUMN_PLAN_INTERVAL),
		goqu.I(COLUMN_COUNTRY),
	}

	rankedSubscriptions := goqu.From(TABLE_ORGANIZATIONS).
		Select(subquerySelects...).
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
		Select(finalQuerySelects...).Where(goqu.I(COLUMN_ROW_NUM).Eq(1))

	supportedOrgFilters := []string{COLUMN_TITLE, COLUMN_CREATED_AT, COLUMN_STATE, COLUMN_COUNTRY, COLUMN_PLAN_NAME, COLUMN_SUBSCRIPTION_STATE}
	rqlSearchSupportedColumns := []string{COLUMN_TITLE, COLUMN_STATE, COLUMN_PLAN_NAME, COLUMN_PLAN_INTERVAL, COLUMN_SUBSCRIPTION_STATE}

	for _, filter := range rql.Filters {
		if slices.Contains(supportedOrgFilters, filter.Name) {
			datatype, err := rqlUtils.GetDataTypeOfField(filter.Name, orgbilling.AggregatedOrganization{})
			if err != nil {
				return "", nil, err
			}
			switch datatype {
			case "string":
				// empty strings require coalesce function check
				if filter.Value.(string) == "" {
					finalQuery = finalQuery.Where(goqu.L(fmt.Sprintf("coalesce(%s, '') = ''", filter.Name)))
				} else {
					finalQuery = finalQuery.Where(goqu.Ex{
						filter.Name: goqu.Op{filter.Operator: filter.Value.(string)},
					})
				}

			case "number":
				finalQuery = finalQuery.Where(goqu.Ex{
					filter.Name: goqu.Op{filter.Operator: filter.Value.(float32)},
				})
			case "bool":
				finalQuery = finalQuery.Where(goqu.Ex{
					filter.Name: goqu.Op{filter.Operator: filter.Value.(bool)},
				})
			case "datetime":
				finalQuery = finalQuery.Where(goqu.Ex{
					filter.Name: goqu.Op{filter.Operator: filter.Value.(string)},
				})
			}
		}
	}

	searchExpressions := make([]goqu.Expression, 0)
	if rql.Search != "" {
		for _, col := range rqlSearchSupportedColumns {
			searchExpressions = append(searchExpressions, goqu.Ex{
				col: goqu.Op{"LIKE": "%" + rql.Search + "%"},
			})
		}
	}

	finalQuery = finalQuery.Where(goqu.Or(searchExpressions...))

	for _, sortItem := range rql.Sort {
		switch sortItem.Order {
		case "asc":
			finalQuery = finalQuery.OrderAppend(goqu.C(sortItem.Name).Asc())
		case "desc":
			finalQuery = finalQuery.OrderAppend(goqu.C(sortItem.Name).Desc())
		default:
		}
	}
	finalQuery = finalQuery.Offset(uint(rql.Offset))
	finalQuery = finalQuery.Limit(uint(rql.Limit))
	//finalQuery = finalQuery.Where(goqu.I(COLUMN_ORG_ID).Eq("045c1b0d-fd38-4f3b-9aee-8f0adac08b33"))
	return finalQuery.ToSQL()
}
