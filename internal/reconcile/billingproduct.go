package reconcile

import (
	"fmt"
	"sort"
	"strings"
)

// KindBillingProduct is the desired-state document kind for billing products.
const KindBillingProduct = "BillingProduct"

// The set of valid behaviors, intervals, usage types, and schemes lives in the
// proto (buf.validate) and the billing/product package, and the server rejects
// bad values through its validate interceptor. This kind does not re-list them,
// so a new value does not need a change here. The one value it must know is the
// tiered scheme, which it cannot represent (there is no tiers field), so it is
// rejected up front.
const billingSchemeTiered = "tiered"

// BillingProductSpec is one desired product. The name is the identity and never
// changes. A product is managed in full: its title, description, behavior,
// config, prices, and features all state the whole desired value. Prices are
// keyed by name within the product and converge on the server (a new name is
// added, an existing name keeps its immutable fields, and an active price the
// list no longer names is deactivated). A product cannot be deleted through the
// API, so the delete flag is rejected.
type BillingProductSpec struct {
	Name        string               `yaml:"name"`
	Title       string               `yaml:"title,omitempty"`
	Description string               `yaml:"description,omitempty"`
	Behavior    string               `yaml:"behavior,omitempty"`
	Config      BillingProductConfig `yaml:"config,omitempty"`
	Prices      []BillingPriceSpec   `yaml:"prices,omitempty"`
	Features    []BillingFeatureRef  `yaml:"features,omitempty"`
	Metadata    map[string]any       `yaml:"metadata,omitempty"`
	Delete      bool                 `yaml:"delete,omitempty"`
}

// BillingProductConfig is the behavior config of a product.
type BillingProductConfig struct {
	CreditAmount int64 `yaml:"credit_amount,omitempty"`
	SeatLimit    int64 `yaml:"seat_limit,omitempty"`
	MinQuantity  int64 `yaml:"min_quantity,omitempty"`
	MaxQuantity  int64 `yaml:"max_quantity,omitempty"`
}

// BillingPriceSpec is one desired price of a product. The name is the identity
// within the product. Amount and the other pricing fields are fixed once a
// price exists, so a change to them must be a new price under a new name.
type BillingPriceSpec struct {
	Name             string `yaml:"name"`
	Amount           int64  `yaml:"amount"`
	Currency         string `yaml:"currency,omitempty"`
	Interval         string `yaml:"interval,omitempty"`
	UsageType        string `yaml:"usage_type,omitempty"`
	BillingScheme    string `yaml:"billing_scheme,omitempty"`
	MeteredAggregate string `yaml:"metered_aggregate,omitempty"`
}

// BillingFeatureRef names a feature attached to a product. The feature is
// created if it does not exist yet.
type BillingFeatureRef struct {
	Name string `yaml:"name"`
}

// currentBillingProduct is one product as returned by ListProducts. Prices are
// the active ones only, since an inactive price is retired and must not be
// treated as part of the current desired state.
type currentBillingProduct struct {
	ID          string
	Name        string
	Title       string
	Description string
	Behavior    string
	Config      BillingProductConfig
	Prices      []BillingPriceSpec
	Features    []string
	Metadata    map[string]any
}

// billingProductOp is a single planned change. spec carries the whole desired
// product; id is set for an update.
type billingProductOp struct {
	action opAction
	spec   BillingProductSpec
	id     string
	detail string
}

func (o billingProductOp) String() string {
	if o.action == opUpdate {
		return fmt.Sprintf("update product %s (%s)", o.spec.Name, o.detail)
	}
	prices := "no prices"
	if len(o.spec.Prices) > 0 {
		names := make([]string, 0, len(o.spec.Prices))
		for _, p := range o.spec.Prices {
			names = append(names, p.Name)
		}
		prices = "prices: " + strings.Join(uniqueSorted(names), ", ")
	}
	return fmt.Sprintf("add product %s [%s]", o.spec.Name, prices)
}

