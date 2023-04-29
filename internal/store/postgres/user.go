package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/odpf/shield/core/user"
)

type User struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Email     string         `db:"email"`
	Metadata  []byte         `db:"metadata"`
	State     sql.NullString `db:"state"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

type UserMetadataKey struct {
	Key         string    `db:"key"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (from User) transformToUser() (user.User, error) {
	var unmarshalledMetadata map[string]any
	if from.Metadata != nil {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return user.User{}, err
		}
	}

	return user.User{
		ID:        from.ID,
		Name:      from.Name,
		Email:     from.Email,
		State:     user.State(from.State.String),
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}

func (from UserMetadataKey) tranformUserMetadataKey() user.UserMetadataKey {
	return user.UserMetadataKey{
		Key:         from.Key,
		Description: from.Description,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}
}
