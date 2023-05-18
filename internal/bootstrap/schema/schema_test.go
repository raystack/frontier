package schema_test

import (
	"testing"

	"github.com/odpf/shield/internal/bootstrap/schema"
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
