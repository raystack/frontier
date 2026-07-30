package reconcile

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type fakeBillingProductAPI struct {
	products []*frontierv1beta1.Product
	created  []*frontierv1beta1.ProductRequestBody
	updated  map[string]*frontierv1beta1.ProductRequestBody
}

func (f *fakeBillingProductAPI) ListProducts(_ context.Context, _ *connect.Request[frontierv1beta1.ListProductsRequest]) (*connect.Response[frontierv1beta1.ListProductsResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListProductsResponse{Products: f.products}), nil
}

func (f *fakeBillingProductAPI) CreateProduct(_ context.Context, req *connect.Request[frontierv1beta1.CreateProductRequest]) (*connect.Response[frontierv1beta1.CreateProductResponse], error) {
	f.created = append(f.created, req.Msg.GetBody())
	return connect.NewResponse(&frontierv1beta1.CreateProductResponse{}), nil
}

func (f *fakeBillingProductAPI) UpdateProduct(_ context.Context, req *connect.Request[frontierv1beta1.UpdateProductRequest]) (*connect.Response[frontierv1beta1.UpdateProductResponse], error) {
	if f.updated == nil {
		f.updated = map[string]*frontierv1beta1.ProductRequestBody{}
	}
	f.updated[req.Msg.GetId()] = req.Msg.GetBody()
	return connect.NewResponse(&frontierv1beta1.UpdateProductResponse{}), nil
}

func pricePB(name string, amount int64, currency, interval, state string) *frontierv1beta1.Price {
	return &frontierv1beta1.Price{
		Name: name, Amount: amount, Currency: currency, Interval: interval,
		UsageType: "licensed", BillingScheme: "flat", State: state,
	}
}

func billingProductPB(id, name, title, behavior string, prices []*frontierv1beta1.Price, features ...string) *frontierv1beta1.Product {
	var fs []*frontierv1beta1.Feature
	for _, f := range features {
		fs = append(fs, &frontierv1beta1.Feature{Name: f})
	}
	return &frontierv1beta1.Product{
		Id: id, Name: name, Title: title, Behavior: behavior,
		Prices: prices, Features: fs,
		BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{},
	}
}

