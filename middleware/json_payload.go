package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/tidwall/gjson"
)

type JSONPayloadHandler struct{}

func (h JSONPayloadHandler) Extract(req *http.Request, key string) (string, error) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()
	// repopulate body
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	field := gjson.GetBytes(reqBody, key)
	if !field.Exists() {
		return "", errors.Errorf("failed to find field: %s", key)
	}
	return field.String(), nil
}
