package schema_generator

import (
	"sort"
	"strings"
	"testing"

	"github.com/odpf/shield/internal/schema"

	_ "embed"

	"github.com/stretchr/testify/assert"
)

//go:embed predefined_schema
var predefinedSchema string

func makeDefnMap(s []string) map[string][]string {
	finalMap := make(map[string][]string)

	for _, v := range s {
		splitedConfigText := strings.Split(v, "\n")
		k := splitedConfigText[0]
		sort.Strings(splitedConfigText)
		finalMap[k] = splitedConfigText
	}

	return finalMap
}

// Test to check difference between predefined_schema.txt and schema defined in predefined.go
func TestPredefinedSchema(t *testing.T) {
	// slice and sort as GenerateSchema() generated the permissions and relations in random order

	scm, err := GenerateSchema(schema.PreDefinedSystemNamespaceConfig)
	assert.Nil(t, err)
	actualPredefinedConfigs := makeDefnMap(scm)
	expectedPredefinedConfigs := makeDefnMap(strings.Split(predefinedSchema, "\n--\n"))
	assert.Equal(t, actualPredefinedConfigs, expectedPredefinedConfigs)
}
