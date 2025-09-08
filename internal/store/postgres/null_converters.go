package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/pkg/metadata"
)

// nullJSONTextToMetadata converts NullJSONText to metadata.Metadata
func nullJSONTextToMetadata(njt types.NullJSONText) metadata.Metadata {
	if !njt.Valid || len(njt.JSONText) == 0 {
		return metadata.Metadata{}
	}

	var m map[string]any
	if err := json.Unmarshal(njt.JSONText, &m); err != nil {
		return metadata.Metadata{}
	}
	return metadata.Build(m)
}

// metadataToNullJSONText converts metadata.Metadata to NullJSONText.
// Empty or nil metadata will be stored as NULL in the database, not as an empty JSON object.
func metadataToNullJSONText(m metadata.Metadata) types.NullJSONText {
	if m == nil || len(m) == 0 {
		return types.NullJSONText{Valid: false}
	}

	data, err := json.Marshal(m)
	if err != nil {
		return types.NullJSONText{Valid: false}
	}

	return types.NullJSONText{
		JSONText: types.JSONText(data),
		Valid:    true,
	}
}

// toNullString converts a string to sql.NullString.
// Empty strings will be stored as NULL in the database, not as empty string "".
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
