package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/pkg/metadata"
)

func setMetadataColumn(record goqu.Record, md metadata.Metadata) error {
	if md == nil {
		return nil
	}
	marshaledMetadata, err := json.Marshal(md)
	if err != nil {
		return fmt.Errorf("%w: %w", errParse, err)
	}
	record["metadata"] = marshaledMetadata
	return nil
}
