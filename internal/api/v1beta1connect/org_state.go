package v1beta1connect

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/serviceuser"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
)

// EnsurePATOrgActive blocks personal access token management when the token's
// org is disabled. The token is looked up scoped to the caller, so the
// outcome for someone else's token id is the same as for an unknown one.
func (h *ConnectHandler) EnsurePATOrgActive(ctx context.Context, patID string) error {
	if isSuperUser, ok := authenticate.GetSuperUserFromContext(ctx); ok && isSuperUser {
		return nil
	}
	if _, err := uuid.Parse(patID); err != nil {
		// the handler keeps its own behavior for malformed ids
		return nil
	}
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return err
	}
	if principal.User == nil {
		// sessions and PAT principals both carry the owning user; only
		// service accounts land here and the handler rejects them itself
		return nil
	}
	token, err := h.userPATService.Get(ctx, principal.User.ID, patID)
	if err != nil {
		if errors.Is(err, paterrors.ErrNotFound) || errors.Is(err, paterrors.ErrDisabled) {
			// unknown or not owned by the caller, or the PAT feature is
			// off: the handler keeps its own behavior
			return nil
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("EnsurePATOrgActive: pat_id=%s: %w", patID, err))
	}
	return h.ensureOrgEnabled(ctx, token.OrgID)
}

// ensureObjectOrgActive blocks a request whose target belongs to a disabled
// organization. It must run only after the permission check has passed, so
// callers without access keep getting PermissionDenied. Platform superusers
// skip the gate so they can inspect and re-enable disabled orgs.
func (h *ConnectHandler) ensureObjectOrgActive(ctx context.Context, object relation.Object) error {
	// the flag is set per request by the authentication interceptor; when
	// absent, the gate applies
	if isSuperUser, ok := authenticate.GetSuperUserFromContext(ctx); ok && isSuperUser {
		return nil
	}
	orgID, found, err := h.resolveObjectOrg(ctx, object)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return h.ensureOrgEnabled(ctx, orgID)
}

// resolveObjectOrg maps an authorization target to the org that owns it.
// found=false means there is no org to check: the object is platform-level,
// or its row is not visible here. In that case the gate steps aside and the
// handler keeps its own not-found behavior, so flows like re-enabling a
// disabled project or group keep working.
func (h *ConnectHandler) resolveObjectOrg(ctx context.Context, object relation.Object) (string, bool, error) {
	switch object.Namespace {
	case schema.OrganizationNamespace:
		return object.ID, true, nil

	case schema.ProjectNamespace:
		proj, err := h.projectService.Get(ctx, object.ID)
		if err != nil {
			if errors.Is(err, project.ErrNotExist) || errors.Is(err, project.ErrInvalidUUID) || errors.Is(err, project.ErrInvalidID) {
				return "", false, nil
			}
			return "", false, connect.NewError(connect.CodeInternal, fmt.Errorf("resolveObjectOrg: project_id=%s: %w", object.ID, err))
		}
		return proj.Organization.ID, true, nil

	case schema.GroupNamespace:
		// fetch with disabled groups included: a disabled group in an enabled
		// org must stay reachable so it can be re-enabled
		grps, err := h.groupService.GetByIDs(ctx, []string{object.ID}, group.Filter{IncludeDisabled: true})
		if err != nil {
			if errors.Is(err, group.ErrNotExist) || errors.Is(err, group.ErrInvalidID) || errors.Is(err, group.ErrInvalidUUID) {
				return "", false, nil
			}
			return "", false, connect.NewError(connect.CodeInternal, fmt.Errorf("resolveObjectOrg: group_id=%s: %w", object.ID, err))
		}
		if len(grps) == 0 {
			return "", false, nil
		}
		return grps[0].OrganizationID, true, nil

	case schema.InvitationNamespace:
		invID, err := uuid.Parse(object.ID)
		if err != nil {
			return "", false, nil
		}
		inv, err := h.invitationService.Get(ctx, invID)
		if err != nil {
			if errors.Is(err, invitation.ErrNotFound) {
				return "", false, nil
			}
			return "", false, connect.NewError(connect.CodeInternal, fmt.Errorf("resolveObjectOrg: invitation_id=%s: %w", object.ID, err))
		}
		return inv.OrgID, true, nil

	case schema.ServiceUserPrincipal:
		svUser, err := h.serviceUserService.Get(ctx, object.ID)
		if err != nil {
			if errors.Is(err, serviceuser.ErrNotExist) || errors.Is(err, serviceuser.ErrInvalidID) {
				return "", false, nil
			}
			return "", false, connect.NewError(connect.CodeInternal, fmt.Errorf("resolveObjectOrg: service_user_id=%s: %w", object.ID, err))
		}
		return svUser.OrgID, true, nil
	}

	// the rest of the predefined app/* objects either have no owning org
	// (platform, user) or reach here only as policy targets (role, pat,
	// rolebinding). The RPCs that manage the org-owned ones authorize
	// against the org object itself, so the gate already covers them there.
	if strings.HasPrefix(object.Namespace, schema.DefaultNamespace+"/") {
		return "", false, nil
	}

	// custom resource: resource row -> project -> org
	res, err := h.resourceService.Get(ctx, object.ID)
	if err != nil {
		if errors.Is(err, resource.ErrNotExist) || errors.Is(err, resource.ErrInvalidID) || errors.Is(err, resource.ErrInvalidUUID) || errors.Is(err, resource.ErrInvalidURN) {
			return "", false, nil
		}
		return "", false, connect.NewError(connect.CodeInternal, fmt.Errorf("resolveObjectOrg: resource_id=%s: %w", object.ID, err))
	}
	proj, err := h.projectService.Get(ctx, res.ProjectID)
	if err != nil {
		if errors.Is(err, project.ErrNotExist) || errors.Is(err, project.ErrInvalidUUID) || errors.Is(err, project.ErrInvalidID) {
			return "", false, nil
		}
		return "", false, connect.NewError(connect.CodeInternal, fmt.Errorf("resolveObjectOrg: resource_id=%s project_id=%s: %w", object.ID, res.ProjectID, err))
	}
	return proj.Organization.ID, true, nil
}

// ensureOrgEnabled blocks with FailedPrecondition when the org is disabled.
// It backs both the object gate and the PAT org check above.
func (h *ConnectHandler) ensureOrgEnabled(ctx context.Context, orgID string) error {
	if _, err := h.orgService.Get(ctx, orgID); err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return connect.NewError(connect.CodeInternal, fmt.Errorf("ensureOrgEnabled: org_id=%s: %w", orgID, err))
		}
	}
	return nil
}
