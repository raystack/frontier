package v1beta1

import (
	"context"
	"errors"

	"github.com/raystack/frontier/billing/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
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

func (h Handler) ListSubscriptions(ctx context.Context, request *frontierv1beta1.ListSubscriptionsRequest) (*frontierv1beta1.ListSubscriptionsResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetOrgId() == "" || request.GetBillingId() == "" {
		return nil, grpcBadBodyError
	}
	planID := request.GetPlan()
	if planID != "" {
		plan, err := h.planService.GetByID(ctx, planID)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyError
		}
		planID = plan.ID
	}

	var subscriptions []*frontierv1beta1.Subscription
	subscriptionList, err := h.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: request.GetBillingId(),
		State:      request.GetState(),
		PlanID:     planID,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range subscriptionList {
		subscriptionPB, err := transformSubscriptionToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		subscriptions = append(subscriptions, subscriptionPB)
	}

	return &frontierv1beta1.ListSubscriptionsResponse{
		Subscriptions: subscriptions,
	}, nil
}

func (h Handler) GetSubscription(ctx context.Context, request *frontierv1beta1.GetSubscriptionRequest) (*frontierv1beta1.GetSubscriptionResponse, error) {
	logger := grpczap.Extract(ctx)

	subscription, err := h.subscriptionService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	subscriptionPB, err := transformSubscriptionToPB(subscription)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.GetSubscriptionResponse{
		Subscription: subscriptionPB,
	}, nil
}

func (h Handler) CancelSubscription(ctx context.Context, request *frontierv1beta1.CancelSubscriptionRequest) (*frontierv1beta1.CancelSubscriptionResponse, error) {
	logger := grpczap.Extract(ctx)

	_, err := h.subscriptionService.Cancel(ctx, request.GetId(), request.GetImmediate())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.CancelSubscriptionResponse{}, nil
}

func (h Handler) ChangeSubscription(ctx context.Context, request *frontierv1beta1.ChangeSubscriptionRequest) (*frontierv1beta1.ChangeSubscriptionResponse, error) {
	logger := grpczap.Extract(ctx)

	changeReq := subscription.ChangeRequest{
		PlanID:         request.GetPlan(),
		Immediate:      request.GetImmediate(),
		CancelUpcoming: false,
	}
	if request.GetPlanChange() != nil {
		changeReq.PlanID = request.GetPlanChange().GetPlan()
		changeReq.Immediate = request.GetPlanChange().GetImmediate()
	}
	if request.GetPhaseChange() != nil {
		changeReq.CancelUpcoming = request.GetPhaseChange().GetCancelUpcomingChanges()
	}
	if changeReq.PlanID != "" && changeReq.CancelUpcoming {
		return nil, status.Error(codes.InvalidArgument, "cannot change plan and cancel upcoming changes at the same time")
	}
	if changeReq.PlanID == "" && !changeReq.CancelUpcoming {
		return nil, status.Error(codes.InvalidArgument, "no change requested")
	}

	phase, err := h.subscriptionService.ChangePlan(ctx, request.GetId(), changeReq)
	if err != nil {
		logger.Error(err.Error())
		if errors.Is(err, product.ErrPerSeatLimitReached) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, subscription.ErrAlreadyOnSamePlan) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, grpcInternalServerError
	}

	phasePb := &frontierv1beta1.Subscription_Phase{
		PlanId: phase.PlanID,
	}
	if !phase.EffectiveAt.IsZero() {
		phasePb.EffectiveAt = timestamppb.New(phase.EffectiveAt)
	}
	return &frontierv1beta1.ChangeSubscriptionResponse{
		Phase: phasePb,
	}, nil
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
