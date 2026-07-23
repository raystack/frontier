package schema_test

import (
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

func TestFQPermissionNameFromNamespace(t *testing.T) {
	type args struct {
		namespace string
		verb      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic namespace and verb",
			args: args{
				namespace: "app/user",
				verb:      "delete",
			},
			want: "app_user_delete",
		},
		{
			name: "namespace using alias",
			args: args{
				namespace: "user",
				verb:      "delete",
			},
			want: "app_user_delete",
		},
		{
			name: "namespace without resource",
			args: args{
				namespace: "hello",
				verb:      "delete",
			},
			want: "hello_default_delete",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := schema.FQPermissionNameFromNamespace(tt.args.namespace, tt.args.verb); got != tt.want {
				t.Errorf("FQPermissionNameFromNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPermissionNamespace(t *testing.T) {
	tests := []struct {
		ns   string
		want bool
	}{
		{"resource/aoi", true},
		{"user/project", true},
		{"org/user", true},
		{"compute/disk", true},
		{"a1/b2", true},
		{"resource_order/item", false}, // underscore in a part collides on the slug
		{"resource/order_item", false},
		{"Compute/order", false}, // uppercase is not a SpiceDB object type
		{"compute/Order", false},
		{"compute", false},    // one part
		{"a/b/c", false},      // three parts
		{"/x", false},         // empty part
		{"x/", false},         // empty part
		{"", false},           // empty
		{"comp-ute/x", false}, // hyphen is not alphanumeric
	}
	for _, tt := range tests {
		t.Run(tt.ns, func(t *testing.T) {
			if got := schema.IsValidPermissionNamespace(tt.ns); got != tt.want {
				t.Errorf("IsValidPermissionNamespace(%q) = %v, want %v", tt.ns, got, tt.want)
			}
		})
	}
}

func TestIsBootstrapServiceUser(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"canonical id", schema.BootstrapServiceUserID, true},
		{"braces form", "{00000000-0000-0000-0000-000000000001}", true},
		{"urn form", "urn:uuid:00000000-0000-0000-0000-000000000001", true},
		{"no dashes", "00000000000000000000000000000001", true},
		{"surrounding space", "  00000000-0000-0000-0000-000000000001  ", true},
		{"different uuid", "00000000-0000-0000-0000-000000000002", false},
		{"not a uuid", "alice@x.com", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := schema.IsBootstrapServiceUser(tt.id); got != tt.want {
				t.Errorf("IsBootstrapServiceUser(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}
