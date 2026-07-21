package reconcile

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// KindWebhook is the desired-state document kind for webhook endpoints.
const KindWebhook = "Webhook"

// Webhook states. The server enables a new endpoint by default, so an entry
// that does not set a state leaves the server's value in place.
const (
	webhookStateEnabled  = "enabled"
	webhookStateDisabled = "disabled"
)

// WebhookSpec is one desired webhook endpoint. The URL is the identity and
// never changes. Subscribed events are the full desired set for the endpoint,
// not a per-field overlay: an empty list (or none) means the endpoint receives
// every event, which is the server's own default, and export always writes the
// field so an all-events endpoint reads as an explicit `subscribed_events: []`.
// Description and state are managed the ordinary way: present is set, omitted
// keeps the server value. Signing secrets are server-owned: the server
// generates one on create and never returns it on read, so they are not part of
// the spec.
type WebhookSpec struct {
	URL              string   `yaml:"url"`
	Description      string   `yaml:"description,omitempty"`
	SubscribedEvents []string `yaml:"subscribed_events"`
	State            string   `yaml:"state,omitempty"`
	Delete           bool     `yaml:"delete,omitempty"`
}

// currentWebhook is one endpoint as returned by ListWebhooks. Headers and
// metadata are carried through an update untouched, so a reconcile that only
// changes a managed field does not wipe values an operator set elsewhere.
// Secrets are never read: the server redacts them on list.
type currentWebhook struct {
	ID               string
	URL              string
	Description      string
	SubscribedEvents []string
	State            string
	Headers          map[string]string
	Metadata         map[string]any
}

// webhookOp is a single planned change. spec carries the final values to send;
// id is set for updates and deletes; headers and metadata are the endpoint's
// current values, carried through an update untouched.
type webhookOp struct {
	action   opAction
	spec     WebhookSpec
	id       string
	detail   string
	headers  map[string]string
	metadata map[string]any
}

func (o webhookOp) String() string {
	switch o.action {
	case opRemove:
		return fmt.Sprintf("delete webhook %s", o.spec.URL)
	case opUpdate:
		return fmt.Sprintf("update webhook %s (%s)", o.spec.URL, o.detail)
	default:
		events := "all events"
		if len(o.spec.SubscribedEvents) > 0 {
			events = strings.Join(o.spec.SubscribedEvents, ", ")
		}
		return fmt.Sprintf("add webhook %s [%s]", o.spec.URL, events)
	}
}

// validateWebhookSpec rejects entries the flow cannot manage. A live entry needs
// only a valid URL: an empty event list is allowed and means the endpoint
// receives every event, which is the server's own default.
func validateWebhookSpec(s WebhookSpec) error {
	if strings.TrimSpace(s.URL) == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.Parse(s.URL)
	if err != nil || !u.IsAbs() {
		return fmt.Errorf("url %q must be a valid absolute URL", s.URL)
	}
	// The server only dispatches over http(s) and rejects other schemes, so
	// reject them at plan time too and keep the export round-trip consistent.
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url %q must use http or https", s.URL)
	}
	if s.Delete {
		return nil
	}
	if s.State != "" && s.State != webhookStateEnabled && s.State != webhookStateDisabled {
		return fmt.Errorf("state must be %q or %q", webhookStateEnabled, webhookStateDisabled)
	}
	return nil
}

// normalizeWebhookSpecs trims each url so it matches the server's own
// normalization (the service trims before it stores), validates every entry,
// and rejects a url listed more than once. It returns the normalized specs so
// diffWebhooks and Validate work from identical, deduplicated input.
func normalizeWebhookSpecs(specs []WebhookSpec) ([]WebhookSpec, error) {
	seen := map[string]struct{}{}
	out := make([]WebhookSpec, 0, len(specs))
	for _, s := range specs {
		s.URL = strings.TrimSpace(s.URL)
		if err := validateWebhookSpec(s); err != nil {
			return nil, fmt.Errorf("invalid webhook spec %q: %w", s.URL, err)
		}
		if _, dup := seen[s.URL]; dup {
			return nil, fmt.Errorf("webhook %q is listed more than once", s.URL)
		}
		seen[s.URL] = struct{}{}
		out = append(out, s)
	}
	return out, nil
}

