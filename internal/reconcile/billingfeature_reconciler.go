package reconcile

import (
	"context"
	"fmt"
	"unicode/utf8"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type BillingFeatureAPI interface {
	ListFeatures(context.Context, *connect.Request[frontierv1beta1.ListFeaturesRequest]) (*connect.Response[frontierv1beta1.ListFeaturesResponse], error)
	CreateFeature(context.Context, *connect.Request[frontierv1beta1.CreateFeatureRequest]) (*connect.Response[frontierv1beta1.CreateFeatureResponse], error)
	UpdateFeature(context.Context, *connect.Request[frontierv1beta1.UpdateFeatureRequest]) (*connect.Response[frontierv1beta1.UpdateFeatureResponse], error)
}

type BillingFeatureReconciler struct {
	client BillingFeatureAPI
	header string
}

func NewBillingFeatureReconciler(client BillingFeatureAPI, header string) *BillingFeatureReconciler {
	return &BillingFeatureReconciler{client: client, header: header}
}

func (r *BillingFeatureReconciler) Kind() string { return KindBillingFeature }

func (r *BillingFeatureReconciler) Validate(spec []byte) error {
	var specs []BillingFeatureSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindBillingFeature, err)
	}
	if err := checkBillingFeatureSpecs(specs); err != nil {
		return err
	}
	for _, s := range specs {
		if _, err := billingFeatureRequest(billingFeatureOp{action: opAdd, spec: s}); err != nil {
			return err
		}
	}
	return nil
}

func (r *BillingFeatureReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []BillingFeatureSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindBillingFeature, err)
	}

	all, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}
	current, outOfScope := splitOutOfScopeFeatures(all)
	if err := checkOutOfScopeFeatures(specs, outOfScope); err != nil {
		return Report{}, err
	}

	ops, err := diffBillingFeatures(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindBillingFeature, DryRun: dryRun}
	for _, op := range ops {
		if _, err := billingFeatureRequest(op); err != nil {
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

func (r *BillingFeatureReconciler) fetchCurrent(ctx context.Context) ([]currentBillingFeature, error) {
	resp, err := r.client.ListFeatures(ctx, authReq(&frontierv1beta1.ListFeaturesRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}
	var current []currentBillingFeature
	for _, f := range resp.Msg.GetFeatures() {
		current = append(current, currentBillingFeature{
			ID:         f.GetId(),
			Name:       f.GetName(),
			Title:      f.GetTitle(),
			Metadata:   f.GetMetadata().AsMap(),
			ProductIDs: f.GetProductIds(),
		})
	}
	return current, nil
}

func splitOutOfScopeFeatures(all []currentBillingFeature) ([]currentBillingFeature, []string) {
	var inScope []currentBillingFeature
	var outOfScope []string
	for _, c := range all {
		if utf8.RuneCountInString(c.Name) < 3 {
			outOfScope = append(outOfScope, c.Name)
			continue
		}
		inScope = append(inScope, c)
	}
	return inScope, outOfScope
}

func checkOutOfScopeFeatures(specs []BillingFeatureSpec, outOfScope []string) error {
	set := make(map[string]struct{}, len(outOfScope))
	for _, n := range outOfScope {
		set[n] = struct{}{}
	}
	for _, s := range specs {
		if _, ok := set[s.Name]; ok {
			return fmt.Errorf("feature %q is out of scope for this kind: the server holds it with a name shorter than three characters, so it cannot be managed from the file; remove the entry and manage it by hand", s.Name)
		}
	}
	return nil
}

func (r *BillingFeatureReconciler) apply(ctx context.Context, op billingFeatureOp) error {
	req, err := billingFeatureRequest(op)
	if err != nil {
		return err
	}
	switch req := req.(type) {
	case *frontierv1beta1.CreateFeatureRequest:
		_, err = r.client.CreateFeature(ctx, authReq(req, r.header))
	case *frontierv1beta1.UpdateFeatureRequest:
		_, err = r.client.UpdateFeature(ctx, authReq(req, r.header))
	}
	return err
}

func billingFeatureRequest(op billingFeatureOp) (proto.Message, error) {
	metadata, err := structpb.NewStruct(op.spec.Metadata)
	if err != nil {
		return nil, fmt.Errorf("feature %q metadata: %w", op.spec.Name, err)
	}
	var req proto.Message
	switch op.action {
	case opAdd:
		req = &frontierv1beta1.CreateFeatureRequest{Body: &frontierv1beta1.FeatureRequestBody{
			Name:     op.spec.Name,
			Title:    op.spec.Title,
			Metadata: metadata,
		}}
	case opUpdate:
		req = &frontierv1beta1.UpdateFeatureRequest{Id: op.current.ID, Body: &frontierv1beta1.FeatureRequestBody{
			Name:       op.current.Name,
			Title:      op.spec.Title,
			Metadata:   metadata,
			ProductIds: op.current.ProductIDs,
		}}
	default:
		return nil, fmt.Errorf("unknown op action %q", op.action)
	}
	if err := protovalidate.Validate(req); err != nil {
		return nil, fmt.Errorf("feature %q: %w", op.spec.Name, err)
	}
	return req, nil
}
