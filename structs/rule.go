package structs

import (
	"context"
	"net/url"
)

type Service struct {
	Name string
	Rules []Rule
}

type Rule struct{
	Method string `yaml:"method"`
	Frontend Frontend `yaml:"frontend"`
	Backend Backend `yaml:"backend"`
	Permissions []Permission `yaml:"permissions"`
}

type Frontend struct {
	URL string `yaml:"url"`
}

type Backend struct {
	URL string `yaml:"url"`
}

type Permission struct {
	Action string `yaml:"action"`
	Attributes map[string]Attribute `yaml:"attributes"` // auth field -> Attribute
}

const (
	AttributeTypeQuery AttributeType = "query"
	AttributeTypeHeader AttributeType = "header"
	AttributeTypeJSONPayload AttributeType = "json_payload"
	AttributeTypeGRPCPayload AttributeType = "grpc_payload"
)

type AttributeType string

type Attribute struct {
	Name string `yaml:"name"`
	Type AttributeType `yaml:"type"`
	Index int `yaml:"index"` // proto index
}

type RuleMatcher interface {
	Match(ctx context.Context, reqMethod string, reqURL *url.URL) (*Rule, error)
}