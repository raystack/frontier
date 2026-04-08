package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgpats"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

// Operator constants for RQL filters.
const (
	opEmpty    = "empty"
	opNotEmpty = "notempty"
	opLike     = "like"
	opNotLike  = "notlike"
	opILike    = "ilike"
	opNotILike = "notilike"
)

// orgPATFilterFields maps RQL filter names to table-qualified SQL columns.
var orgPATFilterFields = map[string]string{
	"id":               "p.id",
	"title":            "p.title",
	"created_by_id":    "p.user_id",
	"created_by_title": "u.title",
	"created_by_email": "u.email",
	"created_at":       "p.created_at",
	"expires_at":       "p.expires_at",
	"last_used_at":     "p.last_used_at",
}

// orgPATSearchColumns are searched with ILIKE when RQL search is used.
var orgPATSearchColumns = []string{
	"p.id",
	"p.title",
	"u.title",
	"u.email",
}

// orgPATSortFields maps RQL sort names to table-qualified SQL columns.
var orgPATSortFields = map[string]string{
	"title":            "p.title",
	"created_by_title": "u.title",
	"created_by_email": "u.email",
	"created_at":       "p.created_at",
	"expires_at":       "p.expires_at",
	"last_used_at":     "p.last_used_at",
}

// OrgPATRow is the flat SQL result row from the joined query.
// Multiple rows per PAT (one per policy).
type OrgPATRow struct {
	PATID          string     `db:"pat_id"`
	PATTitle       string     `db:"pat_title"`
	CreatedByID    string     `db:"created_by_id"`
	CreatedByTitle string     `db:"created_by_title"`
	CreatedByEmail string     `db:"created_by_email"`
	CreatedAt      time.Time  `db:"pat_created_at"`
	ExpiresAt      time.Time  `db:"pat_expires_at"`
	LastUsedAt     *time.Time `db:"pat_last_used_at"`
	RoleID         *string    `db:"role_id"`
	ResourceType   *string    `db:"resource_type"`
	ResourceID     *string    `db:"resource_id"`
	GrantRelation  *string    `db:"grant_relation"`
}

type OrgPATsRepository struct {
	dbc *db.Client
}

func NewOrgPATsRepository(dbc *db.Client) *OrgPATsRepository {
	return &OrgPATsRepository{dbc: dbc}
}

func (r OrgPATsRepository) Search(ctx context.Context, orgID string, rqlQuery *rql.Query) (svc.OrganizationPATs, error) {
	if rqlQuery == nil {
		rqlQuery = utils.NewRQLQuery("", utils.DefaultOffset, utils.DefaultLimit, []rql.Filter{}, []rql.Sort{}, []string{})
	}

	countSQL, countParams, err := r.buildCountQuery(orgID, rqlQuery)
	if err != nil {
		return svc.OrganizationPATs{}, fmt.Errorf("building count query: %w", err)
	}

	dataSQL, dataParams, err := r.buildDataQuery(orgID, rqlQuery)
	if err != nil {
		return svc.OrganizationPATs{}, fmt.Errorf("building data query: %w", err)
	}

	var totalCount int64
	var rows []OrgPATRow

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		if err := r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "SearchOrgPATsCount", func(ctx context.Context) error {
			return tx.GetContext(ctx, &totalCount, countSQL, countParams...)
		}); err != nil {
			return err
		}
		return r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "SearchOrgPATsData", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &rows, dataSQL, dataParams...)
		})
	})
	if err != nil {
		return svc.OrganizationPATs{}, fmt.Errorf("querying org PATs: %w", err)
	}

	return svc.OrganizationPATs{
		PATs: r.groupRows(rows),
		Pagination: utils.Page{
			Offset:     rqlQuery.Offset,
			Limit:      rqlQuery.Limit,
			TotalCount: totalCount,
		},
	}, nil
}

