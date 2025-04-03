package postgres

import (
	"context"
	"database/sql"

	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	svc "github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	PRINCIPAL_TYPE_USER = "app/user"
	COLUMN_AVATARS      = "avatars"
)

type ProjectMemberCount struct {
	ResourceID  string `db:"resource_id"`
	ProjectName string `db:"project_name"`
	UserCount   int    `db:"user_count"`
}

type OrgProjectsRepository struct {
	dbc *db.Client
}

type OrgProjects struct {
	ProjectID      sql.NullString `db:"id"`
	ProjectName    sql.NullString `db:"name"`
	ProjectTitle   sql.NullString `db:"title"`
	ProjectState   sql.NullString `db:"state"`
	MemberCount    sql.NullInt64  `db:"member_count"`
	UserIDs        pq.StringArray `db:"user_ids"`
	CreatedAt      sql.NullTime   `db:"created_at"`
	OrganizationID sql.NullString `db:"org_id"`
}

type OrgProjectsGroup struct {
	Name sql.NullString         `db:"name"`
	Data []OrgProjectsGroupData `db:"data"`
}

type OrgProjectsGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (p *OrgProjects) transformToAggregatedProject() svc.AggregatedProject {
	return svc.AggregatedProject{
		ID:             p.ProjectID.String,
		Name:           p.ProjectName.String,
		Title:          p.ProjectTitle.String,
		State:          project.State(p.ProjectState.String),
		MemberCount:    p.MemberCount.Int64,
		UserIDs:        p.UserIDs,
		CreatedAt:      p.CreatedAt.Time,
		OrganizationID: p.OrganizationID.String,
	}
}

func NewOrgProjectsRepository(dbc *db.Client) *OrgProjectsRepository {
	return &OrgProjectsRepository{
		dbc: dbc,
	}
}

func (r OrgProjectsRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrgProjects, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	if err != nil {
		return svc.OrgProjects{}, err
	}

	var orgProjects []OrgProjects

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "CountProjectMembers", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgProjects, dataQuery, params...)
		})

		if err != nil {
			return err
		}

		return err
	})

	if err != nil {
		return svc.OrgProjects{}, err
	}

	res := make([]svc.AggregatedProject, 0)
	for _, project := range orgProjects {
		res = append(res, project.transformToAggregatedProject())
	}
	return svc.OrgProjects{
		Projects: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r OrgProjectsRepository) prepareDataQuery(orgID string, rqlQuery *rql.Query) (string, []interface{}, error) {
	baseQ := r.baseQuery(orgID)

	baseQWithFilters, err := r.applyFilters(rqlQuery, baseQ)
	if err != nil {
		return "", nil, err
	}

	baseQWithFiltersAndSearch := r.applySearch(rqlQuery, baseQWithFilters)

	baseQWithFiltersAndSearchAndGroupBy := r.addGroupBy(baseQWithFiltersAndSearch)

	baseQWithFiltersAndSearchAndGroupAndSort, err := r.aplyRQLSort(rqlQuery, baseQWithFiltersAndSearchAndGroupBy)
	if err != nil {
		return "", nil, err
	}

	return baseQWithFiltersAndSearchAndGroupAndSort.Offset(uint(rqlQuery.Offset)).Limit(uint(rqlQuery.Limit)).ToSQL()
}

func (r OrgProjectsRepository) baseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_POLICIES).Prepared(true).
		Select(
			goqu.I(TABLE_PROJECTS+"."+COLUMN_ID),
			goqu.I(TABLE_PROJECTS+"."+COLUMN_NAME),
			goqu.I(TABLE_PROJECTS+"."+COLUMN_TITLE),
			goqu.I(TABLE_PROJECTS+"."+COLUMN_STATE),
			goqu.I(TABLE_PROJECTS+"."+COLUMN_CREATED_AT),
			goqu.I(TABLE_PROJECTS+"."+COLUMN_ORG_ID),
			goqu.COUNT(goqu.DISTINCT(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID)).As("member_count"),
			goqu.L("array_agg(DISTINCT "+TABLE_USERS+"."+COLUMN_ID+")").As("user_ids"),
		).
		InnerJoin(
			goqu.T(TABLE_PROJECTS),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(goqu.I(TABLE_PROJECTS+"."+COLUMN_ID))),
		).
		InnerJoin(
			goqu.T(TABLE_USERS),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID))),
		).
		Where(goqu.Ex{
			TABLE_PROJECTS + "." + COLUMN_ORG_ID: orgID,
			COLUMN_PRINCIPAL_TYPE:                PRINCIPAL_TYPE_USER,
		})
}

