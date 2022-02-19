package middleware

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestEnrichPathParams(t *testing.T) {
	t.Parallel()
	type args struct {
		r      *http.Request
		params map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestEnrichPathParams",
			args: args{
				r: &http.Request{
					URL: &url.URL{
						Path: "/api/v1/users/{userId}/{userName}",
					},
				},
				params: map[string]string{
					"userId":   "1",
					"userName": "John",
				},
			},
		},
		{
			name: "TestEnrichPathParamsWithEmptyParams",
			args: args{
				r: &http.Request{
					URL: &url.URL{
						Path: "/api/v1/users/{userId}/{userName}",
					},
				},
				params: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnrichPathParams(tt.args.r, tt.args.params)
			if !reflect.DeepEqual((tt.args.r.Context().Value(ctxPathParamsKey)).(map[string]string), tt.args.params) {
				t.Errorf("EnrichPathParams() = %v, want %v", tt.args.r.Context().Value(ctxPathParamsKey), tt.args.params)
			}
		})
	}
}
