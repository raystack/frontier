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
	if b.GetTitle() != "" {
		p.Title = b.GetTitle()
	}
	if b.GetDescription() != "" {
		p.Description = b.GetDescription()
	}
	if b.GetMetadata() != nil {
		p.Metadata = b.GetMetadata()
	}
	// behavior is never changed on update
	cfg := p.GetBehaviorConfig()
	if cfg == nil {
		cfg = &frontierv1beta1.Product_BehaviorConfig{}
	}
	nb := b.GetBehaviorConfig()
	if nb.GetCreditAmount() > 0 {
		cfg.CreditAmount = nb.GetCreditAmount()
	}
	if nb.GetSeatLimit() > 0 {
		cfg.SeatLimit = nb.GetSeatLimit()
	}
	cfg.MinQuantity = nb.GetMinQuantity()
	cfg.MaxQuantity = nb.GetMaxQuantity()
	p.BehaviorConfig = cfg
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
	want := map[string]bool{}
	have := map[string]bool{}
	for _, pr := range p.GetPrices() {
		if pr.GetState() != "inactive" {
			have[strings.ToLower(pr.GetName())] = true
		}
	}
	for _, pr := range desired {
		name := strings.ToLower(pr.GetName())
		want[name] = true
		if !have[name] {
			p.Prices = append(p.Prices, createdPrice(pr, false)) // update path leaves metered_aggregate empty
		}
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
}