func (r OrgProjectsRepository) applySearch(rqlQuery *rql.Query, stmt *goqu.SelectDataset) *goqu.SelectDataset {
	searchableFields := []string{COLUMN_TITLE, COLUMN_NAME, COLUMN_STATE}
	if rqlQuery.Search != "" {
		searchConditions := []goqu.Expression{}
		for _, field := range searchableFields {
			searchConditions = append(searchConditions,
				goqu.I(TABLE_PROJECTS+"."+field).ILike("%"+rqlQuery.Search+"%"))
		}
		stmt = stmt.Where(goqu.Or(searchConditions...))
	}
	return stmt
}

func (r OrgProjectsRepository) applyFilters(rqlQuery *rql.Query, stmt *goqu.SelectDataset) (*goqu.SelectDataset, error) {
	for _, filter := range rqlQuery.Filters {
		field := TABLE_PROJECTS + "." + filter.Name

		switch filter.Name {
		case COLUMN_TITLE, COLUMN_NAME, COLUMN_ID, COLUMN_STATE:
			stmt = r.applyStringFilter(filter, field, stmt)
		case COLUMN_CREATED_AT:
			stmt = r.applyDatetimeFilter(filter, field, stmt)

		default:
			return nil, fmt.Errorf("%w: filtering not supported for field: %s", ErrBadInput, filter.Name)
		}
	}
	return stmt, nil
}

func (r OrgProjectsRepository) applyStringFilter(filter rql.Filter, field string, stmt *goqu.SelectDataset) *goqu.SelectDataset {
	var condition goqu.Expression

	switch filter.Operator {
	case OPERATOR_EMPTY:
		condition = goqu.L(fmt.Sprintf("coalesce(%s, '') = ''", field))
	case OPERATOR_NOT_EMPTY:
		condition = goqu.L(fmt.Sprintf("coalesce(%s, '') != ''", field))
	case OPERATOR_IN, OPERATOR_NOT_IN:
		condition = goqu.Ex{field: goqu.Op{filter.Operator: strings.Split(filter.Value.(string), ",")}}
	case OPERATOR_LIKE, OPERATOR_NOT_LIKE:
		searchPattern := "%" + filter.Value.(string) + "%"
		if filter.Operator == OPERATOR_LIKE {
			condition = goqu.Cast(goqu.I(field), "TEXT").ILike(searchPattern)
		} else {
			condition = goqu.Cast(goqu.I(field), "TEXT").NotILike(searchPattern)
		}
	default: // eq, neq
		condition = goqu.Ex{field: goqu.Op{filter.Operator: filter.Value}}
	}

	return stmt.Where(condition)
}

func (r OrgProjectsRepository) applyDatetimeFilter(filter rql.Filter, field string, stmt *goqu.SelectDataset) *goqu.SelectDataset {
	condition := goqu.Ex{
		field: goqu.Op{
			filter.Operator: goqu.L(fmt.Sprintf("timestamp '%s'", filter.Value)),
		},
	}
	return stmt.Where(condition)
}

func (r OrgProjectsRepository) addGroupBy(stmt *goqu.SelectDataset) *goqu.SelectDataset {
	return stmt.GroupBy(
		TABLE_PROJECTS+"."+COLUMN_ID,
		TABLE_PROJECTS+"."+COLUMN_NAME,
		TABLE_PROJECTS+"."+COLUMN_TITLE,
		TABLE_PROJECTS+"."+COLUMN_STATE,
		TABLE_PROJECTS+"."+COLUMN_CREATED_AT,
		TABLE_PROJECTS+"."+COLUMN_ORG_ID,
	)
}

func (r OrgProjectsRepository) aplyRQLSort(rql *rql.Query, query *goqu.SelectDataset) (*goqu.SelectDataset, error) {
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
