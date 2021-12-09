package casbin_authz

import (
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/structs"
)

// make sure the request is allowed & ready to be sent to backend
type CasbinAuthz struct {
	log  log.Logger
	next http.Handler
}

type Config struct {
	Action     string                          `yaml:"action" mapstructure:"action"`
	Attributes map[string]middleware.Attribute `yaml:"attributes" mapstructure:"attributes"` // auth field -> Attribute
}

func New(log log.Logger, next http.Handler) *CasbinAuthz {
	return &CasbinAuthz{log: log, next: next}
}

func (c CasbinAuthz) Info() *structs.MiddlewareInfo {
	return &structs.MiddlewareInfo{
		Name:        "authz",
		Description: "rule based authorization using casbin",
	}
}

func (c *CasbinAuthz) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wareSpec, ok := middleware.ExtractMiddleware(req, c.Info().Name)
	if !ok {
		c.next.ServeHTTP(rw, req)
		return
	}

	// TODO: should cache it
	config := Config{}
	if err := mapstructure.Decode(wareSpec.Config, &config); err != nil {
		c.log.Error("middleware: failed to decode authz config", "config", wareSpec.Config, "err", err)
		c.notAllowed(rw)
		return
	}

	// TODO: check if action matchers user capabilities
	// config.Action

	for res, attr := range config.Attributes {
		// TODO: do something about this
		_ = res

		switch attr.Type {
		case middleware.AttributeTypeGRPCPayload:
			// check if grpc request
			if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
				c.log.Error("middleware: not a grpc request", "attr", attr)
				c.notAllowed(rw)
				return
			}

			// TODO: we can optimise this by parsing all field at once
			payloadField, err := middleware.GRPCPayloadHandler{}.Extract(req, attr.Index)
			if err != nil {
				c.log.Error("middleware: failed to parse grpc payload", "err", err, "tr", "here")
				return
			}

			// do something about it
			_ = payloadField

			c.log.Info("middleware: extracted", "field", payloadField, "attr", attr)
		case middleware.AttributeTypeJSONPayload:
			if attr.Key == "" {
				c.log.Error("middleware: payload key field empty")
				c.notAllowed(rw)
				return
			}
			payloadField, err := middleware.JSONPayloadHandler{}.Extract(req, attr.Key)
			if err != nil {
				c.log.Error("middleware: failed to parse grpc payload", "err", err)
				c.notAllowed(rw)
				return
			}

			c.log.Info("middleware: extracted", "field", payloadField, "attr", attr)
		default:
			c.log.Error("middleware: unknown attribute type", "attr", attr)
			c.notAllowed(rw)
			return
		}
	}

	c.next.ServeHTTP(rw, req)
}

func (w CasbinAuthz) notAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusUnauthorized)
	return
}
