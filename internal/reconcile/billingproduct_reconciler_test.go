package reconcile

import (
	"context"
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
