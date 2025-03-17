package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"golang.org/x/exp/slices"
)

const (
	OPERATOR_EMPTY     = "empty"
	OPERATOR_NOT_EMPTY = "notempty"
	OPERATOR_IN        = "in"
	OPERATOR_NOT_IN    = "notin"
	OPERATOR_LIKE      = "like"
	OPERATOR_NOT_LIKE  = "notlike"
)

const (
	COLUMN_ID                        = "id"
	COLUMN_TITLE                     = "title"
	COLUMN_NAME                      = "name"
	COLUMN_STATE                     = "state"
	COLUMN_CREATED_AT                = "created_at"
	COLUMN_POC                       = "poc"
	COLUMN_AVATAR                    = "avatar"
	COLUMN_COUNTRY                   = "country"
	COLUMN_INTERVAL                  = "interval"
	COLUMN_TRIAL_ENDS_AT             = "trial_ends_at"
	COLUMN_CURRENT_PERIOD_END_AT     = "current_period_end_at"
	COLUMN_SUBSCRIPTION_CYCLE_END_AT = "subscription_cycle_end_at"
	COLUMN_CUSTOMER_ID               = "customer_id"
	COLUMN_PLAN_ID                   = "plan_id"
	COLUMN_ORG_ID                    = "org_id"
	COLUMN_CREATED_BY                = "created_by"
	COLUMN_PLAN_NAME                 = "plan_name"
	COLUMN_SUBSCRIPTION_STATE        = "subscription_state"
	COLUMN_UPDATED_AT                = "updated_at"
	COLUMN_ROW_NUM                   = "row_num"
	COLUMN_PLAN_INTERVAL             = "plan_interval"
	COLUMN_COUNT                     = "count"
	COLUMN_VALUES                    = "values"
)

type OrgBillingRepository struct {
	dbc *db.Client
}

type OrgBilling struct {
	OrgID                  string         `db:"id"`
	OrgTitle               string         `db:"title"`
	OrgName                string         `db:"name"`
	OrgState               string         `db:"state"`
	OrgAvatar              string         `db:"avatar"`
	PlanName               sql.NullString `db:"plan_name"`
	OrgCreatedAt           sql.NullTime   `db:"created_at"`
	OrgCreatedBy           sql.NullString `db:"created_by"`
	OrgUpdatedAt           sql.NullTime   `db:"updated_at"`
	SubscriptionCycleEndAt sql.NullTime   `db:"subscription_cycle_end_at"`
	SubscriptionState      sql.NullString `db:"subscription_state"`
	PlanInterval           sql.NullString `db:"plan_interval"`
	Country                sql.NullString `db:"country"`
	PaymentMode            string         `db:"payment_mode"`
	PlanID                 sql.NullString `db:"plan_id"`
}

type OrgBillingGroup struct {
	Name sql.NullString        `db:"name"`
	Data []OrgBillingGroupData `db:"data"`
}

type OrgBillingGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (o *OrgBilling) transformToAggregatedOrganization() svc.AggregatedOrganization {
	return svc.AggregatedOrganization{
		ID:                     o.OrgID,
		Name:                   o.OrgName,
		Title:                  o.OrgTitle,
		CreatedBy:              o.OrgCreatedBy.String,
		Country:                o.Country.String,
		Avatar:                 o.OrgAvatar,
		State:                  organization.State(o.OrgState),
		CreatedAt:              o.OrgCreatedAt.Time,
		UpdatedAt:              o.OrgUpdatedAt.Time,
		PlanName:               o.PlanName.String,
		PlanInterval:           o.PlanInterval.String,
		SubscriptionState:      o.SubscriptionState.String,
		PaymentMode:            o.PaymentMode,
		SubscriptionCycleEndAt: o.SubscriptionCycleEndAt.Time,
		PlanID:                 o.PlanID.String,
	}
}

func (o *OrgBillingGroup) transformToOrgBillingGroup() svc.Group {
	orgBillingGroupData := make([]svc.GroupData, 0)
	for _, groupDataItem := range o.Data {
		orgBillingGroupData = append(orgBillingGroupData, svc.GroupData{
			Name:  groupDataItem.Name.String,
			Count: groupDataItem.Count,
		})
	}
	return svc.Group{
		Name: o.Name.String,
		Data: orgBillingGroupData,
	}
}

func NewOrgBillingRepository(dbc *db.Client) *OrgBillingRepository {
	return &OrgBillingRepository{
		dbc: dbc,
	}
}

