package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/odpf/shield/structs"
)

const (
	ctxRuleKey       = "middleware_rule"
	ctxPathParamsKey = "path_params"
	ctxBodyKey       = "body_ctx"
)

func EnrichRule(r *http.Request, rule *structs.Rule) {
	*r = *r.WithContext(context.WithValue(r.Context(), ctxRuleKey, rule))
}

func EnrichRequestBody(r *http.Request) error {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer (r.Body).Close()

	// repopulate body
	(*r).Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
	*r = *r.WithContext(context.WithValue(r.Context(), ctxBodyKey, reqBody))
	return nil
}

func ExtractRequestBody(r *http.Request) (io.ReadCloser, bool) {
	body, ok := r.Context().Value(ctxBodyKey).([]byte)
	if !ok {
		return nil, false
	}
	return ioutil.NopCloser(bytes.NewBuffer(body)), true
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

func EnrichPathParams(r *http.Request, params map[string]string) {
	*r = *r.WithContext(context.WithValue(r.Context(), ctxPathParamsKey, params))
}

func ExtractPathParams(r *http.Request) (map[string]string, bool) {
	params, ok := r.Context().Value(ctxPathParamsKey).(map[string]string)
	if !ok {
		return nil, false
	}
	return params, true
}

const (
	AttributeTypeQuery       AttributeType = "query"
	AttributeTypeHeader      AttributeType = "header"
	AttributeTypeJSONPayload AttributeType = "json_payload"
	AttributeTypeGRPCPayload AttributeType = "grpc_payload"
	AttributeTypePathParam   AttributeType = "path_param"
)

type AttributeType string

type Attribute struct {
	Key    string        `yaml:"key" mapstructure:"key"`
	Type   AttributeType `yaml:"type" mapstructure:"type"`
	Index  int           `yaml:"index" mapstructure:"index"` // proto index
	Path   string        `yaml:"path" mapstructure:"path"`
	Params []string      `yaml:"params" mapstructure:"params"`
}

func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}
