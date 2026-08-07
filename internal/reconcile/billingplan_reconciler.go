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

// BillingPlanAPI is the API subset the billing plan reconciler needs. Plans live
// on the AdminService: reads come from ListAllPlans, which returns every plan
// (active and inactive) with its products, and writes go through CreatePlan and
// UpdatePlan. CreatePlan associates the referenced products by name; UpdatePlan
// changes only a plan's own fields (title, description, credits, trial days,
// state, metadata).
type BillingPlanAPI interface {
	ListAllPlans(context.Context, *connect.Request[frontierv1beta1.ListAllPlansRequest]) (*connect.Response[frontierv1beta1.ListAllPlansResponse], error)
	CreatePlan(context.Context, *connect.Request[frontierv1beta1.CreatePlanRequest]) (*connect.Response[frontierv1beta1.CreatePlanResponse], error)
	UpdatePlan(context.Context, *connect.Request[frontierv1beta1.UpdatePlanRequest]) (*connect.Response[frontierv1beta1.UpdatePlanResponse], error)
}

// BillingPlanReconciler makes billing plans match the desired spec. The plan name
// is the identity; title, description, on_start_credits, trial_days, and state are
// converged through UpdatePlan. Interval and the product set are create-only, so a
// change to either fails the plan. A plan missing from the file fails the plan,
// because there is no API to remove one.
type BillingPlanReconciler struct {
	client BillingPlanAPI
	header string
}

func NewBillingPlanReconciler(client BillingPlanAPI, header string) *BillingPlanReconciler {
	return &BillingPlanReconciler{client: client, header: header}
}

func (r *BillingPlanReconciler) Kind() string { return KindBillingPlan }

func (r *BillingPlanReconciler) Validate(spec []byte) error {
	var specs []BillingPlanSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindBillingPlan, err)
	}
	normalized, err := normalizeBillingPlanSpecs(specs)
	if err != nil {
		return err
	}
	// check every entry against the proto's own rules (the state enum, the
	// interval enum, the non-negative bounds in buf.validate), server-free, so a
	// bad value fails the whole file up front rather than only when an op happens
	// to be planned for that entry.
	for _, s := range normalized {
		if err := validateBillingPlanRequest(billingPlanOp{action: opAdd, spec: s}); err != nil {
			return err
		}
	}
	return nil
}