func (r OrgBillingRepository) Search(ctx context.Context, rql *rql.Query) (svc.OrgBilling, error) {
	dataQuery, params, err := prepareDataQuery(rql)
	dataQuery = sqlx.Rebind(sqlx.DOLLAR, dataQuery)
	if err != nil {
		return svc.OrgBilling{}, err
	}

	var orgBilling []OrgBilling
	var orgBillingGroupData []OrgBillingGroupData
	var orgBillingGroup OrgBillingGroup

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetOrgBilling", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgBilling, dataQuery, params...)
		})

		if err != nil {
			return err
		}

		if len(rql.GroupBy) > 0 {
			groupByQuery, groupByParams, err := prepareGroupByQuery(rql)
			if err != nil {
				return err
			}

			err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetOrgBillingWithGroup", func(ctx context.Context) error {
				return tx.SelectContext(ctx, &orgBillingGroupData, groupByQuery, groupByParams...)
			})

			if err != nil {
				return err
			}
			orgBillingGroup.Name = sql.NullString{String: rql.GroupBy[0]}
			orgBillingGroup.Data = orgBillingGroupData
		}
		return err
	})

	if err != nil {
		return svc.OrgBilling{}, err
	}

	res := make([]svc.AggregatedOrganization, 0)
	for _, org := range orgBilling {
		res = append(res, org.transformToAggregatedOrganization())
	}
	return svc.OrgBilling{
		Organizations: res,
		Group:         orgBillingGroup.transformToOrgBillingGroup(),
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

// for each organization, fetch the last created billing_subscription entry
func prepareDataQuery(rql *rql.Query) (string, []interface{}, error) {
	dataQuerySelects := []interface{}{
		goqu.I(COLUMN_ID),
		goqu.I(COLUMN_TITLE),
		goqu.I(COLUMN_NAME),
		goqu.I(COLUMN_STATE),
		goqu.I(COLUMN_AVATAR),
		goqu.I(COLUMN_UPDATED_AT),
		goqu.I(COLUMN_CREATED_AT),
		goqu.I(COLUMN_CREATED_BY),
		goqu.I(COLUMN_COUNTRY),
		goqu.I(COLUMN_PLAN_ID),
		goqu.I(COLUMN_PLAN_NAME),
		goqu.I(COLUMN_SUBSCRIPTION_STATE),
		goqu.I(COLUMN_SUBSCRIPTION_CYCLE_END_AT),
		goqu.I(COLUMN_PLAN_INTERVAL),
	}

	rankedSubscriptions := getSubQuery()

	// pick the first entry from the above subquery result
	baseQ := goqu.From(rankedSubscriptions.As("ranked_subscriptions")).Prepared(true).
		Select(dataQuerySelects...).Where(goqu.I(COLUMN_ROW_NUM).Eq(1))

	withFilterQ, err := addRQLFiltersInQuery(baseQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLFiltersInQuery: %w", err)
	}

	withFilterAndSearchQ, err := addRQLSearchInQuery(withFilterQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLSearchInQuery: %w", err)
	}

	withSortAndFilterAndSearchQ, err := addRQLSortInQuery(withFilterAndSearchQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLSortInQuery: %w", err)
	}

	return withSortAndFilterAndSearchQ.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).Prepared(true).ToSQL()
}

// for each organization, fetch the last created billing_subscription entry grouped by first key in rql.GroupBy list
// RQL supports multiple group_by key, but for simplicity of implementation
// and view of Frontier Admin Console only one group_by key is being supported
func prepareGroupByQuery(rql *rql.Query) (string, []interface{}, error) {
	validGroupByKeys := []string{
		COLUMN_STATE,
		COLUMN_PLAN_NAME,
		COLUMN_SUBSCRIPTION_STATE,
		COLUMN_PLAN_INTERVAL,
	}

	if len(rql.GroupBy) == 0 {
		return "", nil, fmt.Errorf("rql group_by is empty list")
	}

	groupByKey := rql.GroupBy[0]

	if !slices.Contains(validGroupByKeys, groupByKey) {
		return "", nil, fmt.Errorf("invalid group_by key %s", groupByKey)
	}

	finalQuerySelects := []interface{}{
		goqu.COUNT("*").As(COLUMN_COUNT),
		goqu.I(rql.GroupBy[0]).As(COLUMN_VALUES),
	}

	rankedSubscriptions := getSubQuery()

	// pick the first entry from the above subquery result
	baseQ := goqu.From(rankedSubscriptions.As("ranked_subscriptions")).
		Select(finalQuerySelects...).Where(goqu.I(COLUMN_ROW_NUM).Eq(1))

	withFiltersQ, err := addRQLFiltersInQuery(baseQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLFiltersInQuery: %w", err)
	}

	withSearchAndFilterQ, err := addRQLSearchInQuery(withFiltersQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLSearhInQuery: %w", err)
	}

	finalQuery := withSearchAndFilterQ.GroupBy(groupByKey)
	return finalQuery.Prepared(true).ToSQL()
}

