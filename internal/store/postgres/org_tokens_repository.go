package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgtokens"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	COLUMN_TYPE        = "type"
	COLUMN_DESCRIPTION = "description"
	COLUMN_USER_ID     = "user_id"
	COLUMN_ACCOUNT_ID  = "account_id"
)

type OrgToken struct {
	Amount      sql.NullInt64  `db:"token_amount"`
	Type        sql.NullString `db:"token_type"`
	Description sql.NullString `db:"token_description"`
	UserID      sql.NullString `db:"token_user_id"`
	UserTitle   sql.NullString `db:"user_title"`
	UserAvatar  sql.NullString `db:"user_avatar"`
	CreatedAt   sql.NullTime   `db:"token_created_at"`
	OrgID       sql.NullString `db:"org_id"`
}

func (t *OrgToken) transformToAggregatedToken() svc.AggregatedToken {
	return svc.AggregatedToken{
		Amount:      t.Amount.Int64,
		Type:        t.Type.String,
		Description: t.Description.String,
		UserID:      t.UserID.String,
		UserTitle:   t.UserTitle.String,
		UserAvatar:  t.UserAvatar.String,
		CreatedAt:   t.CreatedAt.Time,
		OrgID:       t.OrgID.String,
	}
}

type OrgTokensRepository struct {
	dbc *db.Client
}

func NewOrgTokensRepository(dbc *db.Client) *OrgTokensRepository {
	return &OrgTokensRepository{
		dbc: dbc,
	}
}

func (r OrgTokensRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrganizationTokens, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	fmt.Println(dataQuery)
	if err != nil {
		return svc.OrganizationTokens{}, err
	}

	var orgTokens []OrgToken

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "GetOrgTokens", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgTokens, dataQuery, params...)
		})
	})

	if err != nil {
		return svc.OrganizationTokens{}, err
	}

	res := make([]svc.AggregatedToken, 0)
	for _, token := range orgTokens {
		res = append(res, token.transformToAggregatedToken())
	}

	return svc.OrganizationTokens{
		Tokens: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r OrgTokensRepository) prepareDataQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	query := r.buildBaseQuery(orgID)

	var err error
	for _, filter := range rql.Filters {
		query, err = r.addFilter(query, filter)
		if err != nil {
			return "", nil, err
		}
	}

	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	query, err = r.addSort(query, rql.Sort)
	if err != nil {
		return "", nil, err
	}

	return query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
}

// we need to cast the user_id to text since it's stored as text in billing_transactions but the users.id is uuid.
func (r OrgTokensRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_BILLING_TRANSACTIONS).Prepared(false).
		Select(
			goqu.I(TABLE_BILLING_TRANSACTIONS+"."+COLUMN_AMOUNT).As("token_amount"),
			goqu.I(TABLE_BILLING_TRANSACTIONS+"."+COLUMN_TYPE).As("token_type"),
			goqu.I(TABLE_BILLING_TRANSACTIONS+"."+COLUMN_DESCRIPTION).As("token_description"),
			goqu.I(TABLE_BILLING_TRANSACTIONS+"."+COLUMN_USER_ID).As("token_user_id"),
			goqu.I(TABLE_USERS+".title").As("user_title"),
			goqu.I(TABLE_USERS+".avatar").As("user_avatar"),
			goqu.I(TABLE_BILLING_TRANSACTIONS+"."+COLUMN_CREATED_AT).As("token_created_at"),
			goqu.I(TABLE_BILLING_CUSTOMERS+"."+COLUMN_ORG_ID).As("org_id"),
		).
		InnerJoin(
			goqu.T(TABLE_BILLING_CUSTOMERS),
			goqu.On(goqu.I(TABLE_BILLING_TRANSACTIONS+"."+COLUMN_ACCOUNT_ID).Eq(goqu.I(TABLE_BILLING_CUSTOMERS+".id"))),
		).
		LeftJoin(
			goqu.T(TABLE_USERS),
			goqu.On(goqu.L(
				`CASE 
					WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' 
					THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id"
					ELSE false 
				END`,
			)),
		).
		Where(goqu.Ex{
			TABLE_BILLING_CUSTOMERS + "." + COLUMN_ORG_ID: orgID,
		})
}

func (r OrgTokensRepository) addFilter(query *goqu.SelectDataset, filter rql.Filter) (*goqu.SelectDataset, error) {
	validFilterFields := map[string]string{
		COLUMN_TITLE:       TABLE_USERS + "." + COLUMN_TITLE,
		COLUMN_DESCRIPTION: TABLE_BILLING_TRANSACTIONS + "." + COLUMN_DESCRIPTION,
		COLUMN_TYPE:        TABLE_BILLING_TRANSACTIONS + "." + COLUMN_TYPE,
		COLUMN_AMOUNT:      TABLE_BILLING_TRANSACTIONS + "." + COLUMN_AMOUNT,
		COLUMN_CREATED_AT:  TABLE_BILLING_TRANSACTIONS + "." + COLUMN_CREATED_AT,
	}

	field, exists := validFilterFields[filter.Name]
	if !exists {
		return nil, fmt.Errorf("unsupported filter field: %s", filter.Name)
	}

	switch filter.Operator {
	case "empty":
		return query.Where(goqu.Or(goqu.I(field).IsNull(), goqu.I(field).Eq(""))), nil
	case "notempty":
		return query.Where(goqu.And(goqu.I(field).IsNotNull(), goqu.I(field).Neq(""))), nil
	case "like", "notlike":
		value := "%" + filter.Value.(string) + "%"
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: value}}), nil
	default:
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: filter.Value}}), nil
	}
}

func (r OrgTokensRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"

	// Define search fields with their table qualifications
	searchSupportedColumns := []string{
		TABLE_BILLING_TRANSACTIONS + "." + COLUMN_TYPE,
		TABLE_BILLING_TRANSACTIONS + "." + COLUMN_DESCRIPTION,
		TABLE_USERS + "." + COLUMN_TITLE,
		TABLE_BILLING_TRANSACTIONS + "." + COLUMN_AMOUNT,
	}

	searchExpressions := make([]goqu.Expression, 0)
	for _, col := range searchSupportedColumns {
		searchExpressions = append(searchExpressions,
			goqu.Cast(goqu.I(col), "TEXT").ILike(searchPattern),
		)
	}

	return query.Where(goqu.Or(searchExpressions...))
}

func (r OrgTokensRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort) (*goqu.SelectDataset, error) {
	// Define sort fields with their table qualifications
	validSortFields := map[string]string{
		"user_title":  TABLE_USERS + "." + COLUMN_TITLE,
		"description": TABLE_BILLING_TRANSACTIONS + "." + COLUMN_DESCRIPTION,
		"type":        TABLE_BILLING_TRANSACTIONS + "." + COLUMN_TYPE,
		"amount":      TABLE_BILLING_TRANSACTIONS + "." + COLUMN_AMOUNT,
		"created_at":  TABLE_BILLING_TRANSACTIONS + "." + COLUMN_CREATED_AT,
	}

	for _, sort := range sorts {
		field, exists := validSortFields[sort.Name]
		if !exists {
			return nil, fmt.Errorf("unsupported sort field: %s", sort.Name)
		}

		switch sort.Order {
		case "asc":
			query = query.OrderAppend(goqu.I(field).Asc())
		case "desc":
			query = query.OrderAppend(goqu.I(field).Desc())
		}
	}

	return query, nil
}
