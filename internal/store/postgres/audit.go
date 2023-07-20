package postgres

import (
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/shield/core/audit"
)

type Audit struct {
	ID     string `db:"id"`
	OrgID  string `db:"org_id"`
	Source string `db:"source"`
	Action string `db:"action"`

	Actor    types.NullJSONText `db:"actor"`
	Target   types.NullJSONText `db:"target"`
	Metadata types.NullJSONText `db:"metadata"`

	CreatedAt time.Time `db:"created_at"`
}

func (a Audit) transform() (audit.Log, error) {
	var unmarshalledMetadata map[string]string
	if a.Metadata.Valid {
		if err := a.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return audit.Log{}, err
		}
	}
	var actor audit.Actor
	if a.Actor.Valid {
		if err := a.Actor.Unmarshal(&actor); err != nil {
			return audit.Log{}, err
		}
	}
	var target audit.Target
	if a.Target.Valid {
		if err := a.Target.Unmarshal(&target); err != nil {
			return audit.Log{}, err
		}
	}
	return audit.Log{
		ID:        a.ID,
		OrgID:     a.OrgID,
		Source:    a.Source,
		Action:    a.Action,
		CreatedAt: a.CreatedAt,
		Actor:     actor,
		Target:    target,
		Metadata:  unmarshalledMetadata,
	}, nil
}
