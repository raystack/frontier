package interceptors

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"io"
	"net/http"
	"reflect"
	"strings"
)

const (
	// RawBytesMIME is the MIME type for raw bytes
	RawBytesMIME = "application/raw-bytes"
)

var (
	RawPayloadEndpoints = []string{
		"billing/webhooks/callback",
	}
	typeOfBytes = reflect.TypeOf([]byte(nil))
)

// RawJSONPb is a custom runtime.JSONPb that reads the raw JSON data as is.
type RawJSONPb struct {
	*runtime.JSONPb
}

func (*RawJSONPb) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return fmt.Errorf("%T is not a pointer", v)
		}

		rv = rv.Elem()
		if rv.Type() != typeOfBytes {
			return fmt.Errorf("type must be []byte but got %T", v)
		}

		rv.Set(reflect.ValueOf(data))
		return nil
	})
}

// ByteMimeWrapper is a custom runtime.ServeMuxOption that sets the content type to application/raw-bytes
// for specific endpoints. This is useful when the response is a raw byte array and should be serialized as is.
func ByteMimeWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, endpoint := range RawPayloadEndpoints {
			if strings.Contains(r.URL.Path, endpoint) {
				r.Header.Set("Content-Type", RawBytesMIME)
				break
			}
		}
		h.ServeHTTP(w, r)
	})
}
