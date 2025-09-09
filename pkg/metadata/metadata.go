package metadata

import (
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
func Build(m map[string]any) Metadata {
	newMap := make(Metadata)
	for key, value := range m {
		newMap[key] = value
	}
	return newMap
}

// FromString transforms a Metadata from map[string]string
func FromString(m map[string]string) Metadata {
	newMap := make(Metadata)
	for key, value := range m {
		newMap[key] = value
	}
	return newMap
}

// BuildFromProto safely builds Metadata from a protobuf Struct.
// Returns an empty Metadata if the input is nil.
func BuildFromProto(pb *structpb.Struct) Metadata {
	if pb == nil {
		return make(Metadata)
	}
	return Build(pb.AsMap())
}
