package v1beta1connect

import (
	"github.com/raystack/frontier/billing/subscription"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
