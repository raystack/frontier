package body_extractor

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/tidwall/gjson"
)

type JSONPayloadHandler struct{}

func (h JSONPayloadHandler) Extract(body *io.ReadCloser, key string) (string, error) {
	reqBody, err := ioutil.ReadAll(*body)
	if err != nil {
		return "", err
	}
	defer (*body).Close()
	// repopulate body
	*body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	field := gjson.GetBytes(reqBody, key)
	if !field.Exists() {
		return "", errors.Errorf("failed to find field: %s", key)
	}
	return field.String(), nil
}
