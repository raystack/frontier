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

// nullStringToPtr converts a sql.NullString to *string.
// invalid strings will be converted to nil.
func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// nullStringToString converts a sql.NullString to string.
// invalid strings will be converted to empty string "".
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// ptrToString safely converts a string pointer to a string, returning empty string if nil
func ptrToString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// unmarshalNullJSONText unmarshals NullJSONText to map[string]any
func unmarshalNullJSONText(metadata types.NullJSONText) (map[string]any, error) {
	if !metadata.Valid {
		return nil, nil
	}
	var result map[string]any
	if err := metadata.Unmarshal(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// getStringFromMap extracts a string value from a map by key, returns empty string if not found or not a string
func getStringFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
