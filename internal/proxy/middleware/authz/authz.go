package authz

import (
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/pkg/body_extractor"
	"net/http"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/proxy/middleware"
)

type ResourceService interface {
	CheckAuthz(ctx context.Context, resource resource.Resource, act action.Action) (bool, error)
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
}

type Authz struct {
	log             log.Logger
	userIDHeaderKey string
	next            http.Handler
	resourceService ResourceService
	userService     UserService
}

type Config struct {
	Actions     []string                        `yaml:"actions" mapstructure:"actions"`
	Permissions []Permission                    `yaml:"permissions" mapstructure:"permissions"`
	Attributes  map[string]middleware.Attribute `yaml:"attributes" mapstructure:"attributes"`
}

type Permission struct {
	Name      string
	Namespace string
	Attribute string
}

func New(
	log log.Logger,
	next http.Handler,
	userIDHeaderKey string,
	resourceService ResourceService,
	userService UserService) *Authz {
	return &Authz{
		log:             log,
		userIDHeaderKey: userIDHeaderKey,
		next:            next,
		resourceService: resourceService,
		userService:     userService,
	}
}

func (c Authz) Info() *middleware.MiddlewareInfo {
	return &middleware.MiddlewareInfo{
		Name:        "authz",
		Description: "rule based authorization using casbin",
	}
}

func (c *Authz) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	usr, err := c.userService.FetchCurrentUser(req.Context())
	if err != nil {
		c.log.Error("middleware: failed to get user details", "err", err.Error())
		c.notAllowed(rw)
		return
	}

	req.Header.Set(c.userIDHeaderKey, usr.ID)

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

	if rule.Backend.Namespace == "" {
		c.log.Error("namespace is not defined for this rule")
		c.notAllowed(rw)
		return
	}

	// TODO: should cache it
	config := Config{}
	if err := mapstructure.Decode(wareSpec.Config, &config); err != nil {
		c.log.Error("middleware: failed to decode authz config", "config", wareSpec.Config)
		c.notAllowed(rw)
		return
	}

	permissionAttributes := map[string]interface{}{}

	permissionAttributes["namespace"] = rule.Backend.Namespace

	permissionAttributes["user"] = req.Header.Get(c.userIDHeaderKey)

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

		case middleware.AttributeTypeConstant:
			if attr.Value == "" {
				c.log.Error("middleware: constant value empty")
				c.notAllowed(rw)
				return
			}

			permissionAttributes[res] = attr.Value
			c.log.Info("middleware: extracted", "constant_key", res, "attr", permissionAttributes[res])

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

	isAuthorized := false
	for _, permission := range config.Permissions {
		isAuthorized, err = c.resourceService.CheckAuthz(req.Context(), resource.Resource{
			Name:        permissionAttributes[permission.Attribute].(string),
			NamespaceID: permission.Namespace,
		}, action.Action{
			ID: permission.Name,
		})
		if err != nil {
			c.log.Error("error while creating resource obj", "err", err)
			c.notAllowed(rw)
			return
		}
		if isAuthorized {
			break
		}
	}

	c.log.Info("authz check successful", "user", permissionAttributes["user"], "resource", permissionAttributes["resource"], "result", isAuthorized)
	if !isAuthorized {
		c.log.Info("user not allowed to make request", "user", permissionAttributes["user"], "resource", permissionAttributes["resource"], "result", isAuthorized)
		c.notAllowed(rw)
		return
	}

	c.next.ServeHTTP(rw, req)
}

func (w Authz) notAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusUnauthorized)
}

func createResources(permissionAttributes map[string]interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	resourceList, err := getAttributesValues(permissionAttributes["resource"])
	if err != nil {
		return nil, err
	}

	project, err := getAttributesValues(permissionAttributes["project"])
	if err != nil {
		return nil, err
	}

	backendNamespace, err := getAttributesValues(permissionAttributes["namespace"])
	if err != nil {
		return nil, err
	}

	resourceType, err := getAttributesValues(permissionAttributes["resource_type"])
	if err != nil {
		return nil, err
	}

	if len(resourceList) < 1 {
		return nil, fmt.Errorf("resources are required")
	}

	for _, res := range resourceList {
		nsID := namespace.CreateID(backendNamespace[0], resourceType[0])
		resources = append(resources, resource.Resource{
			Name:        res,
			NamespaceID: nsID,
			ProjectID:   project[0],
		})
	}
	return resources, nil
}

func getAttributesValues(attributes interface{}) ([]string, error) {
	var values []string

	switch attributes := attributes.(type) {
	case []string:
		values = append(values, attributes...)
	case string:
		values = append(values, attributes)
	case []interface{}:
		for _, i := range attributes {
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
