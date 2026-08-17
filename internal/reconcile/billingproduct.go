package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/raystack/frontier/billing/product"
)

// KindBillingProduct is the desired-state document kind for billing products.
const KindBillingProduct = "BillingProduct"

// The set of valid behaviors, intervals, usage types, and schemes lives in the
// proto (buf.validate) and the billing/product package, and the server rejects
// bad values through its validate interceptor. This kind does not re-list them,
// so a new value does not need a change here. The one value it must know is the
// tiered scheme, which it cannot represent (there is no tiers field), so it is
// rejected up front; its value comes from the product package so the two cannot
// drift.
const billingSchemeTiered = string(product.BillingSchemeTiered)

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

// currentBillingProduct is one product as returned by ListProducts. Prices holds
// the active prices, which make up the current desired state. RetiredPrices holds
// the inactive ones: they are not part of the desired state, but the server still
// validates a reused name against them, so the diff needs them to tell a safe
// reactivation from a change the server would reject.
type currentBillingProduct struct {
	ID            string
	Name          string
	Title         string
	Description   string
	Behavior      string
	Config        BillingProductConfig
	Prices        []BillingPriceSpec
	RetiredPrices []BillingPriceSpec
	Features      []string
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
// touching the server: a missing or too-short name, a missing title, a delete
// flag (products cannot be removed through the API), and any malformed price or
// feature. It does not re-list the valid behaviors, intervals, or schemes; the
// server rejects a bad value through its validate interceptor. Names must be
// unique within their scope.
func validateBillingProductSpec(s BillingProductSpec) error {
	name := strings.TrimSpace(s.Name)
	if name == "" {
		return fmt.Errorf("product name is required")
	}
	// the server requires a product name of at least three characters, so a
	// shorter one would fail at apply; reject it here instead.
	if len(name) < 3 {
		return fmt.Errorf("product name %q must be at least three characters", name)
	}
	// the billing provider uses the title as the product name and requires it to
	// be non-empty. Since title is written in full, an omitted or empty one would
	// plan a reset the provider rejects, so require it up front.
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("product %q must have a title", name)
	}
	if s.Delete {
		return fmt.Errorf("product %q cannot be deleted: there is no product delete or deactivate API; remove the entry and archive the product by hand", s.Name)
	}

	seenPrice := map[string]struct{}{}
	for _, p := range s.Prices {
		priceName := strings.ToLower(strings.TrimSpace(p.Name))
		if priceName == "" {
			return fmt.Errorf("product %q has a price with no name", s.Name)
		}
		if _, dup := seenPrice[priceName]; dup {
			return fmt.Errorf("product %q lists price %q more than once", s.Name, priceName)
		}
		seenPrice[priceName] = struct{}{}
		if p.Amount < 0 {
			return fmt.Errorf("product %q price %q has a negative amount", s.Name, priceName)
		}
		if strings.ToLower(p.BillingScheme) == billingSchemeTiered {
			return fmt.Errorf("product %q price %q uses a tiered billing scheme, which this kind does not support", s.Name, priceName)
		}
	}

	seenFeature := map[string]struct{}{}
	for _, f := range s.Features {
		featureName := strings.ToLower(strings.TrimSpace(f.Name))
		if featureName == "" {
			return fmt.Errorf("product %q has a feature with no name", s.Name)
		}
		if _, dup := seenFeature[featureName]; dup {
			return fmt.Errorf("product %q lists feature %q more than once", s.Name, featureName)
		}
		seenFeature[featureName] = struct{}{}
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
// UpdateProduct will actually apply. The server writes title, description, and
// the behavior config as given, so a difference in any of them, including toward
// an empty or zero value, is a plannable change. Behavior is create-only, so a
// change to it fails the plan rather than being applied. An immutable price change
// is a hard error too. An empty result means the product already matches and needs
// no update.
func billingProductChanges(s BillingProductSpec, cur currentBillingProduct) ([]string, error) {
	var changes []string
	// title and description state the whole desired value, so any difference is a
	// change, including clearing a field the server currently has.
	if s.Title != cur.Title {
		changes = append(changes, "title")
	}
	if s.Description != cur.Description {
		changes = append(changes, "description")
	}
	// behavior is create-only: the server sets it at create and never changes it
	// on update. The file's behavior is honored as written, so a file that names a
	// behavior the product was not created with fails the plan rather than being
	// silently overridden. Only an omitted behavior on a credit product falls back
	// to "credits", matching the server's create-time default, so an omitted
	// behavior does not read as a change against a credit product.
	expectedBehavior := s.Behavior
	if expectedBehavior == "" && s.Config.CreditAmount > 0 {
		expectedBehavior = "credits"
	}
	if expectedBehavior != "" && expectedBehavior != cur.Behavior {
		return nil, fmt.Errorf("product %q behavior cannot change from %q to %q after creation; create a new product to change behavior", s.Name, cur.Behavior, expectedBehavior)
	}
	if billingConfigChanged(s.Config, cur.Config) {
		changes = append(changes, "config")
	}
	if !billingFeatureSetsEqual(s.Features, cur.Features) {
		changes = append(changes, "features")
	}
	priceChanged, err := billingPriceChange(s.Prices, cur.Prices, cur.RetiredPrices)
	if err != nil {
		return nil, fmt.Errorf("product %q: %w", s.Name, err)
	}
	if priceChanged {
		changes = append(changes, "prices")
	}
	return changes, nil
}

// billingConfigChanged reports whether the desired behavior config differs from
// the current one. The server writes all four fields as given, so any difference,
// including a reset to zero, is a change.
func billingConfigChanged(desired, cur BillingProductConfig) bool {
	return desired.CreditAmount != cur.CreditAmount ||
		desired.SeatLimit != cur.SeatLimit ||
		desired.MinQuantity != cur.MinQuantity ||
		desired.MaxQuantity != cur.MaxQuantity
}

// normalizeFeatureName is the canonical form of a feature name, used by both the
// diff and the apply so a name is compared and written the same way. The server
// looks features up by name, so a case or whitespace difference between the two
// would fork a duplicate feature the diff had reported as unchanged.
func normalizeFeatureName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func billingFeatureSetsEqual(desired []BillingFeatureRef, current []string) bool {
	d := make([]string, 0, len(desired))
	for _, f := range desired {
		d = append(d, normalizeFeatureName(f.Name))
	}
	c := make([]string, 0, len(current))
	for _, name := range current {
		c = append(c, normalizeFeatureName(name))
	}
	return stringSetsEqual(uniqueSorted(d), uniqueSorted(c))
}

// billingPriceChange reports whether the desired prices need a converging update
// against the current prices, and fails when the file asks for a change the
// server cannot apply. Prices converge only when the desired list is non-empty,
// because the server ignores an empty price list, so an empty desired list is
// treated as keep (this kind cannot drop the last price).
//
// The server validates a reused name against every price of the product, active
// or retired, so both are checked here. A desired name that matches an active or
// a retired price but changes an immutable field is rejected, since a provider
// price cannot change in place. A desired name that is new, or that matches a
// retired price with the same fields (a reactivation), or an active name the list
// no longer includes (a deactivation), needs a converging update.
func billingPriceChange(desired, active, retired []BillingPriceSpec) (bool, error) {
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
	dm, am, rm := index(desired), index(active), index(retired)
	// first reject any immutable-field change on a name that already exists,
	// whether it is active or retired. This pass runs over every matched name
	// before adds, reactivations, or deactivations are considered, so the
	// outcome never depends on map iteration order.
	for name, d := range dm {
		if c, ok := am[name]; ok {
			if d != c {
				return false, fmt.Errorf("price %q cannot change its amount, currency, interval, scheme, usage type, or aggregate; provider prices are immutable, so add a new price under a new name", name)
			}
			continue
		}
		if c, ok := rm[name]; ok && d != c {
			return false, fmt.Errorf("price %q was retired earlier with different fields, so it cannot be reused with these values; provider prices are immutable, so add a new price under a new name", name)
		}
	}
	// then a new price name, or a retired name listed again (a reactivation),
	// needs a converging update. A name already active with the same fields does
	// not.
	for name := range dm {
		if _, ok := am[name]; !ok {
			return true, nil
		}
	}
	// an active price the desired list no longer names is deactivated.
	for name := range am {
		if _, ok := dm[name]; !ok {
			return true, nil
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
	if p.UsageType == "" {
		p.UsageType = "licensed"
	}
	p.UsageType = strings.ToLower(p.UsageType)
	// metered_aggregate only means something for a metered price, and the server
	// only checks it then. Default it to "sum" for a metered price and blank it
	// for the rest, so a licensed price carrying a stray aggregate does not
	// false-reject against a stored one.
	if p.UsageType == string(product.PriceUsageTypeMetered) {
		if p.MeteredAggregate == "" {
			p.MeteredAggregate = "sum"
		}
		p.MeteredAggregate = strings.ToLower(p.MeteredAggregate)
	} else {
		p.MeteredAggregate = ""
	}
	if p.BillingScheme == "" {
		p.BillingScheme = "flat"
	}
	p.BillingScheme = strings.ToLower(p.BillingScheme)
	p.Name = ""
	p.Interval = strings.ToLower(p.Interval)
	return p
}
