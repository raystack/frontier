package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SubscriptionService interface {
	GetByID(ctx context.Context, id string) (subscription.Subscription, error)
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
	Cancel(ctx context.Context, id string, immediate bool) (subscription.Subscription, error)
	ChangePlan(ctx context.Context, id string, change subscription.ChangeRequest) (subscription.Phase, error)
	HasUserSubscribedBefore(ctx context.Context, customerID string, planID string) (bool, error)
}

type PlanService interface {
	GetByID(ctx context.Context, id string) (plan.Plan, error)
	Create(ctx context.Context, plan plan.Plan) (plan.Plan, error)
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	UpsertPlans(ctx context.Context, planFile plan.File) error
}

func (h *ConnectHandler) ListSubscriptions(ctx context.Context, request *connect.Request[frontierv1beta1.ListSubscriptionsRequest]) (*connect.Response[frontierv1beta1.ListSubscriptionsResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichListSubscriptionsRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	if enrichedReq.GetOrgId() == "" || enrichedReq.GetBillingId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	planID := enrichedReq.GetPlan()
	if planID != "" {
		plan, err := h.planService.GetByID(ctx, planID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		planID = plan.ID
	}

	var subscriptions []*frontierv1beta1.Subscription
	subscriptionList, err := h.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: enrichedReq.GetBillingId(),
		State:      enrichedReq.GetState(),
		PlanID:     planID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range subscriptionList {
		subscriptionPB, err := transformSubscriptionToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		subscriptions = append(subscriptions, subscriptionPB)
	}

	response := &frontierv1beta1.ListSubscriptionsResponse{
		Subscriptions: subscriptions,
	}

	// Handle response enrichment based on expand field
	response = h.enrichListSubscriptionsResponse(ctx, enrichedReq, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) GetSubscription(ctx context.Context, request *connect.Request[frontierv1beta1.GetSubscriptionRequest]) (*connect.Response[frontierv1beta1.GetSubscriptionResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichGetSubscriptionRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	subscription, err := h.subscriptionService.GetByID(ctx, enrichedReq.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	subscriptionPB, err := transformSubscriptionToPB(subscription)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	response := &frontierv1beta1.GetSubscriptionResponse{
		Subscription: subscriptionPB,
	}

	// Handle response enrichment based on expand field
	response = h.enrichGetSubscriptionResponse(ctx, enrichedReq, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) CancelSubscription(ctx context.Context, request *connect.Request[frontierv1beta1.CancelSubscriptionRequest]) (*connect.Response[frontierv1beta1.CancelSubscriptionResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichCancelSubscriptionRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	_, err = h.subscriptionService.Cancel(ctx, enrichedReq.GetId(), enrichedReq.GetImmediate())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.CancelSubscriptionResponse{}), nil
}

func (h *ConnectHandler) ChangeSubscription(ctx context.Context, request *connect.Request[frontierv1beta1.ChangeSubscriptionRequest]) (*connect.Response[frontierv1beta1.ChangeSubscriptionResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichChangeSubscriptionRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	changeReq := subscription.ChangeRequest{
		PlanID:         enrichedReq.GetPlan(),
		Immediate:      enrichedReq.GetImmediate(),
		CancelUpcoming: false,
	}
	if enrichedReq.GetPlanChange() != nil {
		changeReq.PlanID = enrichedReq.GetPlanChange().GetPlan()
		changeReq.Immediate = enrichedReq.GetPlanChange().GetImmediate()
	}
	if enrichedReq.GetPhaseChange() != nil {
		changeReq.CancelUpcoming = enrichedReq.GetPhaseChange().GetCancelUpcomingChanges()
	}
	if changeReq.PlanID != "" && changeReq.CancelUpcoming {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrConflictingPlanChange)
	}
	if changeReq.PlanID == "" && !changeReq.CancelUpcoming {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNoChangeRequested)
	}

	phase, err := h.subscriptionService.ChangePlan(ctx, enrichedReq.GetId(), changeReq)
	if err != nil {
		if errors.Is(err, product.ErrPerSeatLimitReached) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrPerSeatLimitReached)
		}
		if errors.Is(err, subscription.ErrAlreadyOnSamePlan) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrAlreadyOnSamePlan)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	phasePb := &frontierv1beta1.Subscription_Phase{
		PlanId: phase.PlanID,
		Reason: phase.Reason,
	}
	if !phase.EffectiveAt.IsZero() {
		phasePb.EffectiveAt = timestamppb.New(phase.EffectiveAt)
	}
	return connect.NewResponse(&frontierv1beta1.ChangeSubscriptionResponse{
		Phase: phasePb,
	}), nil
}

func transformSubscriptionToPB(subs subscription.Subscription) (*frontierv1beta1.Subscription, error) {
	metaData, err := subs.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Subscription{}, err
	}
	var createdAt *timestamppb.Timestamp
	if !subs.CreatedAt.IsZero() {
		createdAt = timestamppb.New(subs.CreatedAt)
	}
	var canceledAt *timestamppb.Timestamp
	if !subs.CanceledAt.IsZero() {
		canceledAt = timestamppb.New(subs.CanceledAt)
	}
	var updatedAt *timestamppb.Timestamp
	if !subs.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(subs.UpdatedAt)
	}
	var endedAt *timestamppb.Timestamp
	if !subs.EndedAt.IsZero() {
		endedAt = timestamppb.New(subs.EndedAt)
	}
	var trailEndsAt *timestamppb.Timestamp
	if !subs.TrialEndsAt.IsZero() {
		trailEndsAt = timestamppb.New(subs.TrialEndsAt)
	}
	var currentPeriodStartAt *timestamppb.Timestamp
	if !subs.CurrentPeriodStartAt.IsZero() {
		currentPeriodStartAt = timestamppb.New(subs.CurrentPeriodStartAt)
	}
	var currentPeriodEndAt *timestamppb.Timestamp
	if !subs.CurrentPeriodEndAt.IsZero() {
		currentPeriodEndAt = timestamppb.New(subs.CurrentPeriodEndAt)
	}
	var billingCycleAnchorAt *timestamppb.Timestamp
	if !subs.BillingCycleAnchorAt.IsZero() {
		billingCycleAnchorAt = timestamppb.New(subs.BillingCycleAnchorAt)
	}
	var phases []*frontierv1beta1.Subscription_Phase
	if !subs.Phase.EffectiveAt.IsZero() {
		phases = append(phases, &frontierv1beta1.Subscription_Phase{
			EffectiveAt: timestamppb.New(subs.Phase.EffectiveAt),
			PlanId:      subs.Phase.PlanID,
			Reason:      subs.Phase.Reason,
		})
	}
	subsPb := &frontierv1beta1.Subscription{
		Id:                   subs.ID,
		CustomerId:           subs.CustomerID,
		PlanId:               subs.PlanID,
		ProviderId:           subs.ProviderID,
		State:                subs.State,
		Metadata:             metaData,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		CanceledAt:           canceledAt,
		EndedAt:              endedAt,
		TrialEndsAt:          trailEndsAt,
		CurrentPeriodStartAt: currentPeriodStartAt,
		CurrentPeriodEndAt:   currentPeriodEndAt,
		BillingCycleAnchorAt: billingCycleAnchorAt,
		Phases:               phases,
	}
	return subsPb, nil
}

// enrichGetSubscriptionResponse enriches the response with expanded fields
func (h *ConnectHandler) enrichGetSubscriptionResponse(ctx context.Context, req *frontierv1beta1.GetSubscriptionRequest, resp *frontierv1beta1.GetSubscriptionResponse) *frontierv1beta1.GetSubscriptionResponse {
	expandModels := parseExpandModels(req)
	if len(expandModels) == 0 {
		// no need to enrich the response
		return resp
	}

	if resp.GetSubscription() != nil {
		if expandModels["customer"] {
			ba, _ := h.GetBillingAccount(ctx, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
				Id: resp.GetSubscription().GetCustomerId(),
			}))
			if ba != nil && ba.Msg != nil {
				resp.Subscription.Customer = ba.Msg.GetBillingAccount()
			}
		}
		if expandModels["plan"] {
			plan, _ := h.GetPlan(ctx, connect.NewRequest(&frontierv1beta1.GetPlanRequest{
				Id: resp.GetSubscription().GetPlanId(),
			}))
			if plan != nil && plan.Msg != nil {
				resp.Subscription.Plan = plan.Msg.GetPlan()
			}
		}
	}

	return resp
}

