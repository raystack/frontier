package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	svc "github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	// Organizations Table Columns
	COLUMN_ORG_NAME   = "org_name"
	COLUMN_ORG_TITLE  = "org_title"
	COLUMN_ORG_AVATAR = "org_avatar"

	// Custom Aggregated Columns
	COLUMN_PROJECT_COUNT = "project_count"
	COLUMN_ORG_JOINED_ON = "org_joined_on"
)

type UserOrgsRepository struct {
	dbc *db.Client
}

type UserOrgs struct {
	OrgID        sql.NullString `db:"org_id"`
	OrgTitle     sql.NullString `db:"org_title"`
	OrgName      sql.NullString `db:"org_name"`
	OrgAvatar    sql.NullString `db:"org_avatar"`
	ProjectCount sql.NullInt64  `db:"project_count"`
	RoleNames    pq.StringArray `db:"role_names"`
	RoleTitles   pq.StringArray `db:"role_titles"`
	RoleIDs      pq.StringArray `db:"role_ids"`
	OrgJoinedOn  sql.NullTime   `db:"org_joined_on"`
	UserID       sql.NullString `db:"user_id"`
}

type UserOrgsGroup struct {
	Name sql.NullString      `db:"name"`
	Data []UserOrgsGroupData `db:"data"`
}

type UserOrgsGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (u *UserOrgs) transformToAggregatedUserOrg() svc.AggregatedUserOrganization {
	return svc.AggregatedUserOrganization{
		OrgID:        u.OrgID.String,
		OrgTitle:     u.OrgTitle.String,
		OrgName:      u.OrgName.String,
		OrgAvatar:    u.OrgAvatar.String,
		ProjectCount: u.ProjectCount.Int64,
		RoleNames:    u.RoleNames,
		RoleTitles:   u.RoleTitles,
		RoleIDs:      u.RoleIDs,
		OrgJoinedOn:  u.OrgJoinedOn.Time,
		UserID:       u.UserID.String,
	}
}

func NewUserOrgsRepository(dbc *db.Client) *UserOrgsRepository {
	return &UserOrgsRepository{
		dbc: dbc,
	}
}

func (r UserOrgsRepository) Search(ctx context.Context, userID string, rql *rql.Query) (svc.UserOrgs, error) {
	return svc.UserOrgs{}, fmt.Errorf("not implemented sadasda")
}

// Helper function to prepare the data query (to be implemented)
func (r UserOrgsRepository) prepareDataQuery(userID string, rql *rql.Query) (string, []interface{}, error) {
	return "", nil, fmt.Errorf("not implemented")
}
