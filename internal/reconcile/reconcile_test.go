package reconcile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeReconciler struct{ called int }

func (f *fakeReconciler) Kind() string            { return KindPlatformUser }
func (f *fakeReconciler) Validate(_ []byte) error { return nil }
func (f *fakeReconciler) Reconcile(_ context.Context, _ []byte, _ bool) (Report, error) {
	f.called++
	return Report{Kind: KindPlatformUser}, nil
}
func (f *fakeReconciler) Export(_ context.Context) (any, error) { return []string{}, nil }

// partialReconciler fails after applying some operations, like a real apply that
// dies part-way through.
type partialReconciler struct{}

func (partialReconciler) Kind() string            { return KindPlatformUser }
func (partialReconciler) Validate(_ []byte) error { return nil }
func (partialReconciler) Reconcile(_ context.Context, _ []byte, _ bool) (Report, error) {
	return Report{Kind: KindPlatformUser, Applied: 2, Planned: []string{"add a", "add b", "add c"}},
		errors.New("apply failed on the third op")
}
func (partialReconciler) Export(_ context.Context) (any, error) { return []string{}, nil }

// validateFailReconciler rejects any spec at validation time and records whether
// it was ever dispatched.
type validateFailReconciler struct{ called int }

func (v *validateFailReconciler) Kind() string            { return KindPermission }
func (v *validateFailReconciler) Validate(_ []byte) error { return errors.New("bad entry") }
func (v *validateFailReconciler) Reconcile(_ context.Context, _ []byte, _ bool) (Report, error) {
	v.called++
	return Report{Kind: KindPermission}, nil
}
func (v *validateFailReconciler) Export(_ context.Context) (any, error) { return []string{}, nil }

func TestRun_SpecHandling(t *testing.T) {
	t.Run("rejects a document missing its spec (not an empty list)", func(t *testing.T) {
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

	t.Run("accepts apiVersion v1 and rejects unknown versions", func(t *testing.T) {
		rec := &fakeReconciler{}
		reg := map[string]Reconciler{KindPlatformUser: rec}

		_, err := Run(context.Background(), reg, []byte("apiVersion: v1\nkind: PlatformUser\nspec: []\n"), false)
		assert.NoError(t, err)
		assert.Equal(t, 1, rec.called)

		_, err = Run(context.Background(), reg, []byte("apiVersion: v2\nkind: PlatformUser\nspec: []\n"), false)
		assert.ErrorContains(t, err, `unsupported apiVersion "v2"`)
	})

	t.Run("a bad later document stops the run before anything applies", func(t *testing.T) {
		rec := &fakeReconciler{}
		reg := map[string]Reconciler{KindPlatformUser: rec}
		data := []byte("kind: PlatformUser\nspec: []\n---\nkind: Unknown\nspec: []\n")

		_, err := Run(context.Background(), reg, data, false)
		assert.ErrorContains(t, err, `no reconciler registered for kind "Unknown"`)
		assert.Zero(t, rec.called) // the whole file is checked before any dispatch
	})

	t.Run("a document that fails validation blocks every apply", func(t *testing.T) {
		first := &fakeReconciler{}          // a valid document that records applies
		second := &validateFailReconciler{} // a later document that fails validation
		reg := map[string]Reconciler{KindPlatformUser: first, KindPermission: second}
		data := []byte("kind: PlatformUser\nspec: []\n---\nkind: Permission\nspec: []\n")

		_, err := Run(context.Background(), reg, data, false)
		assert.ErrorContains(t, err, "validate Permission")
		assert.Zero(t, first.called) // validation runs for every document before any apply
		assert.Zero(t, second.called)
	})

	t.Run("a document with content but no kind is rejected", func(t *testing.T) {
		rec := &fakeReconciler{}
		reg := map[string]Reconciler{KindPlatformUser: rec}

		_, err := Run(context.Background(), reg, []byte("spec:\n  - type: user\n"), false)
		assert.ErrorContains(t, err, "content but no kind")
		assert.Zero(t, rec.called)
	})

	t.Run("a blank trailing document is skipped, not rejected", func(t *testing.T) {
		rec := &fakeReconciler{}
		reg := map[string]Reconciler{KindPlatformUser: rec}

		_, err := Run(context.Background(), reg, []byte("kind: PlatformUser\nspec: []\n---\n"), false)
		assert.NoError(t, err)
		assert.Equal(t, 1, rec.called)
	})
}

func TestExport_Errors(t *testing.T) {
	t.Run("unknown kind", func(t *testing.T) {
		_, err := Export(context.Background(), map[string]Reconciler{}, "Nope")
		assert.ErrorContains(t, err, `no reconciler registered for kind "Nope"`)
	})
}

func TestRun_ReturnsPartialReportOnError(t *testing.T) {
	reg := map[string]Reconciler{KindPlatformUser: partialReconciler{}}

	reports, err := Run(context.Background(), reg, []byte("kind: PlatformUser\nspec: []\n"), false)

	// The run stops on the error, but the failing document's report is still
	// returned so the caller can see what was applied before the failure.
	assert.Error(t, err)
	if assert.Len(t, reports, 1) {
		assert.Equal(t, 2, reports[0].Applied)
		assert.Len(t, reports[0].Planned, 3)
	}
}
