package authz

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/odpf/shield/api/handler"
	"github.com/odpf/shield/internal/permission"
	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/model"
	"github.com/odpf/shield/pkg/body_extractor"
	"github.com/odpf/shield/structs"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
)

type AuthzCheckService interface {
	CheckAuthz(ctx context.Context, resource model.Resource, action model.Action) (bool, error)
}

type Authz struct {
	log                 log.Logger
	identityProxyHeader string
	next                http.Handler
	Deps                handler.Deps
	AuthzCheckService   AuthzCheckService
}

type Config struct {
	Actions    []string                        `yaml:"actions" mapstructure:"actions"`
	Attributes map[string]middleware.Attribute `yaml:"attributes" mapstructure:"attributes"` // auth field -> Attribute
}

func New(log log.Logger, identityProxyHeader string, deps handler.Deps, next http.Handler, authzCheckService AuthzCheckService) *Authz {
	return &Authz{log: log, identityProxyHeader: identityProxyHeader, Deps: deps, next: next, AuthzCheckService: authzCheckService}
}

func (c Authz) Info() *structs.MiddlewareInfo {
	return &structs.MiddlewareInfo{
		Name:        "authz",
		Description: "rule based authorization using casbin",
	}
}

func (c *Authz) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rule, ok := middleware.ExtractRule(req)
	if !ok {
		c.next.ServeHTTP(rw, req)
		return
	}

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

	if rule.Backend.Namespace == "" {
		c.log.Error("namespace is not defined for this rule")
		c.notAllowed(rw)
		return
	}

	permissionAttributes := map[string]interface{}{}

	permissionAttributes["namespace"] = rule.Backend.Namespace

	permissionAttributes["user"] = req.Header.Get(c.identityProxyHeader)
	req = req.WithContext(permission.SetEmailToContext(req.Context(), req.Header.Get(c.identityProxyHeader)))

	for res, attr := range config.Attributes {
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
			payloadField, err := body_extractor.GRPCPayloadHandler{}.Extract(&req.Body, attr.Index)
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
			payloadField, err := body_extractor.JSONPayloadHandler{}.Extract(&req.Body, attr.Key)
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

		default:
			c.log.Error("middleware: unknown attribute type", "attr", attr)
			c.notAllowed(rw)
			return
		}
	}

	paramMap, mapExists := middleware.ExtractPathParams(req)
	if !mapExists {
		c.log.Error("middleware: path param map doesn't exist")
		c.notAllowed(rw)
		return
	}

	for key, value := range paramMap {
		permissionAttributes[key] = value
	}

	resources, err := createResources(permissionAttributes)
	if err != nil {
		c.log.Error("error while creating resource obj", "err", err)
		c.notAllowed(rw)
		return
	}
	for _, resource := range resources {
		isAuthorized := false
		for _, actionId := range config.Actions {
			isAuthorized, err = c.AuthzCheckService.CheckAuthz(req.Context(), resource, model.Action{Id: actionId})
			if err != nil {
				c.log.Error("error while creating resource obj", "err", err)
				c.notAllowed(rw)
				return
			}

			if isAuthorized {
				break
			}
		}

		c.log.Info("authz check successful", "user", permissionAttributes["user"], "resource", resource.Name, "result", isAuthorized)
		if !isAuthorized {
			c.log.Info("user not allowed to make request", "user", permissionAttributes["user"], "resource", resource.Name, "result", isAuthorized)
			c.notAllowed(rw)
			return
		}
	}

	c.next.ServeHTTP(rw, req)
}

func (w Authz) notAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusUnauthorized)
	return
}

func createResources(permissionAttributes map[string]interface{}) ([]model.Resource, error) {
	var resources []model.Resource
	resourceList, err := getAttributesValues(permissionAttributes["resource"])
	if err != nil {
		return nil, err
	}

	namespace, err := getAttributesValues(permissionAttributes["namespace"])
	if err != nil {
		return nil, err
	}

	if len(resourceList) < 1 {
		return nil, fmt.Errorf("projects, organizations, resource, and team are required")
	}

	for _, res := range resourceList {
		resources = append(resources, model.Resource{
			Name:        res,
			NamespaceId: namespace[0],
		})
	}
	return resources, nil
}

func getAttributesValues(attributes interface{}) ([]string, error) {
	var values []string
	switch attributes.(type) {
	case []string:
		for _, i := range attributes.([]string) {
			values = append(values, i)
		}
	case string:
		values = append(values, attributes.(string))
	case []interface{}:
		for _, i := range attributes.([]interface{}) {
			values = append(values, i.(string))
		}
	case interface{}:
		values = append(values, attributes.(string))
	case nil:
		return values, nil
	default:
		return values, fmt.Errorf("unsuported attribute type %v", attributes)
	}
	return values, nil
}