// buildInnerSubquery builds the base PAT query with filters and search applied.
func (r OrgPATsRepository) buildInnerSubquery(orgID string, rqlQuery *rql.Query) (*goqu.SelectDataset, error) {
	inner := dialect.From(goqu.T(TABLE_USER_PATS).As("p")).
		LeftJoin(
			goqu.T(TABLE_USERS).As("u"),
			goqu.On(goqu.I("p.user_id").Eq(goqu.I("u.id"))),
		).
		Where(
			goqu.I("p.org_id").Eq(orgID),
			goqu.I("p.deleted_at").IsNull(),
		)

	var err error
	for _, filter := range rqlQuery.Filters {
		inner, err = r.addFilter(inner, filter)
		if err != nil {
			return nil, err
		}
	}

	if rqlQuery.Search != "" {
		inner = r.addSearch(inner, rqlQuery.Search)
	}

	return inner, nil
}

func (r OrgPATsRepository) buildCountQuery(orgID string, rqlQuery *rql.Query) (string, []interface{}, error) {
	inner, err := r.buildInnerSubquery(orgID, rqlQuery)
	if err != nil {
		return "", nil, err
	}
	return inner.Select(goqu.L("COUNT(*)")).Prepared(true).ToSQL()
}

func (r OrgPATsRepository) buildDataQuery(orgID string, rqlQuery *rql.Query) (string, []interface{}, error) {
	inner, err := r.buildInnerSubquery(orgID, rqlQuery)
	if err != nil {
		return "", nil, err
	}

	inner, err = r.addSort(inner, rqlQuery.Sort)
	if err != nil {
		return "", nil, err
	}

	paginatedInner := inner.
		Select(
			goqu.I("p.id"), goqu.I("p.title"), goqu.I("p.user_id"),
			goqu.I("p.created_at"), goqu.I("p.expires_at"), goqu.I("p.last_used_at"),
		).
		Offset(uint(rqlQuery.Offset)).
		Limit(uint(rqlQuery.Limit))

	outer := dialect.From(paginatedInner.As("p")).Prepared(true).
		Select(
			goqu.I("p.id").As("pat_id"),
			goqu.I("p.title").As("pat_title"),
			goqu.I("p.user_id").As("created_by_id"),
			goqu.L("COALESCE(u.title, '')").As("created_by_title"),
			goqu.L("u.email").As("created_by_email"),
			goqu.I("p.created_at").As("pat_created_at"),
			goqu.I("p.expires_at").As("pat_expires_at"),
			goqu.I("p.last_used_at").As("pat_last_used_at"),
			goqu.I("pol.role_id"),
			goqu.I("pol.resource_type"),
			goqu.I("pol.resource_id"),
			goqu.I("pol.grant_relation"),
		).
		LeftJoin(
			goqu.T(TABLE_USERS).As("u"),
			goqu.On(goqu.I("p.user_id").Eq(goqu.I("u.id"))),
		).
		LeftJoin(
			goqu.T(TABLE_POLICIES).As("pol"),
			goqu.On(
				goqu.I("pol.principal_id").Eq(goqu.I("p.id")),
				goqu.I("pol.principal_type").Eq(schema.PATPrincipal),
			),
		).
		Order(
			goqu.I("p.created_at").Desc(),
			goqu.I("p.id").Asc(),
			goqu.I("pol.role_id").Asc(),
		)

	return outer.ToSQL()
}

func (r OrgPATsRepository) addFilter(query *goqu.SelectDataset, filter rql.Filter) (*goqu.SelectDataset, error) {
	field, exists := orgPATFilterFields[filter.Name]
	if !exists {
		return nil, fmt.Errorf("unsupported filter field: %s", filter.Name)
	}

	switch filter.Operator {
	case opEmpty:
		return query.Where(goqu.Or(goqu.I(field).IsNull(), goqu.I(field).Eq(""))), nil
	case opNotEmpty:
		return query.Where(goqu.And(goqu.I(field).IsNotNull(), goqu.I(field).Neq(""))), nil
	case opLike:
		return query.Where(goqu.Cast(goqu.I(field), "TEXT").Like(filter.Value.(string))), nil
	case opNotLike:
		return query.Where(goqu.Cast(goqu.I(field), "TEXT").NotLike(filter.Value.(string))), nil
	case opILike:
		return query.Where(goqu.Cast(goqu.I(field), "TEXT").ILike(filter.Value.(string))), nil
	case opNotILike:
		return query.Where(goqu.Cast(goqu.I(field), "TEXT").NotILike(filter.Value.(string))), nil
	default:
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: filter.Value}}), nil
	}
}