func TestBillingProductReconciler(t *testing.T) {
	ctx := context.Background()

	t.Run("plans and applies an add and an update", func(t *testing.T) {
		api := &fakeBillingProductAPI{products: []*frontierv1beta1.Product{
			billingProductPB("p1", "token", "Tokens", "credits", []*frontierv1beta1.Price{pricePB("default", 100, "usd", "month", "active")}, "f1"),
		}}
		spec := []byte(`
- name: token
  title: Renamed tokens
  behavior: credits
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
  features:
    - {name: f1}
- name: seat
  title: Seat
  behavior: per_seat
  prices:
    - {name: monthly, amount: 15000, currency: usd, interval: month}
`)
		rep, err := NewBillingProductReconciler(api, "").Reconcile(ctx, spec, false)
		assert.NoError(t, err)
		assert.Equal(t, 2, rep.Applied)
		if assert.Len(t, api.created, 1) {
			assert.Equal(t, "seat", api.created[0].GetName())
		}
		if body := api.updated["p1"]; assert.NotNil(t, body) {
			assert.Equal(t, "token", body.GetName()) // identity never changes
			assert.Equal(t, "Renamed tokens", body.GetTitle())
		}
	})

	t.Run("fails when a server product is missing from the file", func(t *testing.T) {
		api := &fakeBillingProductAPI{products: []*frontierv1beta1.Product{
			billingProductPB("p1", "token", "Tokens", "credits", []*frontierv1beta1.Price{pricePB("default", 100, "usd", "month", "active")}),
		}}
		_, err := NewBillingProductReconciler(api, "").Reconcile(ctx, []byte(`[]`), true)
		assert.ErrorContains(t, err, "not in the file")
	})

	t.Run("skips a product the kind cannot represent (tiered price or short name)", func(t *testing.T) {
		api := &fakeBillingProductAPI{products: []*frontierv1beta1.Product{
			billingProductPB("p1", "token", "Tokens", "credits", []*frontierv1beta1.Price{pricePB("default", 100, "usd", "month", "active")}, "f1"),
			billingProductPB("p2", "ab", "Short", "basic", []*frontierv1beta1.Price{pricePB("m", 1, "usd", "month", "active")}),
			billingProductPB("p3", "tieredprod", "Tiered", "basic", []*frontierv1beta1.Price{
				{Name: "graduated", Amount: 1, Currency: "usd", Interval: "month", UsageType: "licensed", BillingScheme: "tiered", State: "active"},
			}),
		}}
		r := NewBillingProductReconciler(api, "")

		// export writes only the representable product, so the short-name and
		// tiered products do not appear and cannot fail their own re-validation.
		exported, err := r.Export(ctx)
		assert.NoError(t, err)
		specs, ok := exported.([]BillingProductSpec)
		if assert.True(t, ok) {
			assert.Len(t, specs, 1)
			assert.Equal(t, "token", specs[0].Name)
		}

		// the out-of-scope products are invisible to the diff too, so listing only
		// the representable product in the file plans no changes (not "unaccounted").
		specBytes, err := yaml.Marshal(exported)
		assert.NoError(t, err)
		rep, err := r.Reconcile(ctx, specBytes, true)
		assert.NoError(t, err)
		assert.Empty(t, rep.Planned)

		// naming an out-of-scope product in the file fails the plan up front, rather
		// than planning a create the apply would reject on the unique name.
		named := []byte(`
- name: tieredprod
  title: Tiered
  prices:
    - {name: flat, amount: 1, currency: usd, interval: month}
`)
		_, err = r.Reconcile(ctx, named, true)
		assert.ErrorContains(t, err, "out of scope")
	})

	t.Run("export round-trips to no changes and drops inactive prices", func(t *testing.T) {
		api := &fakeBillingProductAPI{products: []*frontierv1beta1.Product{
			billingProductPB("p1", "token", "Tokens", "credits", []*frontierv1beta1.Price{
				pricePB("default", 100, "usd", "month", "active"),
				pricePB("old", 50, "usd", "month", "inactive"), // retired, must not appear in the export
			}, "f1"),
			billingProductPB("p2", "seat", "Seats", "per_seat", []*frontierv1beta1.Price{
				pricePB("monthly", 15000, "usd", "month", "active"),
			}, "f1", "f2"),
		}}
		r := NewBillingProductReconciler(api, "")

		exported, err := r.Export(ctx)
		assert.NoError(t, err)
		specs, ok := exported.([]BillingProductSpec)
		if assert.True(t, ok) {
			// the inactive price is left out of the export
			for _, p := range specs {
				for _, pr := range p.Prices {
					assert.NotEqual(t, "old", pr.Name)
				}
			}
		}

		specBytes, err := yaml.Marshal(exported)
		assert.NoError(t, err)
		rep, err := r.Reconcile(ctx, specBytes, true)
		assert.NoError(t, err)
		assert.Empty(t, rep.Planned, "reconciling an export must plan no changes")
	})
}

// TestBillingProductReconciler_ValidatesRequestAgainstProto proves the plan
// reuses the proto's own buf.validate rules, so a bad enum value the kind does
// not re-list is still caught at plan time (dry-run), not left for the apply.
func TestBillingProductReconciler_ValidatesRequestAgainstProto(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects an unknown behavior at plan time", func(t *testing.T) {
		api := &fakeBillingProductAPI{}
		spec := []byte(`
- name: token
  title: Tokens
  behavior: fancy
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
`)
		_, err := NewBillingProductReconciler(api, "").Reconcile(ctx, spec, true)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "behavior")
		}
	})

	t.Run("rejects an unknown interval at plan time", func(t *testing.T) {
		api := &fakeBillingProductAPI{}
		spec := []byte(`
- name: token
  title: Tokens
  behavior: credits
  prices:
    - {name: default, amount: 100, currency: usd, interval: fortnight}
`)
		_, err := NewBillingProductReconciler(api, "").Reconcile(ctx, spec, true)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "interval")
		}
	})

	t.Run("a valid product passes plan-time proto validation", func(t *testing.T) {
		api := &fakeBillingProductAPI{}
		spec := []byte(`
- name: token
  title: Tokens
  behavior: credits
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
`)
		rep, err := NewBillingProductReconciler(api, "").Reconcile(ctx, spec, true)
		assert.NoError(t, err)
		assert.Len(t, rep.Planned, 1)
	})

	t.Run("a one-time price with no interval passes plan-time proto validation", func(t *testing.T) {
		api := &fakeBillingProductAPI{}
		spec := []byte(`
- name: onetime
  title: One time pack
  prices:
    - {name: pack, amount: 5000, currency: usd}
`)
		rep, err := NewBillingProductReconciler(api, "").Reconcile(ctx, spec, true)
		assert.NoError(t, err)
		assert.Len(t, rep.Planned, 1)
	})
}

