package metadata

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// Metadata is a structure to store dynamic values in
// frontier. it could be use as an additional information
// of a specific entity
type Metadata map[string]any

// ToStructPB transforms Metadata to *structpb.Struct
func (m Metadata) ToStructPB() (*structpb.Struct, error) {
	newMap := make(map[string]any)

	for key, value := range m {
		newMap[key] = value
	}

	return structpb.NewStruct(newMap)
}

// Build transforms a Metadata from map[string]any
func Build(m map[string]any) (Metadata, error) {
	newMap := make(Metadata)

	for key, value := range m {
		switch value := value.(type) {
		case any:
			newMap[key] = value
		default:
			return Metadata{}, fmt.Errorf("value for %s key is not string", key)
		}
	}

	return newMap, nil
}
