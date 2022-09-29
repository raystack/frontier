package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/shield/core/role"
)

func PermissionInheritanceFormatter(permissionName, namespaceName string) string {
	return fmt.Sprintf("%s/%s", permissionName, namespaceName)
}

func SpiceDBPermissionInheritanceFormatter(roleName string) string {
	return strings.Replace(roleName, "/", "->", 1)
}

func getRoleAndPrincipal(roleName, namespaceId string) (role.Role, error) {
	splittedString := strings.Split(roleName, "/")
	if len(splittedString) == 1 {
		return role.Role{ID: roleName, NamespaceID: namespaceId}, nil
	} else if len(splittedString) >= 3 || len(splittedString) <= 0 {
		return role.Role{}, errors.New("wrong role format")
	}

	return role.Role{ID: splittedString[1], NamespaceID: splittedString[0]}, nil
}

func AppendIfUnique[T comparable](slice1 []T, slice2 []T) []T {
	for _, i := range slice2 {
		if !Contains(slice1, i) {
			slice1 = append(slice1, i)
		}
	}

	return slice1
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