func (r *BillingPlanReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []BillingPlanSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindBillingPlan, err)
	}

	current, outOfScope, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}
	if err := checkOutOfScopePlans(specs, outOfScope); err != nil {
		return Report{}, err
	}

	ops, err := diffBillingPlans(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindBillingPlan, DryRun: dryRun}
	for _, op := range ops {
		// validate the request the apply would send against the proto's own rules,
		// so a value the server would reject fails the plan instead of the apply.
		if err := validateBillingPlanRequest(op); err != nil {
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

// Export returns the current plans as a desired-state spec, sorted by name, with
// each plan's product references sorted too, so the output is stable. State is
// written (so an inactive plan round-trips and reconciling the export plans no
// change). Ids, timestamps, and metadata are not written: the first two are
// server-owned, and metadata is out of scope for this kind.
func (r *BillingPlanReconciler) Export(ctx context.Context) (any, error) {
	current, _, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(current, func(i, j int) bool { return current[i].Name < current[j].Name })

	specs := make([]BillingPlanSpec, 0, len(current))
	for _, c := range current {
		entry := BillingPlanSpec{
			Name:           c.Name,
			Title:          c.Title,
			Description:    c.Description,
			Interval:       c.Interval,
			OnStartCredits: c.OnStartCredits,
			TrialDays:      c.TrialDays,
			State:          c.State,
		}
		for _, name := range uniqueSorted(c.Products) {
			entry.Products = append(entry.Products, BillingPlanProductRef{Name: name})
		}
		specs = append(specs, entry)
	}
	return specs, nil
}

func (r *BillingPlanReconciler) fetchCurrent(ctx context.Context) ([]currentBillingPlan, []string, error) {
	// ListAllPlans returns every plan, active and inactive, in one response; it
	// does not paginate. The "every plan must appear in the file" rule in
	// diffBillingPlans relies on that: if the API ever paginates, a plan past the
	// first page would look missing here and the plan would try to recreate it.
	resp, err := r.client.ListAllPlans(ctx, authReq(&frontierv1beta1.ListAllPlansRequest{}, r.header))
	if err != nil {
		return nil, nil, fmt.Errorf("list all plans: %w", err)
	}
	var current []currentBillingPlan
	var outOfScope []string
	for _, p := range resp.Msg.GetPlans() {
		// a plan the kind cannot represent is out of scope: a name shorter than
		// three characters or an empty title (the kind requires one). It is neither
		// diffed nor exported, and a file that names it fails the plan.
		if len(p.GetName()) < 3 || p.GetTitle() == "" {
			outOfScope = append(outOfScope, p.GetName())
			continue
		}
		cur := currentBillingPlan{
			ID:             p.GetId(),
			Name:           p.GetName(),
			Title:          p.GetTitle(),
			Description:    p.GetDescription(),
			Interval:       p.GetInterval(),
			OnStartCredits: p.GetOnStartCredits(),
			TrialDays:      p.GetTrialDays(),
			State:          p.GetState(),
			// metadata is out of scope: it is not diffed or exported, but it is
			// carried so an update can re-send it, since UpdatePlan is a full write
			// of the fields it carries and would otherwise clear it.
			Metadata: p.GetMetadata(),
		}
		for _, prod := range p.GetProducts() {
			cur.Products = append(cur.Products, prod.GetName())
		}
		current = append(current, cur)
	}
	return current, outOfScope, nil
}

// checkOutOfScopePlans fails the plan when the file names a plan the server holds
// but this kind cannot represent (an empty title or a name shorter than three
// characters). Such a plan is skipped in fetchCurrent, so without this check the
// diff would see the file entry as new and plan a create the apply would reject on
// the unique name; failing here keeps the plan honest.
func checkOutOfScopePlans(specs []BillingPlanSpec, outOfScope []string) error {
	if len(outOfScope) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(outOfScope))
	for _, n := range outOfScope {
		set[strings.ToLower(strings.TrimSpace(n))] = struct{}{}
	}
	for _, s := range specs {
		if _, ok := set[strings.ToLower(strings.TrimSpace(s.Name))]; ok {
			return fmt.Errorf("plan %q is out of scope for this kind: the server holds it with an empty title or a name shorter than three characters, so it cannot be managed from the file; remove the entry and manage it by hand", s.Name)
		}
	}
	return nil
}

// validateBillingPlanRequest builds the request the apply would send and checks it
// against the proto's buf.validate rules. Reusing the proto's own rules means a
// bad state or interval, or a negative credit or trial-days value, is caught at
// plan time, and the check cannot drift from the server's, since both come from
// the same generated descriptors.
func validateBillingPlanRequest(op billingPlanOp) error {
	var msg proto.Message
	switch op.action {
	case opAdd:
		msg = &frontierv1beta1.CreatePlanRequest{Body: billingPlanCreateBody(op.spec)}
	case opUpdate:
		msg = &frontierv1beta1.UpdatePlanRequest{Id: op.id, Body: billingPlanUpdateBody(op.spec, op.metadata)}
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
	if err := protovalidate.Validate(msg); err != nil {
		return fmt.Errorf("plan %q: %w", op.spec.Name, err)
	}
	return nil
}

func (r *BillingPlanReconciler) apply(ctx context.Context, op billingPlanOp) error {
	switch op.action {
	case opAdd:
		_, err := r.client.CreatePlan(ctx, authReq(&frontierv1beta1.CreatePlanRequest{Body: billingPlanCreateBody(op.spec)}, r.header))
		return err
	case opUpdate:
		_, err := r.client.UpdatePlan(ctx, authReq(&frontierv1beta1.UpdatePlanRequest{Id: op.id, Body: billingPlanUpdateBody(op.spec, op.metadata)}, r.header))
		return err
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
}

// billingPlanCreateBody builds the CreatePlan body. The whole desired plan is
// sent, including its product references by name; CreatePlan's upsert associates
// them. An omitted state defaults to active (as the docs, the diff, and the store
// all treat it), so a file that leaves state out passes the proto's state enum.
// Metadata is out of scope, so it is not set on create.
func billingPlanCreateBody(s BillingPlanSpec) *frontierv1beta1.PlanRequestBody {
	products := make([]*frontierv1beta1.Product, 0, len(s.Products))
	for _, p := range s.Products {
		products = append(products, &frontierv1beta1.Product{Name: p.Name})
	}
	return &frontierv1beta1.PlanRequestBody{
		Name:           s.Name,
		Title:          s.Title,
		Description:    s.Description,
		Interval:       strings.ToLower(s.Interval),
		OnStartCredits: s.OnStartCredits,
		TrialDays:      s.TrialDays,
		State:          normalizeBillingPlanState(s.State),
		Products:       products,
	}
}

// billingPlanUpdateBody builds the UpdatePlan body. It carries only the fields
// UpdatePlan can change; interval, name, and products are create-only. An omitted
// state defaults to active, the same as on create. Metadata is out of scope, but
// UpdatePlan is a full write of the fields it carries, so the current metadata is
// re-sent to keep it rather than clear it.
func billingPlanUpdateBody(s BillingPlanSpec, metadata *structpb.Struct) *frontierv1beta1.UpdatePlanRequestBody {
	return &frontierv1beta1.UpdatePlanRequestBody{
		Title:          s.Title,
		Description:    s.Description,
		OnStartCredits: s.OnStartCredits,
		TrialDays:      s.TrialDays,
		State:          normalizeBillingPlanState(s.State),
		Metadata:       metadata,
	}
}
