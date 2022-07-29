// TODO: @krtkvrm | remove this file before merging

package postgres

import (
	"fmt"
	"testing"
)

func TestGetQuery(t *testing.T) {
	fmt.Println(buildGetResourcesByNamespaceQuery(dialect, true))
}

// SELECT id, namespace_id from resouces where id=$1 AND namespace_id IN (SELECT id from namespaces where backend=$2 AND resource_type=$3);
