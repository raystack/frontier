package body_extractor

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func makeReaderCloser(s string) *io.ReadCloser {
	readCloser := ioutil.NopCloser(strings.NewReader(s))
	return &readCloser
}

func TestJSONPayloadHandler_Extract(t *testing.T) {
	type args struct {
		body *io.ReadCloser
		key  string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Test JSON payload",
			args: args{
				body: makeReaderCloser(`{"key":"value"}`),
				key:  "key",
			},
			want:    "value",
			wantErr: false,
		},
		{
			name: "Test JSON payload with missing key",
			args: args{
				body: makeReaderCloser(`{"key":"value"}`),
				key:  "missing_key",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test JSON payload with invalid JSON",
			args: args{
				body: makeReaderCloser(`{"key":"value`),
				key:  "key",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := JSONPayloadHandler{}
			got, err := h.Extract(tt.args.body, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extract() got = %v, want %v", got, tt.want)
			}
		})
	}
}
