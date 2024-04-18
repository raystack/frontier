package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/serviceuser"
)

type ServiceUser struct {
	ID        string         `db:"id"`
	OrgID     string         `db:"org_id"`
	Title     sql.NullString `db:"title"`
	State     sql.NullString `db:"state"`
	Metadata  []byte         `db:"metadata"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

func (s ServiceUser) transform() (serviceuser.ServiceUser, error) {
	var unmarshalledMetadata map[string]any
	if len(s.Metadata) > 0 {
		if err := json.Unmarshal(s.Metadata, &unmarshalledMetadata); err != nil {
			return serviceuser.ServiceUser{}, err
		}
	}
	return serviceuser.ServiceUser{
		ID:        s.ID,
		OrgID:     s.OrgID,
		Title:     s.Title.String,
		State:     s.State.String,
		Metadata:  unmarshalledMetadata,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}

type ServiceUserCredential struct {
	ID            string         `db:"id"`
	ServiceUserID string         `db:"serviceuser_id"`
	Type          sql.NullString `db:"type"`
	SecretHash    sql.NullString `db:"secret_hash"`
	PublicKey     []byte         `db:"public_key"`
	Title         sql.NullString `db:"title"`
	Metadata      []byte         `db:"metadata"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	DeletedAt     sql.NullTime   `db:"deleted_at"`
}

func (s ServiceUserCredential) transform() (serviceuser.Credential, error) {
	var unmarshalledMetadata map[string]any
	if len(s.Metadata) > 0 {
		if err := json.Unmarshal(s.Metadata, &unmarshalledMetadata); err != nil {
			return serviceuser.Credential{}, err
		}
	}
	var keySet jwk.Set
	if len(s.SecretHash.String) == 0 {
		// if a secret hash is created, public key would be null
		set, err := jwk.Parse(s.PublicKey)
		if err != nil {
			return serviceuser.Credential{}, fmt.Errorf("failed to parse public key: %w", err)
		}
		keySet = set
	}

	return serviceuser.Credential{
		ID:            s.ID,
		ServiceUserID: s.ServiceUserID,
		Type:          serviceuser.CredentialType(s.Type.String),
		SecretHash:    s.SecretHash.String,
		PublicKey:     keySet,
		Title:         s.Title.String,
		Metadata:      unmarshalledMetadata,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}, nil
}
