package hook

import (
	"net/http"

	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/structs"
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

	SourceRequest  AttributeType = "request"
	SourceResponse AttributeType = "response"
)

type AttributeType string

type Attribute struct {
	Key    string        `yaml:"key" mapstructure:"key"`
	Type   AttributeType `yaml:"type" mapstructure:"type"`
	Index  int           `yaml:"index" mapstructure:"index"` // proto index
	Source string        `yaml:"source" mapstructure:"source"`
}

func ExtractHook(r *http.Request, name string) (structs.HookSpec, bool) {
	rl, ok := middleware.ExtractRule(r)
	if !ok {
		return structs.HookSpec{}, false
	}
	return rl.Hooks.Get(name)
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