// TestBillingProductReconciler_ValidateChecksEnumsPerEntry proves Validate() runs
// server-free over every entry, so a bad enum or a missing title fails the whole
// file up front, not only when an op is planned for that entry.
func TestBillingProductReconciler_ValidateChecksEnumsPerEntry(t *testing.T) {
	r := NewBillingProductReconciler(nil, "")

	badBehavior := []byte(`
- name: token
  title: Tokens
  behavior: fancy
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
`)
	if err := r.Validate(badBehavior); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "behavior")
	}

	missingTitle := []byte(`
- name: token
  behavior: credits
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
`)
	if err := r.Validate(missingTitle); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "must have a title")
	}

	good := []byte(`
- name: token
  title: Tokens
  behavior: credits
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
`)
	assert.NoError(t, r.Validate(good))
}

// mergeFakeBillingProductAPI models the server's create and merge-update
// semantics closely enough to prove the reconciler converges against them:
// Create rewrites behavior for credit products and defaults each price's
// metered_aggregate to "sum"; Update keeps title/description/metadata when empty
// and credit/seat when zero, never changes behavior, and converges the price
// list only when it is non-empty.
type mergeFakeBillingProductAPI struct {
	byID   map[string]*frontierv1beta1.Product
	nextID int
}

func newMergeFake() *mergeFakeBillingProductAPI {
	return &mergeFakeBillingProductAPI{byID: map[string]*frontierv1beta1.Product{}}
}

func (f *mergeFakeBillingProductAPI) ListProducts(_ context.Context, _ *connect.Request[frontierv1beta1.ListProductsRequest]) (*connect.Response[frontierv1beta1.ListProductsResponse], error) {
	var ps []*frontierv1beta1.Product
	for _, p := range f.byID {
		ps = append(ps, p)
	}
	return connect.NewResponse(&frontierv1beta1.ListProductsResponse{Products: ps}), nil
}

func (f *mergeFakeBillingProductAPI) CreateProduct(_ context.Context, req *connect.Request[frontierv1beta1.CreateProductRequest]) (*connect.Response[frontierv1beta1.CreateProductResponse], error) {
	b := req.Msg.GetBody()
	f.nextID++
	id := fmt.Sprintf("prod-%d", f.nextID)
	behavior := b.GetBehavior()
	if b.GetBehaviorConfig().GetCreditAmount() > 0 {
		behavior = "credits" // the server rewrites behavior for credit products
	}
	if behavior == "" {
		behavior = "basic"
	}
	p := &frontierv1beta1.Product{
		Id: id, Name: b.GetName(), Title: b.GetTitle(), Description: b.GetDescription(),
		Behavior: behavior, BehaviorConfig: b.GetBehaviorConfig(), Metadata: b.GetMetadata(),
	}
	for _, pr := range b.GetPrices() {
		p.Prices = append(p.Prices, createdPrice(pr, true)) // create path defaults metered_aggregate to "sum"
	}
	for _, ft := range b.GetFeatures() {
		p.Features = append(p.Features, &frontierv1beta1.Feature{Name: ft.GetName()})
	}
	f.byID[id] = p
	return connect.NewResponse(&frontierv1beta1.CreateProductResponse{Product: p}), nil
}

