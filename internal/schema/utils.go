package schema

import "fmt"

func SpiceDBPermissionInheritanceFormatter(permissionName, namespaceName string) string {
	return fmt.Sprintf("%s->%s", permissionName, namespaceName)
}
