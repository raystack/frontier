package blob

import (
	"context"
	"testing"

	"github.com/odpf/shield/internal/schema"

	"github.com/stretchr/testify/assert"
)

func TestA(t *testing.T) {
	testBucket, err := NewStore(context.Background(), "file://testdata", "")
	assert.NoError(t, err)

	s := SchemaConfig{
		bucket: testBucket,
	}

	config, err := s.GetSchema(context.Background())
	assert.NoError(t, err)

	expectedMap := schema.NamespaceConfigMapType{
		"compute/engine": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles:               map[string][]string{},
			Permissions:         map[string][]string{},
		},
		"compute/network": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles:               map[string][]string{},
			Permissions:         map[string][]string{},
		},
		"entropy/dagger": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"database_editor": {"team"},
				"viewer":          {"user"},
			},
			Permissions: map[string][]string{
				"database_edit": {
					"owner",
					"organization/sink_editor",
					"database_editor",
				},
			},
		},
		"entropy/firehose": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"sink/editor": {
					"user",
					"team",
				},
				"viewer": {
					"user",
					"team",
				},
			},
			Permissions: map[string][]string{
				"sink_edit": {
					"owner",
					"sink_editor",
					"organization/sink_editor",
				},
				"view": {
					"owner",
					"organization/owner",
					"viewer",
				},
			},
		},
		"guardian/appeal": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"remover": {
					"user",
				},
				"viewer": {
					"user",
					"team",
				},
			},
			Permissions: map[string][]string{
				"delete": {
					"remover",
					"organization/appleal_owner",
				},
				"view": {
					"owner",
					"organization/owner",
					"viewer",
				},
			},
		},
		"organization": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"appleal_owner": {
					"user",
					"team",
				},
				"database_editor": {
					"team",
				},
				"sink_editor": {
					"user",
					"team",
				},
			},
			Permissions: map[string][]string{},
		},
		"project": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"owner": {
					"group",
				},
				"viewer": {
					"user",
					"team",
				},
			},
			Permissions: map[string][]string{},
		},
	}

	assert.Equal(t, expectedMap, config)
}
