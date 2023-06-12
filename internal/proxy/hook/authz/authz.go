package authz

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/raystack/salt/log"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/core/resource"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/internal/proxy/hook"
	"github.com/raystack/shield/internal/proxy/middleware"
	"github.com/raystack/shield/pkg/body_extractor"
	"github.com/raystack/shield/pkg/telemetry"
)

type ResourceService interface {
	Create(ctx context.Context, resource resource.Resource) (resource.Resource, error)
}

type RelationService interface {
	Create(ctx context.Context, relation relation.RelationV2) (relation.RelationV2, error)
}

type Authz struct {
	log log.Logger

	// To go to next hook
	next hook.Service

	// To skip all the next hooks and just respond back
	escape hook.Service

	identityProxyHeaderKey string

	resourceService ResourceService

	relationService RelationService
}

type ProjectService interface {
	Get(ctx context.Context, id string) (project.Project, error)
}

func New(log log.Logger, next, escape hook.Service, resourceService ResourceService, relationService RelationService, identityProxyHeaderKey string) Authz {
	return Authz{
		log:                    log,
		next:                   next,
		escape:                 escape,
		resourceService:        resourceService,
		relationService:        relationService,
		identityProxyHeaderKey: identityProxyHeaderKey,
	}
}

type Relation struct {
	Role               string `yaml:"role" mapstructure:"role"`
	SubjectPrincipal   string `yaml:"subject_principal" mapstructure:"subject_principal"`
	SubjectIDAttribute string `yaml:"subject_id" mapstructure:"subject_id_attribute"`
}

type Config struct {
	Action     string                    `yaml:"action" mapstructure:"action"`
	Attributes map[string]hook.Attribute `yaml:"attributes" mapstructure:"attributes"`
	Relations  []Relation                `yaml:"relations" mapstructure:"relations"`
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

	isResourceCreated := false
	attributes := map[string]interface{}{}

	defer func(isResourceCreated bool, ctx context.Context, attributes map[string]interface{}) {
		if !isResourceCreated {
			requestDetail := fmt.Sprintf("[status: %d, request: %s, attributes: %s ]", res.StatusCode, res.Request.Method+"@"+res.Request.URL.Host, attributes)
			ctx, err := tag.New(ctx, tag.Insert(telemetry.KeyMethod, "ServeHook"), tag.Insert(telemetry.KeyRequestDetails, requestDetail))
			if err != nil {
				a.log.Debug("failed to add metrics tags: ", err.Error())
			}
			stats.Record(ctx, telemetry.MResourceFailedToCreate.M(1))
		}
	}(isResourceCreated, res.Request.Context(), attributes)

	ruleFromRequest, ok := hook.ExtractRule(res.Request)
	if !ok {
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

	if ruleFromRequest.Backend.Namespace == "" {
		return a.next.ServeHook(res, fmt.Errorf("namespace variable not defined in rules"))
	}

	attributes["namespace"] = ruleFromRequest.Backend.Namespace

	identityProxyHeaderValue := res.Request.Header.Get(a.identityProxyHeaderKey)
	attributes["user"] = identityProxyHeaderValue
	res.Request = res.Request.WithContext(user.SetContextWithEmail(res.Request.Context(), identityProxyHeaderValue))

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
				a.log.Error("middleware: header key field empty")
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

		case hook.AttributeTypeConstant:
			if attr.Value == "" {
				a.log.Error("middleware: constant value empty")
				return a.escape.ServeHook(res, fmt.Errorf("failed to parse json payload"))
			}

			attributes[id] = attr.Value
			a.log.Info("middleware: extracted", "constant_key", res, "attr", attributes[id])

		default:
			a.log.Error("middleware: unknown attribute type", "attr", attr)
			return a.escape.ServeHook(res, fmt.Errorf("unknown attribute type: %v", attr))
		}
	}

	paramMap, _ := middleware.ExtractPathParams(res.Request)
	for key, value := range paramMap {
		attributes[key] = value
	}

	resources, err := a.createResources(attributes)
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

		isResourceCreated = true
		a.log.Info(fmt.Sprintf("Resource %s created with ID %s", newResource.URN, newResource.Idxa))

		for _, rel := range config.Relations {
			subjectId, err := getAttributesValues(attributes[rel.SubjectIDAttribute])
			if err != nil {
				a.log.Error(fmt.Sprintf("cannot create relation: %s not found in attributes", rel.SubjectIDAttribute))

				relationDetail := fmt.Sprintf("%s to %s %s on %s", rel.Role, rel.SubjectPrincipal, rel.SubjectIDAttribute, newResource.Name)
				ctx, err := tag.New(res.Request.Context(), tag.Insert(telemetry.KeyMethod, "ServeHook"), tag.Insert(telemetry.KeyRelationDetails, relationDetail))
				if err != nil {
					a.log.Debug("failed to add metrics tags: ", err.Error())
				}
				stats.Record(ctx, telemetry.MRelationFailedToCreate.M(1))

				continue
			}

			newRelation, err := a.relationService.Create(res.Request.Context(), relation.RelationV2{
				Object: relation.Object{
					ID:        newResource.Idxa,
					Namespace: newResource.NamespaceID,
				},
				Subject: relation.Subject{
					RoleID:    rel.Role,
					Namespace: rel.SubjectPrincipal,
					ID:        subjectId[0],
				},
			})
			if err != nil {
				a.log.Error(err.Error())

				relationDetail := fmt.Sprintf("%s to %s %s on %s", rel.Role, rel.SubjectPrincipal, rel.SubjectIDAttribute, newResource.Name)
				ctx, err := tag.New(res.Request.Context(), tag.Insert(telemetry.KeyMethod, "ServeHook"), tag.Insert(telemetry.KeyRelationDetails, relationDetail))
				if err != nil {
					a.log.Debug("failed to add metrics tags: ", err.Error())
				}
				stats.Record(ctx, telemetry.MRelationFailedToCreate.M(1))

				return a.escape.ServeHook(res, fmt.Errorf(err.Error()))
			}

			a.log.Info(fmt.Sprintf("created relation: %s for %s %s", newRelation.Subject.RoleID, newRelation.Subject.ID, newRelation.Subject.Namespace))
		}
	}

	return a.next.ServeHook(res, nil)
}

func (a Authz) createResources(permissionAttributes map[string]interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	projects, err := getAttributesValues(permissionAttributes["project"])
	if err != nil {
		return nil, err
	}

	//orgs, err := getAttributesValues(permissionAttributes["organization"])
	//if err != nil {
	//	return nil, err
	//}

	// TODO(krtkvrm): this will be decided on type of principal
	//teams, err := getAttributesValues(permissionAttributes["team"])
	//if err != nil {
	//	return nil, err
	//}

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

	if len(projects) < 1 || len(resourceList) < 1 || (backendNamespace[0] == "") || (resourceType[0] == "") {
		return nil, fmt.Errorf("namespace, resource type, projects, resource, and team are required")
	}

	// TODO(krtkvrm): needs revision
	for _, project := range projects {
		for _, res := range resourceList {
			resources = append(resources, resource.Resource{
				Name:        res,
				ProjectID:   project,
				NamespaceID: namespace.CreateID(backendNamespace[0], resourceType[0]),
			})
		}
	}

	return resources, nil
}

func getAttributesValues(attributes interface{}) ([]string, error) {
	var values []string
	switch attributes.(type) {
	case []string:
		values = append(values, attributes.([]string)...)
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
