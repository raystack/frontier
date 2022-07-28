// TODO: @krtkvrm | remove this file before merging

package postgres

import (
	"fmt"
	"testing"
)

func TestGetQuery(t *testing.T) {
	fmt.Println(buildCreateNamespaceQuery(dialect))
}