// prepare a subquery by left joining organizations and billing subscriptions tables
// and sort by descending order of billing_subscriptions.created_at column
func getSubQuery() *goqu.SelectDataset {
	subquerySelects := []interface{}{
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_ID).As(COLUMN_ID),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_TITLE).As(COLUMN_TITLE),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_NAME).As(COLUMN_NAME),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_AVATAR).As(COLUMN_AVATAR),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_CREATED_AT).As(COLUMN_CREATED_AT),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_UPDATED_AT).As(COLUMN_UPDATED_AT),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_STATE).As(COLUMN_STATE),
		goqu.L(fmt.Sprintf("%s.metadata->>'%s'", TABLE_ORGANIZATIONS, COLUMN_COUNTRY)).As(COLUMN_COUNTRY),
		goqu.L(fmt.Sprintf("%s.metadata->>'%s'", TABLE_ORGANIZATIONS, COLUMN_POC)).As(COLUMN_CREATED_BY),
		goqu.I(TABLE_BILLING_PLANS + "." + COLUMN_ID).As(COLUMN_PLAN_ID),
		goqu.I(TABLE_BILLING_PLANS + "." + COLUMN_NAME).As(COLUMN_PLAN_NAME),
		goqu.I(TABLE_BILLING_PLANS + "." + COLUMN_INTERVAL).As(COLUMN_PLAN_INTERVAL),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_STATE).As(COLUMN_SUBSCRIPTION_STATE),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_TRIAL_ENDS_AT),
		goqu.I(TABLE_BILLING_SUBSCRIPTIONS + "." + COLUMN_CURRENT_PERIOD_END_AT).As(COLUMN_SUBSCRIPTION_CYCLE_END_AT),
		goqu.Literal("ROW_NUMBER() OVER (PARTITION BY ? ORDER BY ? DESC)", goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID),
			goqu.I(TABLE_BILLING_SUBSCRIPTIONS+"."+COLUMN_CREATED_AT)).As(COLUMN_ROW_NUM),
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
	return rankedSubscriptions
}

func addRQLFiltersInQuery(query *goqu.SelectDataset, rqlInput *rql.Query) (*goqu.SelectDataset, error) {
	supportedFilters := []string{
		COLUMN_ID,
		COLUMN_TITLE,
		COLUMN_STATE,
		COLUMN_CREATED_AT,
		COLUMN_PLAN_NAME,
		COLUMN_SUBSCRIPTION_STATE,
		COLUMN_SUBSCRIPTION_CYCLE_END_AT,
		COLUMN_PLAN_INTERVAL,
	}

	for _, filter := range rqlInput.Filters {
		if !slices.Contains(supportedFilters, filter.Name) {
			return nil, fmt.Errorf("%s is not supported in filters", filter.Name)
		}
		datatype, err := rql.GetDataTypeOfField(filter.Name, svc.AggregatedOrganization{})
		if err != nil {
			return query, err
		}
		switch datatype {
		case "string":
			query = processStringDataType(filter, query)
		case "number":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(float32)},
			})
		case "bool":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(bool)},
			})
		case "datetime":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(string)},
			})
		}
	}
	return query, nil
}

func addRQLSearchInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, error) {
	// this should contain only those columns that are sql string(text, varchar etc) datatype
	rqlSearchSupportedColumns := []string{
		COLUMN_ID,
		COLUMN_TITLE,
		COLUMN_STATE,
		COLUMN_PLAN_NAME,
		COLUMN_SUBSCRIPTION_STATE,
		COLUMN_PLAN_INTERVAL,
	}

	searchExpressions := make([]goqu.Expression, 0)
	if rql.Search != "" {
		for _, col := range rqlSearchSupportedColumns {
			searchExpressions = append(searchExpressions, goqu.L(
				fmt.Sprintf(`"%s"::TEXT ILIKE '%%%s%%'`, col, rql.Search),
			))
		}
	}
	return query.Where(goqu.Or(searchExpressions...)), nil
}

func addRQLSortInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, error) {
	// If there is a group by parameter added then sort the result
	// by group_by first key in asc order by default before any other sort column
	if len(rql.GroupBy) > 0 {
		query = query.OrderAppend(goqu.C(rql.GroupBy[0]).Asc())
	}

	for _, sortItem := range rql.Sort {
		switch sortItem.Order {
		case "asc":
			query = query.OrderAppend(goqu.C(sortItem.Name).Asc())
		case "desc":
			query = query.OrderAppend(goqu.C(sortItem.Name).Desc())
		default:
		}
	}
	return query, nil
}

func processStringDataType(filter rql.Filter, query *goqu.SelectDataset) *goqu.SelectDataset {
	switch filter.Operator {
	case OPERATOR_EMPTY:
		query = query.Where(goqu.L(fmt.Sprintf("coalesce(%s, '') = ''", filter.Name)))
	case OPERATOR_NOT_EMPTY:
		query = query.Where(goqu.L(fmt.Sprintf("coalesce(%s, '') != ''", filter.Name)))
	case OPERATOR_IN, OPERATOR_NOT_IN:
		// process the values of in and notin operators as comma separated list
		query = query.Where(goqu.Ex{
			filter.Name: goqu.Op{filter.Operator: strings.Split(filter.Value.(string), ",")},
		})
	case OPERATOR_LIKE:
		// some semi string sql types like UUID require casting to text to support like operator
		query = query.Where(goqu.L(fmt.Sprintf(`"%s"::TEXT LIKE '%s'`, filter.Name, filter.Value.(string))))
	case OPERATOR_NOT_LIKE:
		// some semi string sql types like UUID require casting to text to support like operator
		query = query.Where(goqu.L(fmt.Sprintf(`"%s"::TEXT NOT LIKE '%s'`, filter.Name, filter.Value.(string))))
	default:
		query = query.Where(goqu.Ex{filter.Name: goqu.Op{filter.Operator: filter.Value.(string)}})
	}
	return query
}
