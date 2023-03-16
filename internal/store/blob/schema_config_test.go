package blob

import (
	"context"
	"testing"

	"github.com/goto/shield/internal/schema"

	"github.com/stretchr/testify/assert"
)

func TestGetSchema(t *testing.T) {
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
				"database_editor": {schema.GroupPrincipal},
				"viewer":          {schema.UserPrincipal},
			},
			Permissions: map[string][]string{
				"database_edit": {
					"owner",
					"organization:sink_editor",
					"database_editor",
				},
			},
			Type: schema.ResourceGroupNamespace,
		},
		"entropy/firehose": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"sink_editor": {
					schema.UserPrincipal,
					schema.GroupPrincipal,
				},
				"viewer": {
					schema.UserPrincipal,
					schema.GroupPrincipal,
				},
			},
			Permissions: map[string][]string{
				"sink_edit": {
					"owner",
					"sink_editor",
					"organization:sink_editor",
				},
				"view": {
					"owner",
					"organization:owner",
					"viewer",
				},
			},
			Type: schema.ResourceGroupNamespace,
		},
		"guardian/appeal": schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"remover": {
					schema.UserPrincipal,
				},
				"viewer": {
					schema.UserPrincipal,
					schema.GroupPrincipal,
				},
			},
			Permissions: map[string][]string{
				"delete": {
					"remover",
					"organization:appleal_owner",
				},
				"view": {
					"owner",
					"organization:owner",
					"viewer",
				},
			},
			Type: schema.ResourceGroupNamespace,
		},
		schema.OrganizationNamespace: schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"appleal_owner": {
					schema.UserPrincipal,
					schema.GroupPrincipal,
				},
				"database_editor": {
					schema.GroupPrincipal,
				},
				"sink_editor": {
					schema.UserPrincipal,
					schema.GroupPrincipal,
				},
			},
			Permissions: map[string][]string{},
			Type:        schema.SystemNamespace,
		},
		schema.ProjectNamespace: schema.NamespaceConfig{
			InheritedNamespaces: nil,
			Roles: map[string][]string{
				"owner": {
					schema.GroupPrincipal,
				},
				"viewer": {
					schema.UserPrincipal,
					schema.GroupPrincipal,
				},
			},
			Permissions: map[string][]string{},
			Type:        schema.SystemNamespace,
		},
	}

	assert.Equal(t, expectedMap, config)
}
