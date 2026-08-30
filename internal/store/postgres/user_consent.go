package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/raystack/frontier/core/consent"
)

// UserConsent is a row of user_consents: immutable, so no updated_at and no
// deleted_at, and it outlives the user, so no foreign key back to users.
type UserConsent struct {
	ID           string         `db:"id" goqu:"skipinsert"`
	UserID       string         `db:"user_id"`
	UserEmail    string         `db:"user_email"`
	Documents    []byte         `db:"documents"`
	Source       string         `db:"source"`
	AuthStrategy sql.NullString `db:"auth_strategy"`
	IPAddress    sql.NullString `db:"ip_address"`
	ConsentedAt  time.Time      `db:"consented_at"`
	CreatedAt    time.Time      `db:"created_at" goqu:"skipinsert"`
}

// ConsentDocument is one entry in the documents JSONB array, copied from config
// at write time so a record stays correct after the document leaves config.
type ConsentDocument struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

func (from UserConsent) transformToConsent() (consent.Consent, error) {
	var documents []ConsentDocument
	if len(from.Documents) > 0 {
		if err := json.Unmarshal(from.Documents, &documents); err != nil {
			return consent.Consent{}, err
		}
	}

	transformed := make([]consent.Document, 0, len(documents))
	for _, document := range documents {
		transformed = append(transformed, consent.Document{
			ID:      document.ID,
			Title:   document.Title,
			Version: document.Version,
			URL:     document.URL,
		})
	}

	return consent.Consent{
		ID:           from.ID,
		UserID:       from.UserID,
		UserEmail:    from.UserEmail,
		Documents:    transformed,
		Source:       from.Source,
		AuthStrategy: from.AuthStrategy.String,
		IPAddress:    from.IPAddress.String,
		ConsentedAt:  from.ConsentedAt,
		CreatedAt:    from.CreatedAt,
	}, nil
}

// marshalConsentDocuments renders the document list for the JSONB column.
func marshalConsentDocuments(documents []consent.Document) ([]byte, error) {
	rows := make([]ConsentDocument, 0, len(documents))
	for _, document := range documents {
		rows = append(rows, ConsentDocument{
			ID:      document.ID,
			Title:   document.Title,
			Version: document.Version,
			URL:     document.URL,
		})
	}
	return json.Marshal(rows)
}
