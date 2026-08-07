package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// KindBillingPlan is the desired-state document kind for billing plans.
const KindBillingPlan = "BillingPlan"

// BillingPlanSpec is one desired plan. The name is the identity and never
// changes. A plan groups products (referenced by name; the products themselves
// are managed by the BillingProduct kind). Title, description, on_start_credits,
// trial_days, and state are converged through UpdatePlan. Interval and the
// product set are create-only: UpdatePlan cannot change them, so a change to
// either fails the plan. A plan cannot be deleted through the API, so the delete
// flag is rejected. Metadata is out of scope: it is not stated in the file, not
// diffed, and not exported, but it is preserved on update (see the reconciler).
type BillingPlanSpec struct {
	Name           string                  `yaml:"name"`
	Title          string                  `yaml:"title,omitempty"`
	Description    string                  `yaml:"description,omitempty"`
	Interval       string                  `yaml:"interval,omitempty"`
	OnStartCredits int64                   `yaml:"on_start_credits,omitempty"`
	TrialDays      int64                   `yaml:"trial_days,omitempty"`
	State          string                  `yaml:"state,omitempty"`
	Products       []BillingPlanProductRef `yaml:"products,omitempty"`
	Delete         bool                    `yaml:"delete,omitempty"`
}

// BillingPlanProductRef names a product that belongs to the plan. The product is
// managed by the BillingProduct kind; the plan only references it by name.
type BillingPlanProductRef struct {
	Name string `yaml:"name"`
}

// currentBillingPlan is one plan as returned by ListAllPlans, including inactive
// ones. Products holds the names of the products attached to the plan. Metadata
// is carried so an update can re-send it; the plan kind does not otherwise manage
// metadata.
type currentBillingPlan struct {
	ID             string
	Name           string
	Title          string
	Description    string
	Interval       string
	OnStartCredits int64
	TrialDays      int64
	State          string
	Products       []string
	Metadata       *structpb.Struct
}

// billingPlanOp is a single planned change. spec carries the whole desired plan;
// id is set for an update, and metadata holds the current plan's metadata so the
// update can preserve it (UpdatePlan is a full write of the fields it carries).
type billingPlanOp struct {
	action   opAction
	spec     BillingPlanSpec
	id       string
	detail   string
	metadata *structpb.Struct
}

func (o billingPlanOp) String() string {
	if o.action == opUpdate {
		return fmt.Sprintf("update plan %s (%s)", o.spec.Name, o.detail)
	}
	products := "no products"
	if len(o.spec.Products) > 0 {
		names := make([]string, 0, len(o.spec.Products))
		for _, p := range o.spec.Products {
			names = append(names, p.Name)
		}
		products = "products: " + strings.Join(uniqueSorted(names), ", ")
	}
	return fmt.Sprintf("add plan %s [%s]", o.spec.Name, products)
}

// validateBillingPlanSpec rejects entries the flow cannot manage without touching
// the server: a missing or too-short name, a missing title, a delete flag (plans
// cannot be removed through the API), and a duplicate product reference. It does
// not re-list the valid intervals or states; the server rejects a bad value
// through its validate interceptor, checked in the reconciler.
func validateBillingPlanSpec(s BillingPlanSpec) error {
	name := strings.TrimSpace(s.Name)
	if name == "" {
		return fmt.Errorf("plan name is required")
	}
	// the server requires a plan name of at least three characters, so a shorter
	// one would fail at apply; reject it here instead.
	if len(name) < 3 {
		return fmt.Errorf("plan name %q must be at least three characters", name)
	}
	// title is written in full, so an omitted one would plan a reset toward empty;
	// require it up front. A plan should have a title.
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("plan %q must have a title", name)
	}
	if s.Delete {
		return fmt.Errorf("plan %q cannot be deleted: there is no plan delete API; set its state to inactive instead, or remove the entry and archive it by hand", s.Name)
	}

	seenProduct := map[string]struct{}{}
	for _, p := range s.Products {
		productName := strings.ToLower(strings.TrimSpace(p.Name))
		if productName == "" {
			return fmt.Errorf("plan %q references a product with no name", s.Name)
		}
		if _, dup := seenProduct[productName]; dup {
			return fmt.Errorf("plan %q references product %q more than once", s.Name, productName)
		}
		seenProduct[productName] = struct{}{}
	}
	return nil
}

