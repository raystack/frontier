package authz

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/structs"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
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
		c.log.Error("middleware: failed to decode authz config", "config", wareSpec.Config)
		c.notAllowed(rw)
		return
	}

	// TODO: check if action matchers user capabilities
	// config.Action

	permissionAttributes := map[string]string{}
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
				c.log.Error("middleware: failed to parse grpc payload", "err", err)
				return
			}

			permissionAttributes[res] = payloadField
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

			permissionAttributes[res] = payloadField
			c.log.Info("middleware: extracted", "field", payloadField, "attr", attr)

		case middleware.AttributeTypeHeader:
			if attr.Key == "" {
				c.log.Error("middleware: header key field empty")
				c.notAllowed(rw)
				return
			}
			headerAttr := req.Header.Get(attr.Key)
			if headerAttr == "" {
				c.log.Error(fmt.Sprintf("middleware: header %s is empty", attr.Key))
				c.notAllowed(rw)
				return
			}

			permissionAttributes[res] = headerAttr
			c.log.Info("middleware: extracted", "field", headerAttr, "attr", attr)

		case middleware.AttributeTypeQuery:
			if attr.Key == "" {
				c.log.Error("middleware: query key field empty")
				c.notAllowed(rw)
				return
			}
			queryAttr := req.URL.Query().Get(attr.Key)
			if queryAttr == "" {
				c.log.Error(fmt.Sprintf("middleware: query %s is empty", attr.Key))
				c.notAllowed(rw)
				return
			}

			permissionAttributes[res] = queryAttr
			c.log.Info("middleware: extracted", "field", queryAttr, "attr", attr)

		case middleware.AttributeTypePathParam:
			if attr.Path == "" {
				c.log.Error("middleware: path_param path empty")
				c.notAllowed(rw)
				return
			}

			route := new(mux.Route)
			route.Path(attr.Path)
			routeMatcher := mux.RouteMatch{}
			if !route.Match(req, &routeMatcher) {
				c.log.Error(fmt.Sprintf("middleware: path param %s not matching with incoming request %s", attr.Key, req.URL))
				c.notAllowed(rw)
				return
			}

			for _, paramName := range attr.Params {
				paramAttr, ok := routeMatcher.Vars[paramName]
				if !ok {
					c.log.Error(fmt.Sprintf("middleware: path param %s not found", attr.Key))
					c.notAllowed(rw)
					return
				}

				permissionAttributes[res] = paramAttr
				c.log.Info("middleware: extracted", "field", paramAttr, "attr", attr)
			}

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
