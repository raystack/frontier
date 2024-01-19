package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/entitlement"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type EntitlementService interface {
	Check(ctx context.Context, customerID, featureID string) (bool, error)
	CheckPlanEligibility(ctx context.Context, customerID string) error
}

func (h Handler) CheckFeatureEntitlement(ctx context.Context, request *frontierv1beta1.CheckFeatureEntitlementRequest) (*frontierv1beta1.CheckFeatureEntitlementResponse, error) {
	logger := grpczap.Extract(ctx)

	checkStatus, err := h.entitlementService.Check(ctx, request.GetBillingId(), request.GetFeature())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.CheckFeatureEntitlementResponse{
		Status: checkStatus,
	}, nil
}

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
		for _, customer := range customers {
			if err := h.entitlementService.CheckPlanEligibility(ctx, customer.ID); err != nil {
				audit.GetAuditor(ctx, orgID).LogWithAttrs(audit.BillingEntitlementCheckedEvent, audit.Target{
					ID:   customer.ID,
					Type: "billing_account",
				}, map[string]string{})
				return fmt.Errorf("%s: %w", entitlement.ErrPlanEntitlementFailed, err)
			}
		}
	}

	// default condition is true for now to avoid false positives
	return nil
}