func (f *mergeFakeBillingProductAPI) UpdateProduct(_ context.Context, req *connect.Request[frontierv1beta1.UpdateProductRequest]) (*connect.Response[frontierv1beta1.UpdateProductResponse], error) {
	p := f.byID[req.Msg.GetId()]
	b := req.Msg.GetBody()
	// title, description, and config are written as given (a reset when omitted);
	// metadata is kept when empty; behavior is never changed on update.
	p.Title = b.GetTitle()
	p.Description = b.GetDescription()
	if b.GetMetadata() != nil {
		p.Metadata = b.GetMetadata()
	}
	nb := b.GetBehaviorConfig()
	p.BehaviorConfig = &frontierv1beta1.Product_BehaviorConfig{
		CreditAmount: nb.GetCreditAmount(),
		SeatLimit:    nb.GetSeatLimit(),
		MinQuantity:  nb.GetMinQuantity(),
		MaxQuantity:  nb.GetMaxQuantity(),
	}
	p.Features = nil
	for _, ft := range b.GetFeatures() {
		p.Features = append(p.Features, &frontierv1beta1.Feature{Name: ft.GetName()})
	}
	if len(b.GetPrices()) > 0 { // the server ignores an empty price list
		f.convergePrices(p, b.GetPrices())
	}
	return connect.NewResponse(&frontierv1beta1.UpdateProductResponse{Product: p}), nil
}

func (f *mergeFakeBillingProductAPI) convergePrices(p *frontierv1beta1.Product, desired []*frontierv1beta1.Price) {
	byName := map[string]*frontierv1beta1.Price{}
	for _, pr := range p.GetPrices() {
		byName[strings.ToLower(pr.GetName())] = pr
	}
	want := map[string]bool{}
	for _, pr := range desired {
		name := strings.ToLower(pr.GetName())
		want[name] = true
		existing, ok := byName[name]
		if !ok {
			np := createdPrice(pr, false) // a new name is created; update path leaves metered_aggregate empty
			p.Prices = append(p.Prices, np)
			byName[name] = np
			continue
		}
		existing.State = "active" // a retired name listed again is reactivated in place, not duplicated
	}
	for _, pr := range p.GetPrices() {
		if pr.GetState() != "inactive" && !want[strings.ToLower(pr.GetName())] {
			pr.State = "inactive"
		}
	}
}

// createdPrice models how a price is stored. defaultAggregate mirrors the create
// path (SetDefaults fills metered_aggregate "sum"); the update path leaves it as
// given, which is the difference the reconciler's normalization has to smooth.
func createdPrice(pr *frontierv1beta1.Price, defaultAggregate bool) *frontierv1beta1.Price {
	agg := pr.GetMeteredAggregate()
	if agg == "" && defaultAggregate {
		agg = "sum"
	}
	usage := pr.GetUsageType()
	if usage == "" {
		usage = "licensed"
	}
	scheme := pr.GetBillingScheme()
	if scheme == "" {
		scheme = "flat"
	}
	return &frontierv1beta1.Price{
		Name: pr.GetName(), Amount: pr.GetAmount(), Currency: pr.GetCurrency(), Interval: pr.GetInterval(),
		UsageType: usage, BillingScheme: scheme, MeteredAggregate: agg, State: "active",
	}
}

func reconcileTwice(t *testing.T, api *mergeFakeBillingProductAPI, spec []byte) {
	t.Helper()
	r := NewBillingProductReconciler(api, "")
	if _, err := r.Reconcile(context.Background(), spec, false); err != nil {
		t.Fatalf("first reconcile: %v", err)
	}
	rep, err := r.Reconcile(context.Background(), spec, true)
	if err != nil {
		t.Fatalf("second reconcile: %v", err)
	}
	assert.Empty(t, rep.Planned, "a second reconcile of the same spec must plan no changes")
}

