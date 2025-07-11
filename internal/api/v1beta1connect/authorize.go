package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/entitlement"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/str"
)

func (h *ConnectHandler) IsAuthorized(ctx context.Context, object relation.Object, permission string) error {
	if object.Namespace == "" || object.ID == "" {
		return connect.NewError(connect.CodeInvalidArgument, ErrInvalidNamesapceOrID)
	}

	currentUser, principalErr := h.GetLoggedInPrincipal(ctx)
	if principalErr != nil {
		return principalErr
	}
	result, err := h.resourceService.CheckAuthz(ctx, resource.Check{
		Object: object,
		Subject: relation.Subject{
			Namespace: currentUser.Type,
			ID:        currentUser.ID,
		},
		Permission: permission,
	})
	if err != nil {
		return handleAuthErr(err)
	}
	if result {
		return nil
	}

	// for invitation, we need to check if the user is the owner of the invitation by checking its email as well
	if object.Namespace == schema.InvitationNamespace &&
		currentUser.Type == schema.UserPrincipal {
		result2, checkErr := h.resourceService.CheckAuthz(ctx, resource.Check{
			Object: object,
			Subject: relation.Subject{
				Namespace: currentUser.Type,
				ID:        str.GenerateUserSlug(currentUser.User.Email),
			},
			Permission: permission,
		})
		if checkErr != nil {
			return handleAuthErr(checkErr)
		}
		if result2 {
			return nil
		}
	}

	return connect.NewError(connect.CodePermissionDenied, errors.New("permission denied"))
}

func handleAuthErr(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidEmail) || errors.Is(err, errors.ErrUnauthenticated):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, organization.ErrNotExist),
		errors.Is(err, project.ErrNotExist),
		errors.Is(err, resource.ErrNotExist):
		return connect.NewError(connect.CodeNotFound, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func (h *ConnectHandler) IsSuperUser(ctx context.Context) error {
	currentUser, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return err
	}

	if currentUser.Type == schema.UserPrincipal {
		if ok, err := h.userService.IsSudo(ctx, currentUser.ID, schema.PlatformSudoPermission); err != nil {
			return connect.NewError(connect.CodeInternal, err)
		} else if ok {
			return nil
		}
	} else {
		if ok, err := h.serviceUserService.IsSudo(ctx, currentUser.ID, schema.PlatformSudoPermission); err != nil {
			return connect.NewError(connect.CodeInternal, err)
		} else if ok {
			return nil
		}
	}
	return connect.NewError(connect.CodePermissionDenied, ErrUnauthorized)
}

// CheckPlanEntitlement is only currently used to restrict seat based plans
func (h *ConnectHandler) CheckPlanEntitlement(ctx context.Context, obj relation.Object) error {
	// only check for project or org
	var orgID string
	switch obj.Namespace {
	case schema.ProjectNamespace:
		proj, err := h.projectService.Get(ctx, obj.ID)
		if err != nil {
			return err
		}
		orgID = proj.Organization.ID
	case schema.OrganizationNamespace:
		org, err := h.orgService.Get(ctx, obj.ID)
		if err != nil {
			return err
		}
		orgID = org.ID
	}
	if orgID != "" {
		customers, err := h.customerService.List(ctx, customer.Filter{
			OrgID: orgID,
		})
		if err != nil {
			return err
		}
		for _, customr := range customers {
			audit.GetAuditor(ctx, orgID).LogWithAttrs(audit.BillingEntitlementCheckedEvent, audit.Target{
				ID:   customr.ID,
				Type: "billing_account",
			}, map[string]string{})
			if err := h.entitlementService.CheckPlanEligibility(ctx, customr.ID); err != nil {
				return fmt.Errorf("%s: %w", entitlement.ErrPlanEntitlementFailed, err)
			}
		}
	}

	// default condition is true for now to avoid false positives
	return nil
}

func (h *ConnectHandler) GetRawCheckout(ctx context.Context, id string) (checkout.Checkout, error) {
	return h.checkoutService.GetByID(ctx, id)
}
