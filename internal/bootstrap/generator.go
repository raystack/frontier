package bootstrap

import (
	"context"
	"fmt"
	"strings"

	aznamespace "github.com/authzed/spicedb/pkg/namespace"
	azcore "github.com/authzed/spicedb/pkg/proto/core/v1"
	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
	"github.com/authzed/spicedb/pkg/schemadsl/generator"
	"github.com/authzed/spicedb/pkg/schemautil"
	"github.com/odpf/shield/internal/bootstrap/schema"
)

func ValidatePreparedAZSchema(ctx context.Context, azSchemaSource string) error {
	// compile and validate generated schema
	tenantName := "shield"
	updatedSchema, err := compiler.Compile(compiler.InputSchema{
		Source:       "generated",
		SchemaString: azSchemaSource,
	}, &tenantName)
	if err != nil {
		return fmt.Errorf("compile: failed to compile authz schema: %w", err)
	}

	if _, err = schemautil.ValidateSchemaChanges(ctx, updatedSchema, false); err != nil {
		return fmt.Errorf("ValidateSchemaChanges: failed to validate authz schema: %w", err)
	}
	return nil
}

func PrepareSchemaAsSource(authzedDefinitions []*azcore.NamespaceDefinition) (string, error) {
	preparedSchemaString := ""
	for _, def := range authzedDefinitions {
		generatedDefString, _, err := generator.GenerateSource(def)
		if err != nil {
			return "", fmt.Errorf("generateSource: failed to compile authz schema: %w", err)
		}
		preparedSchemaString = fmt.Sprintf("%s\n\n%s", preparedSchemaString, generatedDefString)
	}
	return preparedSchemaString, nil
}

func GetNamespaceName(serviceName, resourceName string) string {
	return fmt.Sprintf("%s/%s", serviceName, resourceName)
}

func GetBaseAZSchema() []*azcore.NamespaceDefinition {
	tenantName := "shield"
	compiledSchema, err := compiler.Compile(compiler.InputSchema{
		Source:       "base_schema.zed",
		SchemaString: schema.BaseSchemaZed,
	}, &tenantName)
	if err != nil {
		// this should not happen
		panic(err)
	}
	return compiledSchema.ObjectDefinitions
}

// BuildServiceDefinitionFromAZSchema converts authzed schema to shield service definition.
// This conversion is lossy, and it only keeps list of permissions used in the schema per resource
func BuildServiceDefinitionFromAZSchema(name string, azDefinitions []*azcore.NamespaceDefinition) (schema.ServiceDefinition, error) {
	resourcePermissionMap := map[string][]string{}
	// iterate over namespace to find services and permissions
	for _, def := range azDefinitions {
		nameParts := strings.Split(def.Name, "/")
		if len(nameParts) != 2 || nameParts[0] != name {
			// error if name is not in "namespace/resource" notation
			continue
		}
		resourcePermissionMap[nameParts[1]] = []string{}

		if def.Name == schema.RoleBindingNamespace {
			// build permission set for all namespaces using roles to bind themselves
			for _, rel := range def.Relation {
				if rel.UsersetRewrite != nil { // not nil for permissions in zed file
					permissionParts := strings.Split(rel.Name, "_")
					var service, resource, permission string
					switch len(permissionParts) {
					case 3:
						service, resource, permission = permissionParts[0], permissionParts[1], permissionParts[2]
					case 2:
						resource, permission = permissionParts[0], permissionParts[1]
					case 1:
						permission = permissionParts[0]
					}
					_ = service
					resourcePermissionMap[resource] = append(resourcePermissionMap[resource], permission)
				}
			}
		}
	}

	appService := schema.ServiceDefinition{
		Name: name,
	}
	for k, v := range resourcePermissionMap {
		defResource := schema.DefinitionResource{
			Name:        k,
			Permissions: nil,
		}
		for _, perm := range v {
			defResource.Permissions = append(defResource.Permissions, schema.ResourcePermission{
				Name: perm,
			})
		}
		appService.Resources = append(appService.Resources, defResource)
	}
	return appService, nil
}