// TestBillingProductReconciler_Converges runs each spec against a fake that
// applies the real server's merge semantics, then re-plans, and asserts the
// second run is a no-op. This is what catches the loops the naive fake hid.
func TestBillingProductReconciler_Converges(t *testing.T) {
	t.Run("a credit product with omitted behavior converges after its own create", func(t *testing.T) {
		reconcileTwice(t, newMergeFake(), []byte(`
- name: tokens
  title: Tokens
  config: {credit_amount: 1, min_quantity: 1, max_quantity: 100000}
  prices:
    - {name: default, amount: 100, currency: usd}
`))
	})

	t.Run("a per-seat product with a price converges", func(t *testing.T) {
		reconcileTwice(t, newMergeFake(), []byte(`
- name: seat
  title: Seat
  behavior: per_seat
  prices:
    - {name: monthly, amount: 15000, currency: usd, interval: month}
`))
	})

	t.Run("a metered price with an omitted aggregate converges", func(t *testing.T) {
		reconcileTwice(t, newMergeFake(), []byte(`
- name: usage
  title: Usage
  prices:
    - {name: metered, amount: 1, currency: usd, interval: month, usage_type: metered}
`))
	})

	t.Run("an update to a title still converges", func(t *testing.T) {
		api := newMergeFake()
		r := NewBillingProductReconciler(api, "")
		base := []byte("- {name: tokens, title: Tokens, config: {credit_amount: 1}, prices: [{name: default, amount: 100, currency: usd}]}")
		if _, err := r.Reconcile(context.Background(), base, false); err != nil {
			t.Fatalf("create: %v", err)
		}
		changed := []byte("- {name: tokens, title: Renamed, config: {credit_amount: 1}, prices: [{name: default, amount: 100, currency: usd}]}")
		if _, err := r.Reconcile(context.Background(), changed, false); err != nil {
			t.Fatalf("update: %v", err)
		}
		rep, err := r.Reconcile(context.Background(), changed, true)
		assert.NoError(t, err)
		assert.Empty(t, rep.Planned, "reconciling the applied spec again must plan no changes")
	})

	t.Run("a one-time price with no interval converges", func(t *testing.T) {
		reconcileTwice(t, newMergeFake(), []byte(`
- name: onetime
  title: One time pack
  prices:
    - {name: pack, amount: 5000, currency: usd}
`))
	})

	twoPrices := []byte(`
- name: tokens
  title: Tokens
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
    - {name: bonus, amount: 200, currency: usd, interval: month}
`)
	onePrice := []byte(`
- name: tokens
  title: Tokens
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
`)

	t.Run("dropping a price retires it and converges", func(t *testing.T) {
		api := newMergeFake()
		r := NewBillingProductReconciler(api, "")
		if _, err := r.Reconcile(context.Background(), twoPrices, false); err != nil {
			t.Fatalf("create with two prices: %v", err)
		}
		if _, err := r.Reconcile(context.Background(), onePrice, false); err != nil {
			t.Fatalf("drop a price: %v", err)
		}
		rep, err := r.Reconcile(context.Background(), onePrice, true)
		assert.NoError(t, err)
		assert.Empty(t, rep.Planned, "after a price is dropped, re-planning the same spec is a no-op")
	})

	t.Run("a retired price listed again with the same fields is reactivated and converges", func(t *testing.T) {
		api := newMergeFake()
		r := NewBillingProductReconciler(api, "")
		if _, err := r.Reconcile(context.Background(), twoPrices, false); err != nil {
			t.Fatalf("create with two prices: %v", err)
		}
		if _, err := r.Reconcile(context.Background(), onePrice, false); err != nil {
			t.Fatalf("drop bonus: %v", err)
		}
		// list bonus again with identical fields; the server reactivates it rather
		// than rejecting the reused name, and the plan must reflect that.
		if _, err := r.Reconcile(context.Background(), twoPrices, false); err != nil {
			t.Fatalf("reactivate bonus: %v", err)
		}
		rep, err := r.Reconcile(context.Background(), twoPrices, true)
		assert.NoError(t, err)
		assert.Empty(t, rep.Planned, "after a retired price is reactivated, re-planning the same spec is a no-op")
	})

	t.Run("removing a feature converges", func(t *testing.T) {
		api := newMergeFake()
		r := NewBillingProductReconciler(api, "")
		withTwo := []byte(`
- name: tokens
  title: Tokens
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
  features:
    - {name: f1}
    - {name: f2}
`)
		withOne := []byte(`
- name: tokens
  title: Tokens
  prices:
    - {name: default, amount: 100, currency: usd, interval: month}
  features:
    - {name: f1}
`)
		if _, err := r.Reconcile(context.Background(), withTwo, false); err != nil {
			t.Fatalf("create with two features: %v", err)
		}
		if _, err := r.Reconcile(context.Background(), withOne, false); err != nil {
			t.Fatalf("remove a feature: %v", err)
		}
		rep, err := r.Reconcile(context.Background(), withOne, true)
		assert.NoError(t, err)
		assert.Empty(t, rep.Planned, "after a feature is removed, re-planning the same spec is a no-op")
	})
}
