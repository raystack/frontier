package pipeline

import (
	"errors"
	"github.com/odpf/shield/structs"
	"net/http"
)

var (
	ErrAuthFails = errors.New("authorization fails")
	DebugMatchString = "shield"
)

type BasicHeaderAuthorizer struct {}

func (b BasicHeaderAuthorizer) Do(req *http.Request, rule *structs.Rule) error {
	// TODO this is just a test
	if req.Header.Get("X-user") != DebugMatchString {
		return ErrAuthFails
	}
	return nil
}

type BasicPathAuthorizer struct {}
func (b BasicPathAuthorizer) Do(req *http.Request, rule *structs.Rule) error {
	return nil
}

type BasicJSONPayloadAuthorizer struct {}
func (b BasicJSONPayloadAuthorizer) Do(req *http.Request, rule *structs.Rule) error {
	return nil
}
