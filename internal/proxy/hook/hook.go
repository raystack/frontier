package hook

import (
	"net/http"

	"github.com/odpf/shield/core/rule"
	"github.com/odpf/shield/internal/proxy/middleware"
)

type Service interface {
	Info() Info
	ServeHook(res *http.Response, err error) (*http.Response, error)
}

type Info struct {
	Name        string
	Description string
}

const (
	AttributeTypeJSONPayload AttributeType = "json_payload"
	AttributeTypeGRPCPayload AttributeType = "grpc_payload"
	AttributeTypeQuery       AttributeType = "query"
	AttributeTypeHeader      AttributeType = "header"
	AttributeTypeConstant    AttributeType = "constant"

	SourceRequest  AttributeType = "request"
	SourceResponse AttributeType = "response"
)

type AttributeType string

type Attribute struct {
	Key    string        `yaml:"key" mapstructure:"key"`
	Type   AttributeType `yaml:"type" mapstructure:"type"`
	Index  string        `yaml:"index" mapstructure:"index"` // proto index
	Source string        `yaml:"source" mapstructure:"source"`
	Value  string        `yaml:"value" mapstructure:"value"`
}

func ExtractHook(r *http.Request, name string) (rule.HookSpec, bool) {
	rl, ok := ExtractRule(r)
	if !ok {
		return rule.HookSpec{}, false
	}
	return rl.Hooks.Get(name)
}

func ExtractRule(r *http.Request) (*rule.Rule, bool) {
	rl, ok := middleware.ExtractRule(r)
	if !ok {
		return nil, false
	}

	return rl, true
}

type Hook struct{}

func New() Hook {
	return Hook{}
}

func (h Hook) Info() Info {
	return Info{}
}

func (h Hook) ServeHook(res *http.Response, err error) (*http.Response, error) {
	if err != nil {
		res.StatusCode = http.StatusInternalServerError
		// TODO: clear or add error body as well
	}

	return res, nil
}
