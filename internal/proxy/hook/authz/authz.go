package authz

import (
	"context"
	"fmt"
	attributes2 "github.com/odpf/shield/internal/proxy/middleware/attributes"
	"net/http"

	"github.com/mitchellh/mapstructure"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/internal/proxy/hook"
)

type ResourceService interface {
	Create(ctx context.Context, resource resource.Resource) (resource.Resource, error)
}

type Authz struct {
	log log.Logger

	// To go to next hook
	next hook.Service

	// To skip all the next hooks and just respond back
	escape hook.Service

	projectService ProjectService

	resourceService ResourceService
}

type ProjectService interface {
	Get(ctx context.Context, id string) (project.Project, error)
}

func New(log log.Logger, next, escape hook.Service, resourceService ResourceService, projectService ProjectService) Authz {
	return Authz{
		log:             log,
		next:            next,
		escape:          escape,
		resourceService: resourceService,
		projectService:  projectService,
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
	if err != nil || res.StatusCode >= 400 {
		return a.escape.ServeHook(res, err)
	}

	attributes, ok := attributes2.GetAttributesFromContext(res.Request.Context())
	if !ok {
		return a.escape.ServeHook(res, fmt.Errorf("unable to fetch permission attributes from context"))
	}
	ns := attributes["namespace"]

	if ns == nil {
		return a.next.ServeHook(res, nil)
	}

	hookSpec, ok := hook.ExtractHook(res.Request, a.Info().Name)
	if !ok {
		return a.next.ServeHook(res, nil)
	}

	config := Config{}
	if err := mapstructure.Decode(hookSpec.Config, &config); err != nil {
		return a.next.ServeHook(res, nil)
	}

	if ns == "" {
		return a.next.ServeHook(res, fmt.Errorf("namespace variable not defined in rules"))
	}

	resources, err := a.createResources(res.Request.Context(), attributes)
	if err != nil {
		a.log.Error(err.Error())
		return a.escape.ServeHook(res, fmt.Errorf(err.Error()))
	}
	for _, resource := range resources {
		newResource, err := a.resourceService.Create(res.Request.Context(), resource)
		if err != nil {
			a.log.Error(err.Error())
			return a.escape.ServeHook(res, fmt.Errorf(err.Error()))
		}
		a.log.Info(fmt.Sprintf("Resource %s created with ID %s", newResource.URN, newResource.Idxa))
	}

	return a.next.ServeHook(res, nil)
}

func (a Authz) createResources(ctx context.Context, permissionAttributes map[string]interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	projects, err := getAttributesValues(permissionAttributes["project"])
	if err != nil {
		return nil, err
	}
	var organizationIds []string
	var projectIds []string
	for _, proj := range projects {
		project, err := a.projectService.Get(ctx, proj)
		if err != nil {
			return nil, err
		}

		organizationId := project.Organization.ID
		organizationIds = append(organizationIds, organizationId)

		projectId := project.ID
		projectIds = append(projectIds, projectId)
	}

	teams, err := getAttributesValues(permissionAttributes["team"])
	if err != nil {
		return nil, err
	}

	resourceList, err := getAttributesValues(permissionAttributes["resource"])
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

	if len(projects) < 1 || len(organizationIds) < 1 || len(resourceList) < 1 || (backendNamespace[0] == "") || (resourceType[0] == "") {
		return nil, fmt.Errorf("namespace, resource type, projects, resource, and team are required")
	}

	for _, org := range organizationIds {
		for _, project := range projectIds {
			for _, res := range resourceList {
				if len(teams) > 0 {
					for _, team := range teams {
						resources = append(resources, resource.Resource{
							Name:           res,
							OrganizationID: org,
							ProjectID:      project,
							GroupID:        team,
							NamespaceID:    namespace.CreateID(backendNamespace[0], resourceType[0]),
						})
					}
				} else {
					resources = append(resources, resource.Resource{
						Name:           res,
						OrganizationID: org,
						ProjectID:      project,
						NamespaceID:    namespace.CreateID(backendNamespace[0], resourceType[0]),
					})
				}
			}
		}
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
			iStr, ok := i.(string)
			if !ok {
				return values, fmt.Errorf("attribute type in []interface{} not string: %v", i)
			}
			values = append(values, iStr)
		}
	case interface{}:
		attrStr, ok := attributes.(string)
		if !ok {
			return values, fmt.Errorf("attribute type interface{} not string: %v", attributes)
		}
		values = append(values, attrStr)
	case nil:
		return values, nil
	default:
		return values, fmt.Errorf("unsuported attribute type: %v", attributes)
	}
	return values, nil
}
