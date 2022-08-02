package authz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/proxy/middleware"
	"github.com/odpf/shield/internal/proxy/middleware/attributes"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/salt/log"
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
	Actions    []string                        `yaml:"actions" mapstructure:"actions"`
	Attributes map[string]middleware.Attribute `yaml:"attributes" mapstructure:"attributes"` // auth field -> Attribute
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

	permissionAttributes, ok := attributes.GetAttributesFromContext(req.Context())
	if !ok {
		c.log.Error("unable to fetch permission attributes from context")
		c.notAllowed(rw)
		return
	}
	ns := permissionAttributes["namespace"]

	if ns == nil {
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

	if ns == "" {
		c.log.Error("namespace is not defined for this rule")
		c.notAllowed(rw)
		return
	}

	resources, err := createResources(permissionAttributes)
	if err != nil {
		c.log.Error("error while creating resource obj", "err", err)
		c.notAllowed(rw)
		return
	}
	for _, resource := range resources {
		isAuthorized := false
		for _, actionID := range config.Actions {
			isAuthorized, err = c.resourceService.CheckAuthz(req.Context(), resource, action.Action{ID: actionID})
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
			Namespace: namespace.Namespace{
				ID:           nsID,
				Backend:      backendNamespace[0],
				ResourceType: resourceType[0],
			},
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
