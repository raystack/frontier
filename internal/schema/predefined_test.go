package schema

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	content, err := ioutil.ReadFile("predefined_schema.txt")
	assert.NoError(t, err)

	// slice and sort as GenerateSchema() generated the permissions and relations in random order
	actualPredefinedConfigs := makeDefnMap(GenerateSchema(PreDefinedSystemNamespaceConfig))
	expectedPredefinedConfigs := makeDefnMap(strings.Split(string(content), "\n--\n"))
	assert.Equal(t, actualPredefinedConfigs, expectedPredefinedConfigs)
}
