package reconcile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeReconciler struct{ called int }

func (f *fakeReconciler) Kind() string { return KindPlatformUser }
func (f *fakeReconciler) Reconcile(_ context.Context, _ []byte, _ bool) (Report, error) {
	f.called++
	return Report{Kind: KindPlatformUser}, nil
}

func TestRun_SpecHandling(t *testing.T) {
	t.Run("rejects a document missing its spec (not an authoritative empty state)", func(t *testing.T) {
		rec := &fakeReconciler{}
		reg := map[string]Reconciler{KindPlatformUser: rec}

		_, err := Run(context.Background(), reg, []byte("kind: PlatformUser\n"), false)
		assert.Error(t, err)
		assert.Zero(t, rec.called) // never dispatched
	})

	t.Run("dispatches a document that has a spec", func(t *testing.T) {
		rec := &fakeReconciler{}
		reg := map[string]Reconciler{KindPlatformUser: rec}

		_, err := Run(context.Background(), reg, []byte("kind: PlatformUser\nspec: []\n"), false)
		assert.NoError(t, err)
		assert.Equal(t, 1, rec.called)
	})
}
