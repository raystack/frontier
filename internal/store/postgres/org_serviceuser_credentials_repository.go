package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgserviceusercredentials"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	TABLE_SERVICE_USERS            = "serviceusers"
	TABLE_SERVICE_USER_CREDENTIALS = "serviceuser_credentials"

	COLUMN_SERVICEUSER_ID = "serviceuser_id"
)

type ServiceUserCredentialRow struct {
	Title            sql.NullString `db:"credential_title"`
	ServiceUserTitle sql.NullString `db:"serviceuser_title"`
	CreatedAt        sql.NullTime   `db:"credential_created_at"`
	OrgID            sql.NullString `db:"org_id"`
}

func (c *ServiceUserCredentialRow) transformToAggregatedServiceUserCredential() svc.AggregatedServiceUserCredential {
	return svc.AggregatedServiceUserCredential{
		Title:            c.Title.String,
		ServiceUserTitle: c.ServiceUserTitle.String,
		CreatedAt:        c.CreatedAt.Time,
		OrgID:            c.OrgID.String,
	}
}

type OrgServiceUserCredentialsRepository struct {
	dbc *db.Client
}

func NewOrgServiceUserCredentialsRepository(dbc *db.Client) *OrgServiceUserCredentialsRepository {
	return &OrgServiceUserCredentialsRepository{
		dbc: dbc,
	}
}

func (r OrgServiceUserCredentialsRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrganizationServiceUserCredentials, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	if err != nil {
		return svc.OrganizationServiceUserCredentials{}, err
	}

	var credentials []ServiceUserCredentialRow

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_SERVICE_USER_CREDENTIALS, "GetOrgServiceUserCredentials", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &credentials, dataQuery, params...)
		})
	})

	if err != nil {
		return svc.OrganizationServiceUserCredentials{}, err
	}

	res := make([]svc.AggregatedServiceUserCredential, 0)
	for _, cred := range credentials {
		res = append(res, cred.transformToAggregatedServiceUserCredential())
	}

	return svc.OrganizationServiceUserCredentials{
		Credentials: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r OrgServiceUserCredentialsRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_SERVICE_USER_CREDENTIALS).Prepared(true).
		Select(
			goqu.I(TABLE_SERVICE_USER_CREDENTIALS+"."+COLUMN_TITLE).As("credential_title"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_TITLE).As("serviceuser_title"),
			goqu.I(TABLE_SERVICE_USER_CREDENTIALS+"."+COLUMN_CREATED_AT).As("credential_created_at"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ORG_ID).As("org_id"),
		).
		InnerJoin(
			goqu.T(TABLE_SERVICE_USERS),
			goqu.On(goqu.I(TABLE_SERVICE_USER_CREDENTIALS+"."+COLUMN_SERVICEUSER_ID).Eq(goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ID))),
		).
		Where(goqu.Ex{
			TABLE_SERVICE_USERS + "." + COLUMN_ORG_ID: orgID,
		})
}

func (r OrgServiceUserCredentialsRepository) prepareDataQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	query := r.buildBaseQuery(orgID)

	// Apply filters
	for _, filter := range rql.Filters {
		query = r.addFilter(query, filter)
	}

	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	// Add sorting
	query, err := r.addSort(query, rql.Sort)
	if err != nil {
		return "", nil, err
	}

	return query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
}

func (r OrgServiceUserCredentialsRepository) addFilter(query *goqu.SelectDataset, filter rql.Filter) *goqu.SelectDataset {
	var field string
	// Map field names to their table-qualified names
	switch filter.Name {
	case "title":
		field = TABLE_SERVICE_USER_CREDENTIALS + "." + COLUMN_TITLE
	case "serviceuser_title":
		field = TABLE_SERVICE_USERS + "." + COLUMN_TITLE
	case "created_at":
		field = TABLE_SERVICE_USER_CREDENTIALS + "." + COLUMN_CREATED_AT
	default:
		return query
	}

	switch filter.Operator {
	case "empty":
		return query.Where(goqu.Or(goqu.I(field).IsNull(), goqu.I(field).Eq("")))
	case "notempty":
		return query.Where(goqu.And(goqu.I(field).IsNotNull(), goqu.I(field).Neq("")))
	case "like", "notlike":
		value := "%" + filter.Value.(string) + "%"
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: value}})
	default:
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: filter.Value}})
	}
}

func (r OrgServiceUserCredentialsRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"

	searchExpressions := make([]goqu.Expression, 0)
	// Search across credential title and service user title
	searchExpressions = append(searchExpressions,
		goqu.Cast(goqu.I(TABLE_SERVICE_USER_CREDENTIALS+"."+COLUMN_TITLE), "TEXT").ILike(searchPattern),
		goqu.Cast(goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_TITLE), "TEXT").ILike(searchPattern),
	)

	return query.Where(goqu.Or(searchExpressions...))
}

func (r OrgServiceUserCredentialsRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort) (*goqu.SelectDataset, error) {
	validSortFields := map[string]string{
		"title":             TABLE_SERVICE_USER_CREDENTIALS + "." + COLUMN_TITLE,
		"serviceuser_title": TABLE_SERVICE_USERS + "." + COLUMN_TITLE,
		"created_at":        TABLE_SERVICE_USER_CREDENTIALS + "." + COLUMN_CREATED_AT,
	}

	for _, sort := range sorts {
		field, exists := validSortFields[sort.Name]
		if !exists {
			return nil, fmt.Errorf("invalid sort field: %s", sort.Name)
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
