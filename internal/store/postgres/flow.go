package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
)

type Flow struct {
	ID        uuid.UUID `db:"id"`
	Method    string    `db:"method"`
	Email     string    `db:"email"`
	Nonce     string    `db:"nonce"`
	Metadata  []byte    `db:"metadata"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (f *Flow) transformToFlow() (*authenticate.Flow, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(f.Metadata, &unmarshalledMetadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata of flow: %w", err)
	}
	startURL := ""
	if val, ok := unmarshalledMetadata["start_url"]; ok {
		startURL = val.(string)
	}
	finishURL := ""
	if val, ok := unmarshalledMetadata["finish_url"]; ok {
		finishURL = val.(string)
	}
	return &authenticate.Flow{
		ID:        f.ID,
		Method:    f.Method,
		Email:     f.Email,
		StartURL:  startURL,
		FinishURL: finishURL,
		Nonce:     f.Nonce,
		CreatedAt: f.CreatedAt,
		ExpiresAt: f.ExpiresAt,
		Metadata:  unmarshalledMetadata,
	}, nil
}
