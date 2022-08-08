package metadata

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

type Metadata map[string]any

func (m Metadata) ToStructPB() (*structpb.Struct, error) {
	newMap := make(map[string]interface{})

	for key, value := range m {
		newMap[key] = value
	}

	return structpb.NewStruct(newMap)
}

func Build(m map[string]interface{}) (Metadata, error) {
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
