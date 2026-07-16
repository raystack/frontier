package reconcile

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// PlatformUserAPI is the subset of the admin API the platform-user reconciler needs.
// frontierv1beta1connect.AdminServiceClient satisfies it.
type PlatformUserAPI interface {
	ListPlatformUsers(context.Context, *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error)
	AddPlatformUser(context.Context, *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error)
	RemovePlatformUser(context.Context, *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error)
}

// PlatformUserReconciler makes platform admins and members match the desired spec.
type PlatformUserReconciler struct {
	client PlatformUserAPI
	header string // "key:value" auth header applied to each request (may be empty)
}

func NewPlatformUserReconciler(client PlatformUserAPI, header string) *PlatformUserReconciler {
	return &PlatformUserReconciler{client: client, header: header}
}

func (r *PlatformUserReconciler) Kind() string { return KindPlatformUser }

func (r *PlatformUserReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []PlatformUserSpec
	if err := yaml.Unmarshal(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindPlatformUser, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffPlatformUsers(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindPlatformUser, DryRun: dryRun}
	for _, op := range ops {
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

// Export returns the current platform users as a desired-state spec: one entry
// per (principal, relation), users referenced by email when they have one.
// Entries are sorted so repeated exports produce identical files.
func (r *PlatformUserReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	specs := make([]PlatformUserSpec, 0, len(current))
	for _, p := range current {
		ref := p.ID
		if p.Type == principalTypeUser && p.Email != "" {
			ref = p.Email
		}
		for _, rel := range platformRelationOrder {
			if _, ok := p.Relations[rel]; ok {
				specs = append(specs, PlatformUserSpec{Type: p.Type, Ref: ref, Relation: rel})
			}
		}
	}
	sort.Slice(specs, func(i, j int) bool {
		a, b := specs[i], specs[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Ref != b.Ref {
			return a.Ref < b.Ref
		}
		return a.Relation < b.Relation
	})
	return specs, nil
}

func (r *PlatformUserReconciler) fetchCurrent(ctx context.Context) ([]platformPrincipal, error) {
	resp, err := r.client.ListPlatformUsers(ctx, authReq(&frontierv1beta1.ListPlatformUsersRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list platform users: %w", err)
	}
	var current []platformPrincipal
	for _, u := range resp.Msg.GetUsers() {
		current = append(current, platformPrincipal{
			Type:      principalTypeUser,
			ID:        u.GetId(),
			Email:     u.GetEmail(),
			Relations: relationsFromMetadata(u.GetMetadata()),
		})
	}
	for _, su := range resp.Msg.GetServiceusers() {
		// Never reconcile the bootstrap SA — it is server-managed and removal-blocked.
		if su.GetId() == schema.BootstrapServiceUserID {
			continue
		}
		current = append(current, platformPrincipal{
			Type:      principalTypeServiceUser,
			ID:        su.GetId(),
			Relations: relationsFromMetadata(su.GetMetadata()),
		})
	}
	return current, nil
}

func (r *PlatformUserReconciler) apply(ctx context.Context, op Op) error {
	userID, serviceUserID := op.principalIDs()
	switch op.Action {
	case opAdd:
		_, err := r.client.AddPlatformUser(ctx, authReq(&frontierv1beta1.AddPlatformUserRequest{
			UserId: userID, ServiceuserId: serviceUserID, Relation: op.Relation,
		}, r.header))
		return err
	case opRemove:
		// relation-selective removal (proton #489)
		_, err := r.client.RemovePlatformUser(ctx, authReq(&frontierv1beta1.RemovePlatformUserRequest{
			UserId: userID, ServiceuserId: serviceUserID, Relation: op.Relation,
		}, r.header))
		return err
	default:
		return fmt.Errorf("unknown op action %q", op.Action)
	}
}

// relationsFromMetadata reads the platform relations ListPlatformUsers records in a
// principal's metadata: the "relations" list, falling back to "relation".
func relationsFromMetadata(md *structpb.Struct) map[string]struct{} {
	rels := map[string]struct{}{}
	if md == nil {
		return rels
	}
	fields := md.GetFields()
	if v, ok := fields["relations"]; ok {
		for _, item := range v.GetListValue().GetValues() {
			if name := item.GetStringValue(); name != "" {
				rels[name] = struct{}{}
			}
		}
	}
	if len(rels) == 0 {
		if v, ok := fields["relation"]; ok {
			if name := v.GetStringValue(); name != "" {
				rels[name] = struct{}{}
			}
		}
	}
	return rels
}

// authReq builds a connect request and applies the optional "key:value" auth header.
func authReq[T any](msg *T, header string) *connect.Request[T] {
	req := connect.NewRequest(msg)
	if header != "" {
		if k, v, ok := strings.Cut(header, ":"); ok {
			req.Header().Set(strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}
	return req
}