// validateBillingProductSpec rejects entries the flow cannot manage without
// touching the server: a missing name, a delete flag (products cannot be
// removed through the API), an unknown behavior, and any malformed price or
// feature. Names must be unique within their scope.
func validateBillingProductSpec(s BillingProductSpec) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("product name is required")
	}
	if s.Delete {
		return fmt.Errorf("product %q cannot be deleted: there is no product delete or deactivate API; remove the entry and archive the product by hand", s.Name)
	}

	seenPrice := map[string]struct{}{}
	for _, p := range s.Prices {
		name := strings.ToLower(strings.TrimSpace(p.Name))
		if name == "" {
			return fmt.Errorf("product %q has a price with no name", s.Name)
		}
		if _, dup := seenPrice[name]; dup {
			return fmt.Errorf("product %q lists price %q more than once", s.Name, name)
		}
		seenPrice[name] = struct{}{}
		if p.Amount < 0 {
			return fmt.Errorf("product %q price %q has a negative amount", s.Name, name)
		}
		if strings.ToLower(p.BillingScheme) == billingSchemeTiered {
			return fmt.Errorf("product %q price %q uses a tiered billing scheme, which this kind does not support", s.Name, name)
		}
	}

	seenFeature := map[string]struct{}{}
	for _, f := range s.Features {
		name := strings.ToLower(strings.TrimSpace(f.Name))
		if name == "" {
			return fmt.Errorf("product %q has a feature with no name", s.Name)
		}
		if _, dup := seenFeature[name]; dup {
			return fmt.Errorf("product %q lists feature %q more than once", s.Name, name)
		}
		seenFeature[name] = struct{}{}
	}
	return nil
}

