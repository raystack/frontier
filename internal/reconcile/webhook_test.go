package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func cw(id, url, desc, state string, events ...string) currentWebhook {
	return currentWebhook{ID: id, URL: url, Description: desc, State: state, SubscribedEvents: events}
}

func TestDiffWebhooks(t *testing.T) {
	t.Run("adds a new endpoint", func(t *testing.T) {
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://a.example/hook", SubscribedEvents: []string{"org.created"}}},
			nil,
		)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opAdd, ops[0].action)
			assert.Equal(t, "add webhook https://a.example/hook [org.created]", ops[0].String())
		}
	})

	t.Run("updates only the changed managed fields, matched by url", func(t *testing.T) {
		current := []currentWebhook{cw("w1", "https://a.example/hook", "old", webhookStateEnabled, "org.created")}
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://a.example/hook", Description: "new", SubscribedEvents: []string{"org.deleted", "org.created"}}},
			current,
		)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opUpdate, ops[0].action)
			assert.Equal(t, "w1", ops[0].id)
			assert.Equal(t, "update webhook https://a.example/hook (description, subscribed_events)", ops[0].String())
			assert.Equal(t, "new", ops[0].spec.Description)
			assert.Equal(t, []string{"org.created", "org.deleted"}, ops[0].spec.SubscribedEvents)
		}
	})

	t.Run("a converged endpoint plans no change", func(t *testing.T) {
		current := []currentWebhook{cw("w1", "https://a.example/hook", "desc", webhookStateEnabled, "org.created")}
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://a.example/hook", Description: "desc", SubscribedEvents: []string{"org.created"}}},
			current,
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an omitted state keeps the server value", func(t *testing.T) {
		current := []currentWebhook{cw("w1", "https://a.example/hook", "desc", webhookStateDisabled, "org.created")}
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://a.example/hook", Description: "desc", SubscribedEvents: []string{"org.created"}}},
			current,
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("state change is planned when listed and different", func(t *testing.T) {
		current := []currentWebhook{cw("w1", "https://a.example/hook", "desc", webhookStateEnabled, "org.created")}
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://a.example/hook", Description: "desc", SubscribedEvents: []string{"org.created"}, State: webhookStateDisabled}},
			current,
		)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update webhook https://a.example/hook (state)", ops[0].String())
			assert.Equal(t, webhookStateDisabled, ops[0].spec.State)
		}
	})

	t.Run("delete removes a listed endpoint", func(t *testing.T) {
		current := []currentWebhook{cw("w1", "https://a.example/hook", "", webhookStateEnabled, "org.created")}
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://a.example/hook", Delete: true}},
			current,
		)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opRemove, ops[0].action)
			assert.Equal(t, "w1", ops[0].id)
			assert.Equal(t, "delete webhook https://a.example/hook", ops[0].String())
		}
	})

	t.Run("deleting an endpoint that is not on the server is a no-op", func(t *testing.T) {
		ops, err := diffWebhooks(
			[]WebhookSpec{{URL: "https://gone.example/hook", Delete: true}},
			nil,
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an endpoint on the server but not in the file fails the plan", func(t *testing.T) {
		current := []currentWebhook{cw("w1", "https://a.example/hook", "", webhookStateEnabled, "org.created")}
		_, err := diffWebhooks([]WebhookSpec{}, current)
		assert.ErrorContains(t, err, "webhooks exist on the server but are not in the file")
		assert.ErrorContains(t, err, "https://a.example/hook")
	})

	t.Run("adds and updates run before removes", func(t *testing.T) {
		current := []currentWebhook{
			cw("w1", "https://update.example/hook", "old", webhookStateEnabled, "org.created"),
			cw("w2", "https://delete.example/hook", "", webhookStateEnabled, "org.created"),
		}
		ops, err := diffWebhooks([]WebhookSpec{
			{URL: "https://add.example/hook", SubscribedEvents: []string{"org.created"}},
			{URL: "https://update.example/hook", Description: "new", SubscribedEvents: []string{"org.created"}},
			{URL: "https://delete.example/hook", Delete: true},
		}, current)
		assert.NoError(t, err)
		if assert.Len(t, ops, 3) {
			assert.Equal(t, opAdd, ops[0].action)
			assert.Equal(t, opUpdate, ops[1].action)
			assert.Equal(t, opRemove, ops[2].action)
		}
	})

	t.Run("a url listed twice in the file fails", func(t *testing.T) {
		_, err := diffWebhooks([]WebhookSpec{
			{URL: "https://a.example/hook", SubscribedEvents: []string{"org.created"}},
			{URL: "https://a.example/hook", SubscribedEvents: []string{"org.deleted"}},
		}, nil)
		assert.ErrorContains(t, err, "listed more than once")
	})

	t.Run("two server endpoints sharing a url fail with their ids", func(t *testing.T) {
		current := []currentWebhook{
			cw("w1", "https://a.example/hook", "", webhookStateEnabled, "org.created"),
			cw("w2", "https://a.example/hook", "", webhookStateEnabled, "org.created"),
		}
		_, err := diffWebhooks([]WebhookSpec{
			{URL: "https://a.example/hook", SubscribedEvents: []string{"org.created"}},
		}, current)
		assert.ErrorContains(t, err, "ambiguous")
		assert.ErrorContains(t, err, "w1")
		assert.ErrorContains(t, err, "w2")
	})

	t.Run("validation rejects a url that is not absolute", func(t *testing.T) {
		_, err := diffWebhooks([]WebhookSpec{{URL: "not-a-url", SubscribedEvents: []string{"org.created"}}}, nil)
		assert.ErrorContains(t, err, "absolute URL")
	})

	t.Run("validation rejects an entry with no events", func(t *testing.T) {
		_, err := diffWebhooks([]WebhookSpec{{URL: "https://a.example/hook"}}, nil)
		assert.ErrorContains(t, err, "at least one event")
	})

	t.Run("validation rejects an unknown state", func(t *testing.T) {
		_, err := diffWebhooks([]WebhookSpec{{URL: "https://a.example/hook", SubscribedEvents: []string{"org.created"}, State: "paused"}}, nil)
		assert.ErrorContains(t, err, "state must be")
	})
}
