package attributes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/proxy/middleware"
	"github.com/odpf/shield/pkg/body_extractor"
)

type Attributes struct {
	log                    log.Logger
	next                   http.Handler
	identityProxyHeaderKey string
}

type Config struct {
	Attributes map[string]middleware.Attribute `yaml:"attributes" mapstructure:"attributes"`
}

func New(
	log log.Logger,
	next http.Handler,
	identityProxyHeaderKey string,
) *Attributes {
	return &Attributes{
		log:                    log,
		next:                   next,
		identityProxyHeaderKey: identityProxyHeaderKey,
	}
}

func (a Attributes) Info() *middleware.MiddlewareInfo {
	return &middleware.MiddlewareInfo{
		Name:        "attributes",
		Description: "Attribute Extraction",
	}
}

func (a Attributes) notAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusUnauthorized)
	return
}

func (a *Attributes) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestAttributes := map[string]any{}

	wareSpec, _ := middleware.ExtractMiddleware(req, a.Info().Name)

	req = req.WithContext(user.SetContextWithEmail(req.Context(), req.Header.Get(a.identityProxyHeaderKey)))
	requestAttributes["user"] = req.Header.Get(a.identityProxyHeaderKey)

	config := Config{}
	if err := mapstructure.Decode(wareSpec.Config, &config); err != nil {
		a.log.Error("middleware: invalid config", "config", wareSpec.Config)
		a.notAllowed(rw)
		return
	}

	rule, ok := middleware.ExtractRule(req)
	if ok {
		requestAttributes["namespace"] = rule.Backend.Namespace
		requestAttributes["prefix"] = rule.Backend.Prefix
	}

	for res, attr := range config.Attributes {
		_ = res

		switch attr.Type {
		case middleware.AttributeTypeGRPCPayload:
			// check if grpc request
			if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
				a.log.Error("middleware: not a grpc request", "attr", attr)
				a.notAllowed(rw)
				return
			}

			// TODO: we can optimise this by parsing all field at once
			payloadField, err := body_extractor.GRPCPayloadHandler{}.Extract(&req.Body, attr.Index)
			if err != nil {
				a.log.Error("middleware: failed to parse grpc payload", "err", err)
				return
			}

			requestAttributes[res] = payloadField
			a.log.Info("middleware: extracted", "field", payloadField, "attr", attr)

		case middleware.AttributeTypeJSONPayload:
			if attr.Key == "" {
				a.log.Error("middleware: payload key field empty")
				a.notAllowed(rw)
				return
			}
			payloadField, err := body_extractor.JSONPayloadHandler{}.Extract(&req.Body, attr.Key)
			if err != nil {
				a.log.Error("middleware: failed to parse grpc payload", "err", err)
				a.notAllowed(rw)
				return
			}

			requestAttributes[res] = payloadField
			a.log.Info("middleware: extracted", "field", payloadField, "attr", attr)

		case middleware.AttributeTypeHeader:
			if attr.Key == "" {
				a.log.Error("middleware: header key field empty")
				a.notAllowed(rw)
				return
			}
			headerAttr := req.Header.Get(attr.Key)
			if headerAttr == "" {
				a.log.Error(fmt.Sprintf("middleware: header %s is empty", attr.Key))
				a.notAllowed(rw)
				return
			}

			requestAttributes[res] = headerAttr
			a.log.Info("middleware: extracted", "field", headerAttr, "attr", attr)

		case middleware.AttributeTypeQuery:
			if attr.Key == "" {
				a.log.Error("middleware: query key field empty")
				a.notAllowed(rw)
				return
			}
			queryAttr := req.URL.Query().Get(attr.Key)
			if queryAttr == "" {
				a.log.Error(fmt.Sprintf("middleware: query %s is empty", attr.Key))
				a.notAllowed(rw)
				return
			}

			requestAttributes[res] = queryAttr
			a.log.Info("middleware: extracted", "field", queryAttr, "attr", attr)

		case middleware.AttributeTypeConstant:
			if attr.Value == "" {
				a.log.Error("middleware: constant value empty")
				a.notAllowed(rw)
				return
			}

			requestAttributes[res] = attr.Value
			a.log.Info("middleware: extracted", "constant_key", res, "attr", requestAttributes[res])

		default:
			a.log.Error("middleware: unknown attribute type", "attr", attr)
			a.notAllowed(rw)
			return
		}
	}

	paramMap, mapExists := middleware.ExtractPathParams(req)
	if !mapExists {
		a.log.Error("middleware: path param map doesn't exist")
		a.notAllowed(rw)
		return
	}

	for key, value := range paramMap {
		requestAttributes[key] = value
	}

	attributesContext := context.WithValue(req.Context(), "requestAttributes", requestAttributes)
	req = req.WithContext(attributesContext)

	a.next.ServeHTTP(rw, req)
}