// normalizeBillingProductSpecs trims each product name, validates every entry,
// and rejects a product listed more than once, so Validate and diff work from
// identical, deduplicated input.
func normalizeBillingProductSpecs(specs []BillingProductSpec) ([]BillingProductSpec, error) {
	seen := map[string]struct{}{}
	out := make([]BillingProductSpec, 0, len(specs))
	for _, s := range specs {
		s.Name = strings.TrimSpace(s.Name)
		if err := validateBillingProductSpec(s); err != nil {
			return nil, fmt.Errorf("invalid billing product spec %q: %w", s.Name, err)
		}
		key := strings.ToLower(s.Name)
		if _, dup := seen[key]; dup {
			return nil, fmt.Errorf("product %q is listed more than once", s.Name)
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out, nil
}

// diffBillingProducts returns the ops that make the current products match the
// desired spec. The name is the identity: a product not on the server is added,
// a product whose managed fields differ is updated, and a product on the server
// that the file does not list fails the plan, since a product cannot be removed
// through the API. Nested prices and features converge on the server, so an
// update sends the whole desired product.
func diffBillingProducts(desired []BillingProductSpec, current []currentBillingProduct) ([]billingProductOp, error) {
	desired, err := normalizeBillingProductSpecs(desired)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]currentBillingProduct, len(current))
	for _, c := range current {
		byName[strings.ToLower(c.Name)] = c
	}

	seen := map[string]struct{}{}
	var adds, updates []billingProductOp
	for _, s := range desired {
		key := strings.ToLower(s.Name)
		seen[key] = struct{}{}

		cur, exists := byName[key]
		if !exists {
			adds = append(adds, billingProductOp{action: opAdd, spec: s})
			continue
		}
		changes, err := billingProductChanges(s, cur)
		if err != nil {
			return nil, err
		}
		if len(changes) > 0 {
			updates = append(updates, billingProductOp{
				action: opUpdate,
				spec:   s,
				id:     cur.ID,
				detail: strings.Join(changes, ", "),
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
		return nil, fmt.Errorf("products exist on the server but are not in the file: %s; a product cannot be removed through the API, so add it back to the file", strings.Join(unaccounted, ", "))
	}

	return append(adds, updates...), nil
}

// billingProductChanges lists the managed-field categories that differ between a
// desired product and its current state, in a way that matches what the server's
// UpdateProduct will actually apply. That server merges rather than full-writes:
// it keeps title, description, and metadata when empty and keeps credit_amount
// and seat_limit when zero, and it never changes a product's behavior. So a
// change toward one of those unsettable values is not planned, since planning it
// would loop forever against a merge that drops it. An immutable price change is
// a hard error, not a plannable update. An empty result means the product
// already matches and needs no update.
func billingProductChanges(s BillingProductSpec, cur currentBillingProduct) ([]string, error) {
	var changes []string
	// title and description are keep-if-empty on the server, so a change is only
	// plannable when the desired value is non-empty and differs.
	if s.Title != "" && s.Title != cur.Title {
		changes = append(changes, "title")
	}
	if s.Description != "" && s.Description != cur.Description {
		changes = append(changes, "description")
	}
	// behavior is create-only: the server never changes it on update, and it may
	// rewrite it on create (credit_amount > 0 becomes "credits"), so a difference
	// is not plannable and is left out.
	if billingConfigChanged(s.Config, cur.Config) {
		changes = append(changes, "config")
	}
	if !billingFeatureSetsEqual(s.Features, cur.Features) {
		changes = append(changes, "features")
	}
	priceChanged, err := billingPriceChange(s.Prices, cur.Prices)
	if err != nil {
		return nil, fmt.Errorf("product %q: %w", s.Name, err)
	}
	if priceChanged {
		changes = append(changes, "prices")
	}
	return changes, nil
}

// billingConfigChanged reports whether a config change is both wanted and
// applicable. credit_amount and seat_limit are keep-if-zero on the server, so a
// change to them is only plannable when the desired value is non-zero. min and
// max quantity are written unconditionally, so any difference is a change.
func billingConfigChanged(desired, cur BillingProductConfig) bool {
	if desired.CreditAmount > 0 && desired.CreditAmount != cur.CreditAmount {
		return true
	}
	if desired.SeatLimit > 0 && desired.SeatLimit != cur.SeatLimit {
		return true
	}
	return desired.MinQuantity != cur.MinQuantity || desired.MaxQuantity != cur.MaxQuantity
}

func billingFeatureSetsEqual(desired []BillingFeatureRef, current []string) bool {
	d := make([]string, 0, len(desired))
	for _, f := range desired {
		d = append(d, strings.ToLower(strings.TrimSpace(f.Name)))
	}
	c := make([]string, 0, len(current))
	for _, name := range current {
		c = append(c, strings.ToLower(strings.TrimSpace(name)))
	}
	return stringSetsEqual(uniqueSorted(d), uniqueSorted(c))
}

// billingPriceChange reports whether the desired prices need a converging update
// against the current active prices, and fails when the file asks for a change
// the server cannot apply. Prices converge only when the desired list is
// non-empty, because the server ignores an empty price list, so an empty desired
// list is treated as keep (this kind cannot drop the last price). A desired name
// that is new, or an active name the list no longer includes, needs an update. A
// desired name that already exists but changes an immutable field is rejected:
// a provider price cannot change in place, so it must be a new price name.
func billingPriceChange(desired, current []BillingPriceSpec) (bool, error) {
	if len(desired) == 0 {
		return false, nil
	}
	index := func(prices []BillingPriceSpec) map[string]BillingPriceSpec {
		m := make(map[string]BillingPriceSpec, len(prices))
		for _, p := range prices {
			m[strings.ToLower(strings.TrimSpace(p.Name))] = normalizeBillingPrice(p)
		}
		return m
	}
	dm, cm := index(desired), index(current)
	for name, d := range dm {
		c, ok := cm[name]
		if !ok {
			return true, nil // a new price name to add
		}
		if d != c {
			return false, fmt.Errorf("price %q cannot change its amount, currency, interval, scheme, usage type, or aggregate; provider prices are immutable, so add a new price under a new name", name)
		}
	}
	for name := range cm {
		if _, ok := dm[name]; !ok {
			return true, nil // an active price to retire
		}
	}
	return false, nil
}

// normalizeBillingPrice fills the defaults the billing service applies and
// lowercases the case-insensitive fields, so a price that omits a defaulted
// field, or writes its currency in a different case, compares equal to the
// stored one. metered_aggregate defaults to "sum" because the create path fills
// it while the add-a-price path leaves it empty, and both should read the same.
// The name is dropped from the comparison value because it is the map key.
func normalizeBillingPrice(p BillingPriceSpec) BillingPriceSpec {
	if p.Currency == "" {
		p.Currency = "usd"
	}
	p.Currency = strings.ToLower(p.Currency)
	if p.MeteredAggregate == "" {
		p.MeteredAggregate = "sum"
	}
	if p.UsageType == "" {
		p.UsageType = "licensed"
	}
	if p.BillingScheme == "" {
		p.BillingScheme = "flat"
	}
	p.Name = ""
	p.Interval = strings.ToLower(p.Interval)
	return p
}
