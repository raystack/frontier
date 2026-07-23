package webhook_test

import (
	"context"
	"testing"

	"github.com/raystack/frontier/core/webhook"
	"github.com/stretchr/testify/assert"
)

// fakeEndpointRepo is an in-memory webhook.EndpointRepository for service tests.
type fakeEndpointRepo struct {
	items []webhook.Endpoint
}

func (f *fakeEndpointRepo) Create(_ context.Context, e webhook.Endpoint) (webhook.Endpoint, error) {
	f.items = append(f.items, e)
	return e, nil
}

func (f *fakeEndpointRepo) UpdateByID(_ context.Context, e webhook.Endpoint) (webhook.Endpoint, error) {
	for i := range f.items {
		if f.items[i].ID == e.ID {
			f.items[i] = e
			return e, nil
		}
	}
	return webhook.Endpoint{}, webhook.ErrNotFound
}

func (f *fakeEndpointRepo) Delete(_ context.Context, _ string) error { return nil }

func (f *fakeEndpointRepo) List(_ context.Context, _ webhook.EndpointFilter) ([]webhook.Endpoint, error) {
	return f.items, nil
}

func TestServiceCreateEndpointValidation(t *testing.T) {
	t.Run("rejects a non-absolute url", func(t *testing.T) {
		s := webhook.NewService(&fakeEndpointRepo{})
		_, err := s.CreateEndpoint(context.Background(), webhook.Endpoint{URL: "not-a-url"})
		assert.ErrorIs(t, err, webhook.ErrInvalidDetail)
	})

	t.Run("rejects an empty url", func(t *testing.T) {
		s := webhook.NewService(&fakeEndpointRepo{})
		_, err := s.CreateEndpoint(context.Background(), webhook.Endpoint{URL: "   "})
		assert.ErrorIs(t, err, webhook.ErrInvalidDetail)
	})

	t.Run("rejects a non-http(s) scheme", func(t *testing.T) {
		s := webhook.NewService(&fakeEndpointRepo{})
		_, err := s.CreateEndpoint(context.Background(), webhook.Endpoint{URL: "ftp://a.example/hook"})
		assert.ErrorIs(t, err, webhook.ErrInvalidDetail)
	})

	t.Run("rejects an http(s) url with no host", func(t *testing.T) {
		// "https://" and "http:///path" parse as absolute http(s) URLs with an
		// empty host, but a webhook with no host to deliver to is useless.
		s := webhook.NewService(&fakeEndpointRepo{})
		_, err := s.CreateEndpoint(context.Background(), webhook.Endpoint{URL: "https://"})
		assert.ErrorIs(t, err, webhook.ErrInvalidDetail)
	})

	t.Run("rejects an unknown state", func(t *testing.T) {
		s := webhook.NewService(&fakeEndpointRepo{})
		_, err := s.CreateEndpoint(context.Background(), webhook.Endpoint{URL: "https://a.example/hook", State: "paused"})
		assert.ErrorIs(t, err, webhook.ErrInvalidDetail)
	})

	t.Run("rejects a url another endpoint already uses", func(t *testing.T) {
		repo := &fakeEndpointRepo{items: []webhook.Endpoint{{ID: "e1", URL: "https://a.example/hook"}}}
		_, err := webhook.NewService(repo).CreateEndpoint(context.Background(), webhook.Endpoint{URL: "https://a.example/hook"})
		assert.ErrorIs(t, err, webhook.ErrConflict)
	})

	t.Run("creates a valid endpoint, defaulting state and generating a secret", func(t *testing.T) {
		got, err := webhook.NewService(&fakeEndpointRepo{}).CreateEndpoint(
			context.Background(), webhook.Endpoint{URL: "https://a.example/hook"})
		assert.NoError(t, err)
		assert.Equal(t, webhook.Enabled, got.State) // defaulted
		assert.Len(t, got.Secrets, 1)               // server-generated signing secret
	})
}

func TestServiceUpdateEndpointURLUniqueness(t *testing.T) {
	t.Run("rejects a url used by a different endpoint", func(t *testing.T) {
		repo := &fakeEndpointRepo{items: []webhook.Endpoint{
			{ID: "e1", URL: "https://a.example/hook", State: webhook.Enabled},
			{ID: "e2", URL: "https://b.example/hook", State: webhook.Enabled},
		}}
		_, err := webhook.NewService(repo).UpdateEndpoint(context.Background(),
			webhook.Endpoint{ID: "e2", URL: "https://a.example/hook", State: webhook.Enabled})
		assert.ErrorIs(t, err, webhook.ErrConflict)
	})

	t.Run("lets an endpoint keep its own url", func(t *testing.T) {
		repo := &fakeEndpointRepo{items: []webhook.Endpoint{
			{ID: "e1", URL: "https://a.example/hook", State: webhook.Enabled},
		}}
		_, err := webhook.NewService(repo).UpdateEndpoint(context.Background(),
			webhook.Endpoint{ID: "e1", URL: "https://a.example/hook", State: webhook.Disabled})
		assert.NoError(t, err)
	})
}
