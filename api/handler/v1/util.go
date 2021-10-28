package v1

import (
	"fmt"
	"strings"
)

func mapOfStringValues(m map[string]interface{}) (map[string]string, error) {
	newMap := make(map[string]string)

	for key, value := range m {
		switch value := value.(type) {
		case string:
			newMap[key] = value
		default:
			return map[string]string{}, fmt.Errorf("value for %s key is not string", key)
		}
	}

	return newMap, nil
}

func mapOfInterfaceValues(m map[string]string) map[string]interface{} {
	newMap := make(map[string]interface{})

	for key, value := range m {
		newMap[key] = value
	}

	return newMap
}

func generateSlug(name string) string {
	preProcessed := strings.ReplaceAll(strings.TrimSpace(strings.TrimSpace(name)), "_", "-")
	return strings.Join(
		strings.Split(preProcessed, " "),
		"-",
	)
}
