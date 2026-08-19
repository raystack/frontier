package schema_test

import (
	"strings"
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

func TestPermissionNamespaceAndNameFromKey(t *testing.T) {
	tests := []struct {
		key           string
		wantNamespace string
		wantName      string
	}{
		{"compute.instance.delete", "compute/instance", "delete"},
		{"app.organization.get", "app/organization", "get"},
		// dots after the namespace belong to the name
		{"compute.instance.soft.delete", "compute/instance", "soft.delete"},
		{"compute.instance", "", ""}, // too few parts
		{"compute", "", ""},
		{"", "", ""},
		{".instance.delete", "", ""},  // empty service
		{"compute..delete", "", ""},   // empty resource
		{"compute.instance.", "", ""}, // empty name
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			ns, name := schema.PermissionNamespaceAndNameFromKey(tt.key)
			if ns != tt.wantNamespace || name != tt.wantName {
				t.Errorf("PermissionNamespaceAndNameFromKey(%q) = (%q, %q), want (%q, %q)", tt.key, ns, name, tt.wantNamespace, tt.wantName)
			}
		})
	}
}

func TestPermissionKeyRoundTrip(t *testing.T) {
	tests := []struct {
		namespace string
		name      string
	}{
		{"compute/instance", "delete"},
		{"app/organization", "get"},
		{"database/instance", "soft.delete"},
	}
	for _, tt := range tests {
		t.Run(tt.namespace+":"+tt.name, func(t *testing.T) {
			key := schema.PermissionKeyFromNamespaceAndName(tt.namespace, tt.name)
			ns, name := schema.PermissionNamespaceAndNameFromKey(key)
			if ns != tt.namespace || name != tt.name {
				t.Errorf("round trip through key %q = (%q, %q), want (%q, %q)", key, ns, name, tt.namespace, tt.name)
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
		{"a1/b2", false},               // parts shorter than 3 are not valid SpiceDB object types
		{"ab/order", false},            // service part shorter than 3
		{"compute/ab", false},          // resource part shorter than 3
		{"1compute/order", false},      // a part cannot start with a digit
		{"compute/2order", false},      // a part cannot start with a digit
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

func TestIsValidPermissionName(t *testing.T) {
	// A permission's verb becomes a SpiceDB relation name directly, so it must
	// satisfy SpiceDB's relation grammar: start with a lowercase letter, then
	// lowercase alphanumerics, at least three characters. Underscore is left out
	// because the slug joins service, resource, and verb with "_".
	tests := []struct {
		name string
		want bool
	}{
		{"get", true},
		{"create", true},
		{"list", true},
		{"abc", true}, // exactly three characters
		{"id", false}, // shorter than three characters
		{"ab", false},
		{"Get", false},     // uppercase is not a SpiceDB relation
		{"getAll", false},  // uppercase anywhere
		{"1get", false},    // cannot start with a digit
		{"get_all", false}, // underscore is not allowed in a verb
		{"get-all", false}, // hyphen is not alphanumeric
		{"", false},        // empty
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := schema.IsValidPermissionName(tt.name); got != tt.want {
				t.Errorf("IsValidPermissionName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPermissionSlugWithinLimit(t *testing.T) {
	// The slug service_resource_verb becomes a SpiceDB relation, which is capped at
	// sixty-four characters. Each part is valid on its own, but together they can
	// overflow, so a plan can pass and the schema still fail to compile.
	tests := []struct {
		ns, name string
		want     bool
	}{
		{"compute/order", "get", true},
		// 30 + 29 + 3 plus the two underscores is exactly sixty-four
		{strings.Repeat("a", 30) + "/" + strings.Repeat("b", 29), "get", true},
		// one more character tips it over the limit
		{strings.Repeat("a", 30) + "/" + strings.Repeat("b", 30), "get", false},
	}
	for _, tt := range tests {
		t.Run(tt.ns+"."+tt.name, func(t *testing.T) {
			if got := schema.PermissionSlugWithinLimit(tt.ns, tt.name); got != tt.want {
				t.Errorf("PermissionSlugWithinLimit(%q, %q) = %v, want %v", tt.ns, tt.name, got, tt.want)
			}
		})
	}
}

func TestIsReservedPermissionVerb(t *testing.T) {
	// These are the relations the generator adds to every custom resource, so a
	// verb equal to one of them would declare the same relation twice.
	tests := []struct {
		name string
		want bool
	}{
		{"owner", true},
		{"project", true},
		{"granted", true},
		{"Owner", true}, // reserved check is case-insensitive
		{"get", false},
		{"delete", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := schema.IsReservedPermissionVerb(tt.name); got != tt.want {
				t.Errorf("IsReservedPermissionVerb(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestValidateCustomPermission(t *testing.T) {
	tests := []struct {
		title   string
		ns      string
		name    string
		wantErr string // substring the error must contain; empty means no error
	}{
		{"valid", "compute/order", "get", ""},
		{"uppercase verb", "compute/order", "Read", "invalid permission verb"},
		{"short verb", "compute/order", "id", "invalid permission verb"},
		{"reserved verb owner", "compute/order", "owner", "reserved"},
		{"reserved verb granted", "compute/order", "granted", "reserved"},
		{"underscore namespace", "res_order/item", "get", "invalid permission namespace"},
		{"missing resource", "compute", "get", "invalid permission namespace"},
		{"slug too long", strings.Repeat("a", 30) + "/" + strings.Repeat("b", 30), "get", "too long"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			err := schema.ValidateCustomPermission(tt.ns, tt.name)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateCustomPermission(%q, %q) = %v, want nil", tt.ns, tt.name, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateCustomPermission(%q, %q) = %v, want error containing %q", tt.ns, tt.name, err, tt.wantErr)
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