// normalizeBillingPlanSpecs trims each plan name, validates every entry, and
// rejects a plan listed more than once, so Validate and diff work from identical,
// deduplicated input.
func normalizeBillingPlanSpecs(specs []BillingPlanSpec) ([]BillingPlanSpec, error) {
	seen := map[string]struct{}{}
	out := make([]BillingPlanSpec, 0, len(specs))
	for _, s := range specs {
		s.Name = strings.TrimSpace(s.Name)
		if err := validateBillingPlanSpec(s); err != nil {
			return nil, fmt.Errorf("invalid billing plan spec %q: %w", s.Name, err)
		}
		key := strings.ToLower(s.Name)
		if _, dup := seen[key]; dup {
			return nil, fmt.Errorf("plan %q is listed more than once", s.Name)
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out, nil
}

// diffBillingPlans returns the ops that make the current plans match the desired
// spec. The name is the identity: a plan not on the server is added, a plan whose
// managed fields differ is updated, and a plan on the server that the file does
// not list fails the plan, since a plan cannot be removed through the API (retire
// it with state instead).
func diffBillingPlans(desired []BillingPlanSpec, current []currentBillingPlan) ([]billingPlanOp, error) {
	desired, err := normalizeBillingPlanSpecs(desired)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]currentBillingPlan, len(current))
	for _, c := range current {
		byName[strings.ToLower(c.Name)] = c
	}

	seen := map[string]struct{}{}
	var adds, updates []billingPlanOp
	for _, s := range desired {
		key := strings.ToLower(s.Name)
		seen[key] = struct{}{}

		cur, exists := byName[key]
		if !exists {
			adds = append(adds, billingPlanOp{action: opAdd, spec: s})
			continue
		}
		changes, err := billingPlanChanges(s, cur)
		if err != nil {
			return nil, err
		}
		if len(changes) > 0 {
			updates = append(updates, billingPlanOp{
				action:   opUpdate,
				spec:     s,
				id:       cur.ID,
				detail:   strings.Join(changes, ", "),
				metadata: cur.Metadata,
			})
		}
	}

	var unaccounted []string
	for _, c := range current {
		if _, ok := seen[strings.ToLower(c.Name)]; !ok {
			unaccounted = append(unaccounted, c.Name)
		}
	}
	if len(unaccounted) > 0 {
		sort.Strings(unaccounted)
		return nil, fmt.Errorf("plans exist on the server but are not in the file: %s; a plan cannot be removed through the API, so add it back to the file (set its state to inactive to retire it)", strings.Join(unaccounted, ", "))
	}

	return append(adds, updates...), nil
}

// billingPlanChanges lists the managed fields that differ between a desired plan
// and its current state, matching what UpdatePlan will apply. Title, description,
// on_start_credits, trial_days, and state are written in full, so any difference
// is a plannable change. Interval and the product set are create-only: UpdatePlan
// does not touch them, so a change to either fails the plan. An empty result means
// the plan already matches and needs no update.
func billingPlanChanges(s BillingPlanSpec, cur currentBillingPlan) ([]string, error) {
	// interval is create-only: the server sets it at create and UpdatePlan cannot
	// change it. A file that asks to change it cannot apply, so fail the plan.
	if !strings.EqualFold(strings.TrimSpace(s.Interval), strings.TrimSpace(cur.Interval)) {
		return nil, fmt.Errorf("plan %q interval cannot change from %q to %q after creation; create a new plan to change the interval", s.Name, cur.Interval, s.Interval)
	}
	// the product set is create-only too: UpdatePlan does not change a plan's
	// products, so a change to them cannot apply.
	if !billingPlanProductSetsEqual(s.Products, cur.Products) {
		return nil, fmt.Errorf("plan %q products cannot change after creation; UpdatePlan does not change a plan's products, so create a new plan", s.Name)
	}

	var changes []string
	if s.Title != cur.Title {
		changes = append(changes, "title")
	}
	if s.Description != cur.Description {
		changes = append(changes, "description")
	}
	if s.OnStartCredits != cur.OnStartCredits {
		changes = append(changes, "on_start_credits")
	}
	if s.TrialDays != cur.TrialDays {
		changes = append(changes, "trial_days")
	}
	if normalizeBillingPlanState(s.State) != normalizeBillingPlanState(cur.State) {
		changes = append(changes, "state")
	}
	return changes, nil
}

// normalizeBillingPlanState maps an empty state to "active" (the server default)
// and lowercases the value, so a plan that omits state or writes it in another
// case compares equal to the stored one.
func normalizeBillingPlanState(state string) string {
	s := strings.ToLower(strings.TrimSpace(state))
	if s == "" {
		return "active"
	}
	return s
}

// billingPlanProductSetsEqual reports whether the desired product references name
// the same set as the current plan, case-insensitively and order-independently.
func billingPlanProductSetsEqual(desired []BillingPlanProductRef, current []string) bool {
	d := make([]string, 0, len(desired))
	for _, p := range desired {
		d = append(d, strings.ToLower(strings.TrimSpace(p.Name)))
	}
	c := make([]string, 0, len(current))
	for _, name := range current {
		c = append(c, strings.ToLower(strings.TrimSpace(name)))
	}
	return stringSetsEqual(uniqueSorted(d), uniqueSorted(c))
}