// diffWebhooks returns the ops that make the current webhook endpoints match the
// desired spec. The URL is the identity: every endpoint on the server must
// appear in the file — kept, or marked delete — so nothing is removed by
// omission, and an unaccounted endpoint fails the plan. Because the server does
// not enforce URL uniqueness, two endpoints sharing a URL make the identity
// ambiguous and also fail the plan, naming the ids to clean up.
func diffWebhooks(desired []WebhookSpec, current []currentWebhook) ([]webhookOp, error) {
	idsByURL := map[string][]string{}
	for _, c := range current {
		idsByURL[c.URL] = append(idsByURL[c.URL], c.ID)
	}
	var ambiguous []string
	for u, ids := range idsByURL {
		if len(ids) > 1 {
			sort.Strings(ids)
			ambiguous = append(ambiguous, fmt.Sprintf("%s (ids: %s)", u, strings.Join(ids, ", ")))
		}
	}
	if len(ambiguous) > 0 {
		sort.Strings(ambiguous)
		return nil, fmt.Errorf("the server has webhooks that share a url, so the url identity is ambiguous: %s; delete the extra one by hand, then reconcile", strings.Join(ambiguous, "; "))
	}

	byURL := make(map[string]currentWebhook, len(current))
	for _, c := range current {
		byURL[c.URL] = c
	}

	desired, err := normalizeWebhookSpecs(desired)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	var adds, updates, removes []webhookOp
	for _, s := range desired {
		seen[s.URL] = struct{}{}

		cur, exists := byURL[s.URL]
		if s.Delete {
			if exists {
				removes = append(removes, webhookOp{action: opRemove, spec: s, id: cur.ID})
			}
			continue
		}
		desiredEvents := uniqueSorted(s.SubscribedEvents)
		if !exists {
			adds = append(adds, webhookOp{action: opAdd, spec: WebhookSpec{
				URL:              s.URL,
				Description:      s.Description,
				SubscribedEvents: desiredEvents,
				State:            s.State,
			}})
			continue
		}

		merged := WebhookSpec{
			URL:              s.URL,
			Description:      cur.Description,
			SubscribedEvents: uniqueSorted(cur.SubscribedEvents),
			State:            cur.State,
		}
		var changes []string
		if s.Description != "" && s.Description != cur.Description {
			merged.Description = s.Description
			changes = append(changes, "description")
		}
		// Events are the full desired set, always compared. An empty or omitted
		// list means every event, so leaving it out sets the endpoint to all
		// events rather than keeping the server's set. The set is deduplicated so
		// a hand-written list that repeats a value does not send duplicates or
		// plan a spurious update.
		if !stringSetsEqual(desiredEvents, uniqueSorted(cur.SubscribedEvents)) {
			merged.SubscribedEvents = desiredEvents
			changes = append(changes, "subscribed_events")
		}
		if s.State != "" && s.State != cur.State {
			merged.State = s.State
			changes = append(changes, "state")
		}
		if len(changes) > 0 {
			updates = append(updates, webhookOp{
				action:   opUpdate,
				spec:     merged,
				id:       cur.ID,
				detail:   strings.Join(changes, ", "),
				headers:  cur.Headers,
				metadata: cur.Metadata,
			})
		}
	}

	var unaccounted []string
	for u := range byURL {
		if _, ok := seen[u]; !ok {
			unaccounted = append(unaccounted, u)
		}
	}
	if len(unaccounted) > 0 {
		sort.Strings(unaccounted)
		return nil, fmt.Errorf("webhooks exist on the server but are not in the file: %s; keep them or mark them 'delete: true'", strings.Join(unaccounted, ", "))
	}

	ops := append(adds, updates...)
	return append(ops, removes...), nil
}