func (r OrgPATsRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"
	expressions := make([]goqu.Expression, 0, len(orgPATSearchColumns))
	for _, col := range orgPATSearchColumns {
		expressions = append(expressions, goqu.Cast(goqu.I(col), "TEXT").ILike(searchPattern))
	}
	return query.Where(goqu.Or(expressions...))
}

func (r OrgPATsRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort) (*goqu.SelectDataset, error) {
	for _, sort := range sorts {
		field, exists := orgPATSortFields[sort.Name]
		if !exists {
			return nil, fmt.Errorf("unsupported sort field: %s", sort.Name)
		}
		switch sort.Order {
		case "desc":
			query = query.OrderAppend(goqu.I(field).Desc())
		default:
			query = query.OrderAppend(goqu.I(field).Asc())
		}
	}

	if len(sorts) == 0 {
		query = query.OrderAppend(goqu.I("p.created_at").Desc())
	}

	return query, nil
}

// groupRows groups flat SQL rows by PAT ID into AggregatedPAT with scopes.
func (r OrgPATsRepository) groupRows(rows []OrgPATRow) []svc.AggregatedPAT {
	type scopeKey struct {
		roleID       string
		resourceType string
	}

	patOrder := make([]string, 0)
	patMap := make(map[string]*svc.AggregatedPAT)
	scopeMap := make(map[string]map[scopeKey]*patmodels.PATScope)
	allProjects := make(map[string]map[scopeKey]bool)

	for _, row := range rows {
		if row.PATID == "" {
			continue
		}

		if _, exists := patMap[row.PATID]; !exists {
			pat := &svc.AggregatedPAT{
				ID:    row.PATID,
				Title: row.PATTitle,
				CreatedBy: svc.CreatedBy{
					ID:    row.CreatedByID,
					Title: row.CreatedByTitle,
					Email: row.CreatedByEmail,
				},
				CreatedAt:  row.CreatedAt,
				ExpiresAt:  row.ExpiresAt,
				LastUsedAt: row.LastUsedAt,
				UserID:     row.CreatedByID,
			}
			patMap[row.PATID] = pat
			patOrder = append(patOrder, row.PATID)
			scopeMap[row.PATID] = make(map[scopeKey]*patmodels.PATScope)
			allProjects[row.PATID] = make(map[scopeKey]bool)
		}

		// Build scopes from policy rows
		if row.RoleID == nil || *row.RoleID == "" {
			continue
		}

		var key scopeKey
		var isAllProjects bool

		switch {
		case row.ResourceType != nil && *row.ResourceType == schema.ProjectNamespace:
			key = scopeKey{*row.RoleID, schema.ProjectNamespace}
		case row.GrantRelation != nil && *row.GrantRelation == schema.PATGrantRelationName:
			key = scopeKey{*row.RoleID, schema.ProjectNamespace}
			isAllProjects = true
		case row.ResourceType != nil && *row.ResourceType == schema.OrganizationNamespace:
			key = scopeKey{*row.RoleID, schema.OrganizationNamespace}
		default:
			continue
		}

		sc, ok := scopeMap[row.PATID][key]
		if !ok {
			sc = &patmodels.PATScope{
				RoleID:       key.roleID,
				ResourceType: key.resourceType,
			}
			scopeMap[row.PATID][key] = sc
		}

		if isAllProjects {
			allProjects[row.PATID][key] = true
			sc.ResourceIDs = nil
		} else if !allProjects[row.PATID][key] && row.ResourceID != nil {
			sc.ResourceIDs = append(sc.ResourceIDs, *row.ResourceID)
		}
	}

	result := make([]svc.AggregatedPAT, 0, len(patOrder))
	for _, patID := range patOrder {
		pat := patMap[patID]
		scopes := make([]patmodels.PATScope, 0, len(scopeMap[patID]))
		for _, sc := range scopeMap[patID] {
			scopes = append(scopes, *sc)
		}
		pat.Scopes = scopes
		result = append(result, *pat)
	}

	return result
}
