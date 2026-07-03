package cmd

import (
	"testing"

	"github.com/raystack/frontier/internal/reconcile"
	"github.com/stretchr/testify/assert"
)

func TestResolveKind(t *testing.T) {
	registry := map[string]reconcile.Reconciler{reconcile.KindPlatformUser: nil}

	for _, arg := range []string{"PlatformUser", "platformuser", "PLATFORMUSERS", "platformusers"} {
		kind, err := resolveKind(arg, registry)
		assert.NoError(t, err, arg)
		assert.Equal(t, reconcile.KindPlatformUser, kind, arg)
	}

	_, err := resolveKind("nope", registry)
	assert.ErrorContains(t, err, `unknown kind "nope" (available: PlatformUser)`)
}
