package postgres

import (
	"github.com/raystack/frontier/core/kyc"
	"time"

	"database/sql"
)

type KYC struct {
	OrgID     string       `db:"org_id"`
	Status    bool         `db:"status"`
	Link      string       `db:"link"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (from KYC) transformToKyc() (kyc.KYC, error) {
	return kyc.KYC{
		OrgID:     from.OrgID,
		Status:    from.Status,
		Link:      from.Link,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