// ApplyServiceDefinitionOverAZSchema applies the provided user defined service over existing schema
// and returns the updated schema
func ApplyServiceDefinitionOverAZSchema(serviceDef schema.ServiceDefinition, existingDefinitions []*azcore.NamespaceDefinition) ([]*azcore.NamespaceDefinition, error) {
	// keep relations/permissions required to be appended in existing definitions
	// this is required to hook user roles over application authz hierarchy
	var relationsForOrg []*azcore.Relation
	var relationsForProject []*azcore.Relation
	var relationsForRole []*azcore.Relation
	var relationsForRoleBinding []*azcore.Relation

	// prepare new definition with it own relations and permissions
	var userDefinedServices []*azcore.NamespaceDefinition
	for _, serviceResource := range serviceDef.Resources {
		var relationsForResource []*azcore.Relation
		for _, resourcePerms := range serviceResource.Permissions {
			permName := FQPermissionName(serviceDef.Name, serviceResource.Name, resourcePerms.Name)

			// create permissions
			{
				// for resource
				nsRel, err := aznamespace.Relation(resourcePerms.Name, aznamespace.Union(
					aznamespace.ComputedUserset("owner"),
					aznamespace.TupleToUserset("project", "app_project_administer"),
					aznamespace.TupleToUserset("project", permName),
					aznamespace.TupleToUserset("granted", permName),
				), nil)
				if err != nil {
					return nil, err
				}
				relationsForResource = append(relationsForResource, nsRel)
			}
			{
				// for org
				nsRel, err := aznamespace.Relation(permName, aznamespace.Union(
					aznamespace.ComputedUserset("owner"),
					aznamespace.TupleToUserset("platform", "superuser"),
					aznamespace.TupleToUserset("granted", "app_organization_administer"),
					aznamespace.TupleToUserset("granted", permName),
				), nil)
				if err != nil {
					return nil, err
				}
				relationsForOrg = append(relationsForOrg, nsRel)
			}
			{
				// for project
				nsRel, err := aznamespace.Relation(permName, aznamespace.Union(
					aznamespace.TupleToUserset("org", permName),
					aznamespace.TupleToUserset("granted", "app_project_administer"),
					aznamespace.TupleToUserset("granted", permName),
				), nil)
				if err != nil {
					return nil, err
				}
				relationsForProject = append(relationsForProject, nsRel)
			}
			{
				// for rolebinding
				nsRel, err := aznamespace.Relation(permName, aznamespace.Intersection(
					aznamespace.ComputedUserset("bearer"),
					aznamespace.TupleToUserset("role", permName),
				), nil)
				if err != nil {
					return nil, err
				}
				relationsForRoleBinding = append(relationsForRoleBinding, nsRel)
			}
			{
				// for role
				nsRel, err := aznamespace.Relation(permName, nil,
					aznamespace.AllowedPublicNamespace(schema.UserPrincipal),
					aznamespace.AllowedPublicNamespace(schema.ServiceUserPrincipal),
				)
				if err != nil {
					return nil, err
				}
				relationsForRole = append(relationsForRole, nsRel)
			}
		}

		// prepare an owner relation
		// either we can attach each user who creates the resource with owner relation or
		// create an owner role and assign it to the user when the resource is created
		relationsForResource = append(relationsForResource, aznamespace.MustRelation("owner", nil, aznamespace.AllowedRelation(schema.UserPrincipal, generator.Ellipsis)))
		// attach service to project
		relationsForResource = append(relationsForResource, aznamespace.MustRelation("project", nil, aznamespace.AllowedRelation(schema.ProjectNamespace, generator.Ellipsis)))
		// attach role binding to service
		relationsForResource = append(relationsForResource, aznamespace.MustRelation("granted", nil, aznamespace.AllowedRelation(schema.RoleBindingNamespace, generator.Ellipsis)))

		// prepare namespace
		resourceDef := aznamespace.Namespace(fmt.Sprintf("%s/%s", serviceDef.Name, serviceResource.Name), relationsForResource...)
		userDefinedServices = append(userDefinedServices, resourceDef)
	}

	// append new definition to existing definitions
	existingDefinitions = append(existingDefinitions, userDefinedServices...)

	if len(relationsForOrg) > 0 {
		for _, baseDef := range existingDefinitions {
			var err error

			switch baseDef.Name {
			case schema.OrganizationNamespace:
				// populate app/organization with service permissions to allow bounding service roles at org level

				// add comment for the top relation
				relationsForOrg[0].Metadata, err = aznamespace.AddComment(nil, fmt.Sprintf("// auto-generated for %s", serviceDef.Name))
				if err != nil {
					return nil, err
				}
				baseDef.Relation = append(baseDef.Relation, relationsForOrg...)
			case schema.ProjectNamespace:
				// populate app/project with service permissions to allow bounding service roles at project level
				relationsForProject[0].Metadata, err = aznamespace.AddComment(relationsForProject[0].Metadata, fmt.Sprintf("// auto-generated for %s", serviceDef.Name))
				if err != nil {
					return nil, err
				}
				baseDef.Relation = append(baseDef.Relation, relationsForProject...)
			case schema.RoleBindingNamespace:
				// populate app/rolebinding with service relations to allow checking service roles with permissions
				relationsForRoleBinding[0].Metadata, err = aznamespace.AddComment(relationsForRoleBinding[0].Metadata, fmt.Sprintf("// auto-generated for %s", serviceDef.Name))
				if err != nil {
					return nil, err
				}
				baseDef.Relation = append(baseDef.Relation, relationsForRoleBinding...)
			case schema.RoleNamespace:
				// populate app/role with service permissions to allow building service roles with permissions
				relationsForRole[0].Metadata, err = aznamespace.AddComment(relationsForRole[0].Metadata, fmt.Sprintf("// auto-generated for %s", serviceDef.Name))
				if err != nil {
					return nil, err
				}
				baseDef.Relation = append(baseDef.Relation, relationsForRole...)
			}
		}
	}

	return existingDefinitions, nil
}

func FQPermissionName(service, resource, verb string) string {
	return fmt.Sprintf("%s_%s_%s", service, resource, verb)
}
