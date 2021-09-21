package middleware

import (
	"context"
	"net/http"

	"github.com/odpf/shield/structs"
)

const (
	ctxRuleKey = "middleware_rule"
)

func EnrichRule(r *http.Request, rule *structs.Rule) {
	*r = *r.WithContext(context.WithValue(r.Context(), ctxRuleKey, rule))
}

func ExtractRule(r *http.Request) (*structs.Rule, bool) {
	rl, ok := r.Context().Value(ctxRuleKey).(*structs.Rule)
	return rl, ok
}

func ExtractMiddleware(r *http.Request, name string) (structs.MiddlewareSpec, bool) {
	rl, ok := r.Context().Value(ctxRuleKey).(*structs.Rule)
	if !ok {
		return structs.MiddlewareSpec{}, false
	}
	return rl.Middlewares.Get(name)
}

const (
	AttributeTypeQuery       AttributeType = "query"
	AttributeTypeHeader      AttributeType = "header"
	AttributeTypeJSONPayload AttributeType = "json_payload"
	AttributeTypeGRPCPayload AttributeType = "grpc_payload"
)

type AttributeType string

type Attribute struct {
	Key   string        `yaml:"key" mapstructure:"key"`
	Type  AttributeType `yaml:"type" mapstructure:"type"`
	Index int           `yaml:"index" mapstructure:"index"` // proto index
}
