package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/raystack/frontier/core/user"
)

type User struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Email     string         `db:"email"`
	Title     sql.NullString `db:"title"`
	Metadata  []byte         `db:"metadata"`
	State     sql.NullString `db:"state"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

func (from User) transformToUser() (user.User, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return user.User{}, err
		}
	}

	return user.User{
		ID:        from.ID,
		Name:      from.Name,
		Email:     from.Email,
		State:     user.State(from.State.String),
		Title:     from.Title.String,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
