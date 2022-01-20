package utils

import (
	"fmt"

	"github.com/odpf/shield/model"
)

func CreateResourceId(resource model.Resource) string {
	return fmt.Sprintf("%s/%s", resource.NamespaceId, resource.Name)
}
