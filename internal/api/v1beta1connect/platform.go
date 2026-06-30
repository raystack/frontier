package v1beta1connect

import (
	"context"
	"fmt"
	"sort"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) AddPlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error) {
	relationName := req.Msg.GetRelation()

	if !schema.IsPlatformRelation(relationName) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	if req.Msg.GetUserId() != "" {
		if err := h.userService.Sudo(ctx, req.Msg.GetUserId(), relationName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AddPlatformUser.UserSudo: user_id=%s relation=%s: %w", req.Msg.GetUserId(), relationName, err))
		}
	} else if req.Msg.GetServiceuserId() != "" {
		if err := h.serviceUserService.Sudo(ctx, req.Msg.GetServiceuserId(), relationName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AddPlatformUser.ServiceUserSudo: service_user_id=%s relation=%s: %w", req.Msg.GetServiceuserId(), relationName, err))
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.AddPlatformUserResponse{}), nil
}

func (h *ConnectHandler) RemovePlatformUser(ctx context.Context, req *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error) {
	// By default remove the principal from the platform entirely (both admin and
	// member). When a relation is set, scope removal to just that one — e.g. to
	// demote an admin to member. Each UnSudo is a no-op for a relation not held.
	platformRelations := []string{schema.AdminRelationName, schema.MemberRelationName}
	if rel := req.Msg.GetRelation(); rel != "" {
		if !schema.IsPlatformRelation(rel) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		platformRelations = []string{rel}
	}

	if req.Msg.GetUserId() != "" {
		for _, relationName := range platformRelations {
			if err := h.userService.UnSudo(ctx, req.Msg.GetUserId(), relationName); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemovePlatformUser.UserUnSudo: user_id=%s relation=%s: %w", req.Msg.GetUserId(), relationName, err))
			}
		}
	} else if req.Msg.GetServiceuserId() != "" {
		// Protect the config-bootstrapped break-glass SA (well-known id). It is
		// seeded and managed at boot, not via this API, while reconcile is
		// authoritative over service accounts — without this guard an apply (or a
		// stray call) would strip its superuser access until the next restart.
		if req.Msg.GetServiceuserId() == schema.BootstrapServiceUserID {
			return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot remove the bootstrap superuser service account"))
		}
		for _, relationName := range platformRelations {
			if err := h.serviceUserService.UnSudo(ctx, req.Msg.GetServiceuserId(), relationName); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemovePlatformUser.ServiceUserUnSudo: service_user_id=%s relation=%s: %w", req.Msg.GetServiceuserId(), relationName, err))
			}
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	return connect.NewResponse(&frontierv1beta1.RemovePlatformUserResponse{}), nil
}

func (h *ConnectHandler) ListPlatformUsers(ctx context.Context, req *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error) {
	relations, err := h.relationService.List(ctx, relation.Filter{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.ListRelations: %w", err))
	}

	// A principal can hold both admin and member; collect every relation per
	// subject (deduped) so callers see the full set rather than just whichever
	// tuple happened to be listed last. The reconciler relies on this to know
	// exactly which relations to revoke.
	userRelations := platformRelationsBySubject(relations, schema.UserPrincipal)
	serviceUserRelations := platformRelationsBySubject(relations, schema.ServiceUserPrincipal)

	// fetch users
	userIDs := sortedSubjectIDs(userRelations)
	userPBs := make([]*frontierv1beta1.User, 0, len(userIDs))
	if len(userIDs) > 0 {
		users, err := h.userService.GetByIDs(ctx, userIDs)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.GetUsersByIDs: user_ids=%v: %w", userIDs, err))
		}
		for _, u := range users {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			stampPlatformRelations(u.Metadata, userRelations[u.ID])
			userPB, err := transformUserToPB(u)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.TransformUser: entity_id=%s: %w", u.ID, err))
			}
			userPBs = append(userPBs, userPB)
		}
	}

	// fetch service users
	serviceUserIDs := sortedSubjectIDs(serviceUserRelations)
	serviceUserPBs := make([]*frontierv1beta1.ServiceUser, 0, len(serviceUserIDs))
	if len(serviceUserIDs) > 0 {
		serviceUsers, err := h.serviceUserService.GetByIDs(ctx, serviceUserIDs)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.GetServiceUsersByIDs: service_user_ids=%v: %w", serviceUserIDs, err))
		}
		for _, u := range serviceUsers {
			if u.Metadata == nil {
				u.Metadata = make(map[string]any)
			}
			stampPlatformRelations(u.Metadata, serviceUserRelations[u.ID])
			serviceUserPB, err := transformServiceUserToPB(u)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPlatformUsers.TransformServiceUser: entity_id=%s: %w", u.ID, err))
			}
			serviceUserPBs = append(serviceUserPBs, serviceUserPB)
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListPlatformUsersResponse{
		Users:        userPBs,
		Serviceusers: serviceUserPBs,
	}), nil
}

// platformRelationsBySubject groups platform relation names (admin/member) by
// subject id for the given principal namespace, deduped and sorted for stable
// output.
func platformRelationsBySubject(relations []relation.Relation, namespace string) map[string][]string {
	sets := map[string]map[string]struct{}{}
	for _, r := range relations {
		if r.Subject.Namespace != namespace {
			continue
		}
		if sets[r.Subject.ID] == nil {
			sets[r.Subject.ID] = map[string]struct{}{}
		}
		sets[r.Subject.ID][r.RelationName] = struct{}{}
	}
	out := make(map[string][]string, len(sets))
	for id, set := range sets {
		rels := make([]string, 0, len(set))
		for rel := range set {
			rels = append(rels, rel)
		}
		sort.Strings(rels)
		out[id] = rels
	}
	return out
}

// sortedSubjectIDs returns the subject ids in deterministic order.
func sortedSubjectIDs(m map[string][]string) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// stampPlatformRelations records a principal's platform relations in its
// metadata: "relations" carries the full set (consumed by the reconciler) while
// "relation" keeps the first one for backward compatibility.
func stampPlatformRelations(md map[string]any, rels []string) {
	if len(rels) == 0 {
		return
	}
	md["relation"] = rels[0]
	anyRels := make([]any, len(rels))
	for i, r := range rels {
		anyRels[i] = r
	}
	md["relations"] = anyRels
}
