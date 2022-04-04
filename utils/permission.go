package utils

import (
	"fmt"

	"github.com/odpf/shield/internal/bootstrap/definition"

	"github.com/odpf/shield/model"
)

const NON_RESOURCE_ID = "*"

var systemNSIds = []string{definition.TeamNamespace.Id, definition.UserNamespace.Id, definition.OrgNamespace.Id, definition.ProjectNamespace.Id}

func StrListHas(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

/*
 /project/uuid/
*/
func CreateResourceURN(resource model.Resource) string {
	if ok := StrListHas(systemNSIds, resource.NamespaceId); ok {
		return resource.Name
	}
	if resource.Name == NON_RESOURCE_ID {
		return fmt.Sprintf("p/%s/%s", resource.ProjectId, resource.NamespaceId)
	}
	return fmt.Sprintf("r/%s/%s", resource.NamespaceId, resource.Name)
}

func CreateNamespaceID(backend, resourceType string) string {
	return fmt.Sprintf("%s_%s", backend, resourceType)
}