// enrichListSubscriptionsResponse enriches the response with expanded fields
func (h *ConnectHandler) enrichListSubscriptionsResponse(ctx context.Context, req *frontierv1beta1.ListSubscriptionsRequest, resp *frontierv1beta1.ListSubscriptionsResponse) *frontierv1beta1.ListSubscriptionsResponse {
	expandModels := parseExpandModels(req)
	if len(expandModels) == 0 {
		// no need to enrich the response
		return resp
	}

	if len(resp.GetSubscriptions()) > 0 {
		for sIdx, s := range resp.GetSubscriptions() {
			if expandModels["customer"] {
				ba, _ := h.GetBillingAccount(ctx, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
					Id: s.GetCustomerId(),
				}))
				if ba != nil && ba.Msg != nil {
					resp.Subscriptions[sIdx].Customer = ba.Msg.GetBillingAccount()
				}
			}
			if expandModels["plan"] {
				plan, _ := h.GetPlan(ctx, connect.NewRequest(&frontierv1beta1.GetPlanRequest{
					Id: s.GetPlanId(),
				}))
				if plan != nil && plan.Msg != nil {
					resp.Subscriptions[sIdx].Plan = plan.Msg.GetPlan()
				}
			}
		}
	}

	return resp
}

// enrichGetSubscriptionRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichGetSubscriptionRequest(ctx context.Context, req *frontierv1beta1.GetSubscriptionRequest) (*frontierv1beta1.GetSubscriptionRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.GetSubscriptionRequest{
		Id:        req.GetId(),
		BillingId: req.GetBillingId(),
		OrgId:     req.GetOrgId(),
		Expand:    req.GetExpand(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichCancelSubscriptionRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichCancelSubscriptionRequest(ctx context.Context, req *frontierv1beta1.CancelSubscriptionRequest) (*frontierv1beta1.CancelSubscriptionRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.CancelSubscriptionRequest{
		Id:        req.GetId(),
		BillingId: req.GetBillingId(),
		OrgId:     req.GetOrgId(),
		Immediate: req.GetImmediate(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichListSubscriptionsRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichListSubscriptionsRequest(ctx context.Context, req *frontierv1beta1.ListSubscriptionsRequest) (*frontierv1beta1.ListSubscriptionsRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.ListSubscriptionsRequest{
		OrgId:     req.GetOrgId(),
		BillingId: req.GetBillingId(),
		Plan:      req.GetPlan(),
		State:     req.GetState(),
		Expand:    req.GetExpand(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichChangeSubscriptionRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichChangeSubscriptionRequest(ctx context.Context, req *frontierv1beta1.ChangeSubscriptionRequest) (*frontierv1beta1.ChangeSubscriptionRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.ChangeSubscriptionRequest{
		Id:          req.GetId(),
		BillingId:   req.GetBillingId(),
		OrgId:       req.GetOrgId(),
		Plan:        req.GetPlan(),
		Immediate:   req.GetImmediate(),
		Change:  req.GetChange(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}
