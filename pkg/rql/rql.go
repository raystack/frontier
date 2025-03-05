package rql

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const TAG = "rql"

func GetDataTypeOfField(fieldName string, checkAgainst interface{}) (string, error) {
	val := reflect.ValueOf(checkAgainst)
	filterIdx := searchKeyInsideStruct(fieldName, val)
	if filterIdx < 0 {
		return "", fmt.Errorf("'%s' is not a valid filter key", fieldName)
	}
	structKeyTag := val.Type().Field(filterIdx).Tag.Get(TAG)
	return getDataTypeOfField(structKeyTag), nil
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func searchKeyInsideStruct(keyName string, val reflect.Value) int {
	for i := 0; i < val.NumField(); i++ {
		if strings.ToLower(val.Type().Field(i).Name) == strings.ToLower(clearString(keyName)) {
			return i
		}
	}
	return -1
}

func getDataTypeOfField(tagString string) string {
	res := "string"
	splitted := strings.Split(tagString, ",")
	for _, item := range splitted {
		kvSplitted := strings.Split(item, "=")
		if len(kvSplitted) == 2 {
			if kvSplitted[0] == "type" {
				return kvSplitted[1]
			}
		}
	}
	//fallback to string if type not found in tag value
	return res
}
