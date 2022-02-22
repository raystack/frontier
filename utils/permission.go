package utils

import (
	"fmt"

	"github.com/odpf/shield/model"
)

const NON_RESOURCE_ID = "*"

func CreateResourceId(resource model.Resource) string {
	if resource.Name == NON_RESOURCE_ID {
		return fmt.Sprintf("p/%s/%s", resource.ProjectId, resource.NamespaceId)
	}
	return fmt.Sprintf("r/%s/%s", resource.NamespaceId, resource.Name)
}

func CreateNamespaceID(backend, resourceType string) string {
	return fmt.Sprintf("%s_%s", backend, resourceType)
}

/*
 /project/uuid/
*/
