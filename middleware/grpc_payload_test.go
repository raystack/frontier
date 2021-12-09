package middleware

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewQuery(t *testing.T) {
	som,  err := ParseQuery("1.2.3[1].45")
	fmt.Println(err)
	fmt.Println(prettyPrint(som))
	//fmt.Println(e)
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}