package authz

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/odpf/shield/hook"
	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/pkg/body_extractor"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
)

type Authz struct {
	log log.Logger

	// To go to next hook
	next hook.Service

	// To skip all the next hooks and just respond back
	escape hook.Service
}

func New(log log.Logger, next, escape hook.Service) Authz {
	return Authz{
		log:    log,
		next:   next,
		escape: escape,
	}
}

type Config struct {
	Action     string                    `yaml:"action" mapstructure:"action"`
	Attributes map[string]hook.Attribute `yaml:"attributes" mapstructure:"attributes"`
}

func (a Authz) Info() hook.Info {
	return hook.Info{
		Name:        "authz",
		Description: "hook to modify permissions for the resource",
	}
}

func (a Authz) ServeHook(res *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return a.escape.ServeHook(res, err)
	}

	hookSpec, ok := hook.ExtractHook(res.Request, a.Info().Name)
	if !ok {
		return a.next.ServeHook(res, nil)
	}

	config := Config{}
	if err := mapstructure.Decode(hookSpec.Config, &config); err != nil {
		return a.next.ServeHook(res, nil)
	}

	attributes := map[string]interface{}{}
	for id, attr := range config.Attributes {
		bdy, _ := middleware.ExtractRequestBody(res.Request)
		bodySource := &res.Body
		if attr.Source == string(hook.SourceRequest) {
			bodySource = &bdy
		}

		headerSource := &res.Header
		if attr.Source == string(hook.SourceRequest) {
			headerSource = &res.Request.Header
		}

		switch attr.Type {
		case hook.AttributeTypeGRPCPayload:
			if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/grpc") {
				a.log.Error("middleware: not a grpc request", "attr", attr)
				return a.escape.ServeHook(res, fmt.Errorf("invalid header for http request: %s", res.Header.Get("Content-Type")))
			}

			payloadField, err := body_extractor.GRPCPayloadHandler{}.Extract(bodySource, attr.Index)
			if err != nil {
				a.log.Error("middleware: failed to parse grpc payload", "err", err)
				return a.escape.ServeHook(res, fmt.Errorf("unable to parse grpc payload"))
			}
			attributes[id] = payloadField

			a.log.Info("middleware: extracted", "field", payloadField, "attr", attr)
		case hook.AttributeTypeJSONPayload:
			if attr.Key == "" {
				a.log.Error("middleware: payload key field empty")
				return a.escape.ServeHook(res, fmt.Errorf("payload key field empty"))
			}
			payloadField, err := body_extractor.JSONPayloadHandler{}.Extract(bodySource, attr.Key)
			if err != nil {
				a.log.Error("middleware: failed to parse json payload", "err", err)
				return a.escape.ServeHook(res, fmt.Errorf("failed to parse json payload"))
			}
			attributes[id] = payloadField

			a.log.Info("middleware: extracted", "field", payloadField, "attr", attr)
		case hook.AttributeTypeHeader:
			if attr.Key == "" {
				a.log.Error("middleware: header key field empty", "err", err)
				return a.escape.ServeHook(res, fmt.Errorf("failed to parse json payload"))
			}
			headerAttr := headerSource.Get(attr.Key)
			if headerAttr == "" {
				a.log.Error(fmt.Sprintf("middleware: header %s is empty", attr.Key))
				return a.escape.ServeHook(res, fmt.Errorf("failed to parse json payload"))
			}

			attributes[id] = headerAttr
			a.log.Info("middleware: extracted", "field", headerAttr, "attr", attr)

		case hook.AttributeTypeQuery:
			if attr.Key == "" {
				a.log.Error("middleware: query key field empty")
				return a.escape.ServeHook(res, fmt.Errorf("failed to parse json payload"))
			}
			queryAttr := res.Request.URL.Query().Get(attr.Key)
			if queryAttr == "" {
				a.log.Error(fmt.Sprintf("middleware: query %s is empty", attr.Key))
				return a.escape.ServeHook(res, fmt.Errorf("failed to parse json payload"))
			}

			attributes[id] = queryAttr
			a.log.Info("middleware: extracted", "field", queryAttr, "attr", attr)
		default:
			a.log.Error("middleware: unknown attribute type", "attr", attr)
			return a.escape.ServeHook(res, fmt.Errorf("unknown attribute type: %v", attr))
		}
	}

	//Change after merging PR#32
	//paramMap, _ := middleware.ExtractPathParams(req)
	//for key, value := range paramMap {
	//	permissionAttributes[key] = value
	//}

	// use attributes to modify authz
	// add respu

	return a.next.ServeHook(res, nil)
}
