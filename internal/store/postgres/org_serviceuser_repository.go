package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

type ServiceUserRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	OrgID       string       `db:"org_id"`
	ProjectData string       `db:"project_data"`
	CreatedAt   sql.NullTime `db:"created_at"`
}

func (c *ServiceUserRow) transformToAggregatedServiceUser(orgID string) svc.AggregatedServiceUser {
	var projects []svc.Project
	if c.ProjectData != "" && c.ProjectData != "null" {
		// Parse JSON array of project objects
		err := json.Unmarshal([]byte(c.ProjectData), &projects)
		if err != nil {
			// If JSON parsing fails, return empty projects array
			projects = []svc.Project{}
		}
	}

	return svc.AggregatedServiceUser{
		ID:        c.ID,
		OrgID:     orgID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt.Time,
		Projects:  projects,
	}
}

type OrgServiceUserRepository struct {
	dbc *db.Client
}

func NewOrgServiceUserRepository(dbc *db.Client) *OrgServiceUserRepository {
	return &OrgServiceUserRepository{
		dbc: dbc,
	}
}

func (r OrgServiceUserRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrganizationServiceUsers, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	if err != nil {
		return svc.OrganizationServiceUsers{}, err
	}

	var serviceUsers []ServiceUserRow

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_SERVICE_USERS, "SearchOrgServiceUsers", func(ctx context.Context) error {
			return r.dbc.SelectContext(ctx, &serviceUsers, dataQuery, params...)
		})
	})

	if err != nil {
		return svc.OrganizationServiceUsers{}, err
	}

	res := make([]svc.AggregatedServiceUser, 0)
	for _, serviceuser := range serviceUsers {
		transformed := serviceuser.transformToAggregatedServiceUser(orgID)
		res = append(res, transformed)
	}

	return svc.OrganizationServiceUsers{
		ServiceUsers: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r OrgServiceUserRepository) prepareDataQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	var query *goqu.SelectDataset

	query = r.buildBaseQuery(orgID)

	if rql != nil {
		for _, filter := range rql.Filters {
			query = r.addFilter(query, filter)
		}
		if rql.Search != "" {
			query = r.addSearch(query, rql.Search)
		}
		if len(rql.Sort) > 0 {
			var err error
			query, err = r.addSort(query, rql.Sort)
			if err != nil {
				return "", nil, err
			}
		}
		if rql.Limit > 0 {
			query = query.Limit(uint(rql.Limit))
		}
		if rql.Offset > 0 {
			query = query.Offset(uint(rql.Offset))
		}
	}

	return query.ToSQL()
}

func (r OrgServiceUserRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_SERVICE_USERS).Prepared(true).
		Select(
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ID).As("id"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_TITLE).As("title"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ORG_ID).As("org_id"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_CREATED_AT).As("created_at"),
			goqu.L("JSON_AGG(JSON_BUILD_OBJECT('id', "+TABLE_PROJECTS+"."+COLUMN_ID+", 'title', "+TABLE_PROJECTS+"."+COLUMN_TITLE+", 'name', "+TABLE_PROJECTS+"."+COLUMN_NAME+"))").As("project_data"),
		).
		InnerJoin(
			goqu.T(TABLE_POLICIES),
			goqu.On(
				goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID)),
				goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_TYPE).Eq(schema.ServiceUserPrincipal),
				goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq(schema.ProjectNamespace),
			),
		).
		InnerJoin(
			goqu.T(TABLE_PROJECTS),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(goqu.I(TABLE_PROJECTS+"."+COLUMN_ID))),
		).
		Where(goqu.Ex{
			TABLE_SERVICE_USERS + "." + COLUMN_ORG_ID: orgID,
		}).
		GroupBy(
			TABLE_SERVICE_USERS + "." + COLUMN_ID,
		).
		Order(goqu.I(TABLE_SERVICE_USERS + "." + COLUMN_TITLE).Asc())
}

func (r OrgServiceUserRepository) addFilter(query *goqu.SelectDataset, filter rql.Filter) *goqu.SelectDataset {
	var field string
	// Map field names to their table-qualified names
	switch filter.Name {
	case "title":
		field = TABLE_SERVICE_USERS + "." + COLUMN_TITLE
	case "created_at":
		field = TABLE_SERVICE_USERS + "." + COLUMN_CREATED_AT
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

func (r OrgServiceUserRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"
	searchExpressions := make([]goqu.Expression, 0)

	searchExpressions = append(searchExpressions,
		goqu.Cast(goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_TITLE), "TEXT").ILike(searchPattern),
	)

	return query.Where(goqu.Or(searchExpressions...))
}

func (r OrgServiceUserRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort) (*goqu.SelectDataset, error) {
	validSortFields := map[string]string{
		"title":      TABLE_SERVICE_USERS + "." + COLUMN_TITLE,
		"created_at": TABLE_SERVICE_USERS + "." + COLUMN_CREATED_AT,
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
