package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/entitlement"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type EntitlementService interface {
	Check(ctx context.Context, customerID, featureID string) (bool, error)
	CheckPlanEligibility(ctx context.Context, customerID string) error
}

func (h Handler) CheckFeatureEntitlement(ctx context.Context, request *frontierv1beta1.CheckFeatureEntitlementRequest) (*frontierv1beta1.CheckFeatureEntitlementResponse, error) {
	checkStatus, err := h.entitlementService.Check(ctx, request.GetBillingId(), request.GetFeature())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.CheckFeatureEntitlementResponse{
		Status: checkStatus,
	}, nil
}

// CheckPlanEntitlement is only currently used to restrict seat based plans
func (h Handler) CheckPlanEntitlement(ctx context.Context, obj relation.Object) error {
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
