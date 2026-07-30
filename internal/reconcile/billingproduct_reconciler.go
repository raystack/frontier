package reconcile

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// BillingProductAPI is the API subset the billing product reconciler needs.
// Billing products live on the FrontierService. Reads come from ListProducts
// (which returns each product with its prices and features), and writes go
// through CreateProduct and UpdateProduct. UpdateProduct converges a product's
// prices, so the reconciler sends the whole desired product and lets the server
// add, keep, or deactivate each price.
type BillingProductAPI interface {
	ListProducts(context.Context, *connect.Request[frontierv1beta1.ListProductsRequest]) (*connect.Response[frontierv1beta1.ListProductsResponse], error)
	CreateProduct(context.Context, *connect.Request[frontierv1beta1.CreateProductRequest]) (*connect.Response[frontierv1beta1.CreateProductResponse], error)
	UpdateProduct(context.Context, *connect.Request[frontierv1beta1.UpdateProductRequest]) (*connect.Response[frontierv1beta1.UpdateProductResponse], error)
}

// BillingProductReconciler makes billing products match the desired spec. The
// product name is the identity; title, description, behavior, config, prices,
// and features are the managed fields. A product missing from the file fails the
// plan, because there is no API to remove a product.
type BillingProductReconciler struct {
	client BillingProductAPI
	header string
}

func NewBillingProductReconciler(client BillingProductAPI, header string) *BillingProductReconciler {
	return &BillingProductReconciler{client: client, header: header}
}

func (r *BillingProductReconciler) Kind() string { return KindBillingProduct }

func (r *BillingProductReconciler) Validate(spec []byte) error {
	var specs []BillingProductSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindBillingProduct, err)
	}
	_, err := normalizeBillingProductSpecs(specs)
	return err
}

func (r *BillingProductReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []BillingProductSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindBillingProduct, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffBillingProducts(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindBillingProduct, DryRun: dryRun}
	for _, op := range ops {
		// validate the request the apply would send against the proto's own rules,
		// so a value the server would reject, like an unknown behavior or interval,
		// fails the plan instead of the apply.
		if err := validateBillingProductRequest(op); err != nil {
			return Report{}, fmt.Errorf("plan %s: %w", op, err)
		}
		rep.Planned = append(rep.Planned, op.String())
	}
	if dryRun {
		return rep, nil
	}
	for _, op := range ops {
		if err := r.apply(ctx, op); err != nil {
			return rep, fmt.Errorf("apply [%s]: %w", op, err)
		}
		rep.Applied++
	}
	return rep, nil
}

// Export returns the current products as a desired-state spec, sorted by name,
// with each product's prices sorted by name too, so the output is stable. Only
// active prices are written, so reconciling an export leaves the already inactive
// ones alone and plans no changes. Provider ids, timestamps, and price state are
// server-owned and never written.
func (r *BillingProductReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(current, func(i, j int) bool { return current[i].Name < current[j].Name })

	specs := make([]BillingProductSpec, 0, len(current))
	for _, c := range current {
		entry := BillingProductSpec{
			Name:        c.Name,
			Title:       c.Title,
			Description: c.Description,
			Behavior:    c.Behavior,
			Config:      c.Config,
			Metadata:    c.Metadata,
		}
		entry.Prices = append(entry.Prices, c.Prices...)
		sort.Slice(entry.Prices, func(i, j int) bool { return entry.Prices[i].Name < entry.Prices[j].Name })
		for _, name := range uniqueSorted(c.Features) {
			entry.Features = append(entry.Features, BillingFeatureRef{Name: name})
		}
		specs = append(specs, entry)
	}
	return specs, nil
}

