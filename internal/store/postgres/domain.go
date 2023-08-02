package postgres

import (
	"database/sql"
	"time"

	"github.com/raystack/frontier/core/domain"
)

type Domain struct {
	ID         string       `db:"id"`
	OrgID      string       `db:"org_id"`
	Name       string       `db:"name"`
	Token      string       `db:"token"`
	Verified   bool         `db:"verified"`
	VerifiedAt sql.NullTime `db:"verified_at"`
	CreatedAt  time.Time    `db:"created_at"`
}

func (d Domain) transform() domain.Domain {
	return domain.Domain{
		ID:         d.ID,
		OrgID:      d.OrgID,
		Name:       d.Name,
		Token:      d.Token,
		Verified:   d.Verified,
		VerifiedAt: d.VerifiedAt.Time,
		CreatedAt:  d.CreatedAt,
	}
}
