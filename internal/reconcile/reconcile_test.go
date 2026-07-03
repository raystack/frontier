package reconcile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeReconciler struct{ called int }

func (f *fakeReconciler) Kind() string { return KindPlatformUser }
func (f *fakeReconciler) Reconcile(_ context.Context, _ []byte, _ bool) (Report, error) {
	f.called++
	return Report{Kind: KindPlatformUser}, nil
}

// partialReconciler fails after applying some operations, like a real apply that
// dies part-way through.
type partialReconciler struct{}

func (partialReconciler) Kind() string { return KindPlatformUser }
func (partialReconciler) Reconcile(_ context.Context, _ []byte, _ bool) (Report, error) {
	return Report{Kind: KindPlatformUser, Applied: 2, Planned: []string{"add a", "add b", "add c"}},
		errors.New("apply failed on the third op")
}

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
}

func TestExport_Errors(t *testing.T) {
	t.Run("unknown kind", func(t *testing.T) {
		_, err := Export(context.Background(), map[string]Reconciler{}, "Nope")
		assert.ErrorContains(t, err, `no reconciler registered for kind "Nope"`)
	})

	t.Run("kind without export support", func(t *testing.T) {
		reg := map[string]Reconciler{KindPlatformUser: &fakeReconciler{}}
		_, err := Export(context.Background(), reg, KindPlatformUser)
		assert.ErrorContains(t, err, `does not support export`)
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