func (r *BillingProductReconciler) fetchCurrent(ctx context.Context) ([]currentBillingProduct, error) {
	resp, err := r.client.ListProducts(ctx, authReq(&frontierv1beta1.ListProductsRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	var current []currentBillingProduct
	for _, p := range resp.Msg.GetProducts() {
		var md map[string]any
		if m := p.GetMetadata(); m != nil {
			md = m.AsMap()
		}
		cur := currentBillingProduct{
			ID:          p.GetId(),
			Name:        p.GetName(),
			Title:       p.GetTitle(),
			Description: p.GetDescription(),
			Behavior:    p.GetBehavior(),
			Metadata:    md,
		}
		if cfg := p.GetBehaviorConfig(); cfg != nil {
			cur.Config = BillingProductConfig{
				CreditAmount: cfg.GetCreditAmount(),
				SeatLimit:    cfg.GetSeatLimit(),
				MinQuantity:  cfg.GetMinQuantity(),
				MaxQuantity:  cfg.GetMaxQuantity(),
			}
		}
		for _, price := range p.GetPrices() {
			ps := BillingPriceSpec{
				Name:             price.GetName(),
				Amount:           price.GetAmount(),
				Currency:         price.GetCurrency(),
				Interval:         price.GetInterval(),
				UsageType:        price.GetUsageType(),
				BillingScheme:    price.GetBillingScheme(),
				MeteredAggregate: price.GetMeteredAggregate(),
			}
			// an active price is part of the current desired state; an inactive
			// one is retired. The diff keeps the retired prices apart so a reused
			// name can be told from a change the server would reject, and the
			// export writes only the active ones.
			if billingPriceStateActive(price.GetState()) {
				cur.Prices = append(cur.Prices, ps)
			} else {
				cur.RetiredPrices = append(cur.RetiredPrices, ps)
			}
		}
		for _, f := range p.GetFeatures() {
			cur.Features = append(cur.Features, f.GetName())
		}
		current = append(current, cur)
	}
	return current, nil
}

// validateBillingProductRequest builds the request the apply would send and
// checks it against the proto's buf.validate rules. Reusing the proto's own rules
// means a bad enum value (behavior, interval, usage type, scheme) or a malformed
// field is caught at plan time without this kind re-listing the valid values, and
// the check cannot drift from the server's, since both come from the same
// generated descriptors.
func validateBillingProductRequest(op billingProductOp) error {
	body, err := billingProductBody(op.spec, op.action == opAdd)
	if err != nil {
		return err
	}
	var msg proto.Message
	switch op.action {
	case opAdd:
		msg = &frontierv1beta1.CreateProductRequest{Body: body}
	case opUpdate:
		msg = &frontierv1beta1.UpdateProductRequest{Id: op.id, Body: body}
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
	if err := protovalidate.Validate(msg); err != nil {
		return fmt.Errorf("product %q: %w", op.spec.Name, err)
	}
	return nil
}

func (r *BillingProductReconciler) apply(ctx context.Context, op billingProductOp) error {
	body, err := billingProductBody(op.spec, op.action == opAdd)
	if err != nil {
		return err
	}
	switch op.action {
	case opAdd:
		_, err = r.client.CreateProduct(ctx, authReq(&frontierv1beta1.CreateProductRequest{Body: body}, r.header))
		return err
	case opUpdate:
		_, err = r.client.UpdateProduct(ctx, authReq(&frontierv1beta1.UpdateProductRequest{Id: op.id, Body: body}, r.header))
		return err
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
}

// billingProductBody builds the request body for a create or update. The whole
// desired product is sent: UpdateProduct converges the prices and features on
// the server, so the reconciler does not compute per-price calls itself.
// Metadata is sent only on create; the server merges it (keep-if-empty) and this
// kind does not diff it, so sending it on update would apply it out of step with
// the plan.
func billingProductBody(s BillingProductSpec, includeMetadata bool) (*frontierv1beta1.ProductRequestBody, error) {
	var md *structpb.Struct
	if includeMetadata && len(s.Metadata) > 0 {
		var err error
		md, err = structpb.NewStruct(s.Metadata)
		if err != nil {
			return nil, fmt.Errorf("build product %q metadata: %w", s.Name, err)
		}
	}
	prices := make([]*frontierv1beta1.Price, 0, len(s.Prices))
	for _, p := range s.Prices {
		// lowercase the case-insensitive fields so a value written in another
		// case, like an interval of "Month", passes the server's validation and
		// matches what the diff compared against.
		prices = append(prices, &frontierv1beta1.Price{
			Name:             p.Name,
			Amount:           p.Amount,
			Currency:         strings.ToLower(p.Currency),
			Interval:         strings.ToLower(p.Interval),
			UsageType:        strings.ToLower(p.UsageType),
			BillingScheme:    strings.ToLower(p.BillingScheme),
			MeteredAggregate: strings.ToLower(p.MeteredAggregate),
		})
	}
	features := make([]*frontierv1beta1.Feature, 0, len(s.Features))
	for _, f := range s.Features {
		features = append(features, &frontierv1beta1.Feature{Name: f.Name})
	}
	return &frontierv1beta1.ProductRequestBody{
		Name:        s.Name,
		Title:       s.Title,
		Description: s.Description,
		Behavior:    s.Behavior,
		Prices:      prices,
		Features:    features,
		BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
			CreditAmount: s.Config.CreditAmount,
			SeatLimit:    s.Config.SeatLimit,
			MinQuantity:  s.Config.MinQuantity,
			MaxQuantity:  s.Config.MaxQuantity,
		},
		Metadata: md,
	}, nil
}

// billingPriceStateActive reports whether a price state read from the server
// counts as active. An empty state is active, matching the database default.
func billingPriceStateActive(state string) bool {
	return state == "" || state == "active"
}
