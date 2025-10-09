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
	if request.Msg.GetOrgId() == "" || request.Msg.GetBillingId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	planID := request.Msg.GetPlan()
	if planID != "" {
		plan, err := h.planService.GetByID(ctx, planID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		planID = plan.ID
	}

	var subscriptions []*frontierv1beta1.Subscription
	subscriptionList, err := h.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: request.Msg.GetBillingId(),
		State:      request.Msg.GetState(),
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
	response = h.enrichListSubscriptionsResponse(ctx, request.Msg, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) GetSubscription(ctx context.Context, request *connect.Request[frontierv1beta1.GetSubscriptionRequest]) (*connect.Response[frontierv1beta1.GetSubscriptionResponse], error) {
	subscription, err := h.subscriptionService.GetByID(ctx, request.Msg.GetId())
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
	response = h.enrichGetSubscriptionResponse(ctx, request.Msg, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) CancelSubscription(ctx context.Context, request *connect.Request[frontierv1beta1.CancelSubscriptionRequest]) (*connect.Response[frontierv1beta1.CancelSubscriptionResponse], error) {
	_, err := h.subscriptionService.Cancel(ctx, request.Msg.GetId(), request.Msg.GetImmediate())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.CancelSubscriptionResponse{}), nil
}

func (h *ConnectHandler) ChangeSubscription(ctx context.Context, request *connect.Request[frontierv1beta1.ChangeSubscriptionRequest]) (*connect.Response[frontierv1beta1.ChangeSubscriptionResponse], error) {
	changeReq := subscription.ChangeRequest{
		PlanID:         request.Msg.GetPlan(),
		Immediate:      request.Msg.GetImmediate(),
		CancelUpcoming: false,
	}
	if request.Msg.GetPlanChange() != nil {
		changeReq.PlanID = request.Msg.GetPlanChange().GetPlan()
		changeReq.Immediate = request.Msg.GetPlanChange().GetImmediate()
	}
	if request.Msg.GetPhaseChange() != nil {
		changeReq.CancelUpcoming = request.Msg.GetPhaseChange().GetCancelUpcomingChanges()
	}
	if changeReq.PlanID != "" && changeReq.CancelUpcoming {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrConflictingPlanChange)
	}
	if changeReq.PlanID == "" && !changeReq.CancelUpcoming {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNoChangeRequested)
	}

	phase, err := h.subscriptionService.ChangePlan(ctx, request.Msg.GetId(), changeReq)
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
