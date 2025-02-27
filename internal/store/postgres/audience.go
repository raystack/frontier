package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/audience"
)

type Audience struct {
	ID        uuid.UUID       `db:"id" goqu:"skipinsert"`
	Name      sql.NullString  `db:"name"`
	Email     string          `db:"email"`
	Phone     sql.NullString  `db:"phone"`
	Activity  string          `db:"activity"`
	Status    audience.Status `db:"status"`
	ChangedAt time.Time       `db:"changed_at"`
	Source    string          `db:"source"`
	Verified  bool            `db:"verified"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
	Metadata  []byte          `db:"metadata"`
}

func (a *Audience) transformToAudience() (audience.Audience, error) {
	var unmarshalledMetadata map[string]any
	if len(a.Metadata) > 0 {
		if err := json.Unmarshal(a.Metadata, &unmarshalledMetadata); err != nil {
		}
	}
	return audience.Audience{
		ID:        a.ID.String(),
		Name:      a.Name.String,
		Email:     a.Email,
		Phone:     a.Phone.String,
		Activity:  a.Activity,
		Status:    a.Status.ToDB(),
		ChangedAt: a.ChangedAt,
		Source:    a.Source,
		Verified:  a.Verified,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Metadata:  unmarshalledMetadata,
	}, nil
}
