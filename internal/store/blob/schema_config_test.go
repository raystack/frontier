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
			Type:                schema.ResourceGroupNamespace,
		},
		"compute/network": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles:               map[string][]string{},
			Permissions:         map[string][]string{},
			Type:                schema.ResourceGroupNamespace,
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
			Type: schema.ResourceGroupNamespace,
		},
		"entropy/firehose": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"sink_editor": {
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
			Type: schema.ResourceGroupNamespace,
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
			Type: schema.ResourceGroupNamespace,
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
			Type:        schema.SystemNamespace,
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
			Type:        schema.SystemNamespace,
		},
	}

	assert.Equal(t, expectedMap, config)
}
