package subscription

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/internal/metrics"

	"github.com/google/uuid"

	"github.com/raystack/frontier/billing/credit"

	"github.com/raystack/frontier/billing"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/utils"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	"github.com/raystack/frontier/billing/plan"

	"github.com/raystack/frontier/billing/customer"
	"github.com/stripe/stripe-go/v79/client"
)

const (
	ProviderTestResource   = "test_resource"
	InitiatorIDMetadataKey = "initiated_by"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Subscription, error)
	Create(ctx context.Context, subs Subscription) (Subscription, error)
	UpdateByID(ctx context.Context, subs Subscription) (Subscription, error)
	List(ctx context.Context, filter Filter) ([]Subscription, error)
	GetByProviderID(ctx context.Context, id string) (Subscription, error)
	Delete(ctx context.Context, id string) error
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
}

type PlanService interface {
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	GetByID(ctx context.Context, id string) (plan.Plan, error)
}

type OrganizationService interface {
	MemberCount(ctx context.Context, orgID string) (int64, error)
}

type ProductService interface {
	GetByProviderID(ctx context.Context, id string) (product.Product, error)
}

type CreditService interface {
	Add(ctx context.Context, cred credit.Credit) error
	GetByID(ctx context.Context, id string) (credit.Transaction, error)
}

type Service struct {
	repository      Repository
	stripeClient    *client.API
	customerService CustomerService
	planService     PlanService
	orgService      OrganizationService
	productService  ProductService
	creditService   CreditService

	syncJob *cron.Cron
	mu      sync.Mutex
	config  billing.Config
}

func NewService(stripeClient *client.API, config billing.Config, repository Repository,
	customerService CustomerService, planService PlanService,
	orgService OrganizationService, productService ProductService,
	creditService CreditService) *Service {
	return &Service{
		stripeClient:    stripeClient,
		repository:      repository,
		customerService: customerService,
		planService:     planService,
		orgService:      orgService,
		productService:  productService,
		creditService:   creditService,
		config:          config,
	}
}

func (s *Service) Create(ctx context.Context, sub Subscription) (Subscription, error) {
	return s.repository.Create(ctx, sub)
}

func (s *Service) GetByID(ctx context.Context, id string) (Subscription, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetByProviderID(ctx context.Context, id string) (Subscription, error) {
	return s.repository.GetByProviderID(ctx, id)
}

func (s *Service) Init(ctx context.Context) error {
	syncDelay := s.config.RefreshInterval.Subscription
	if syncDelay == time.Duration(0) {
		return nil
	}
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
	}

	s.syncJob = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))
	if _, err := s.syncJob.AddFunc(fmt.Sprintf("@every %s", syncDelay.String()), func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		s.backgroundSync(ctx)
	}); err != nil {
		return err
	}
	s.syncJob.Start()
	return nil
}

func (s *Service) Close() error {
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
		return s.syncJob.Stop().Err()
	}
	return nil
}

func (s *Service) backgroundSync(ctx context.Context) {
	start := time.Now()
	if metrics.BillingSyncLatency != nil {
		record := metrics.BillingSyncLatency("subscription")
		defer record()
	}
	logger := grpczap.Extract(ctx)
	customers, err := s.customerService.List(ctx, customer.Filter{
		State: customer.ActiveState,
	})
	if err != nil {
		logger.Error("subscription.backgroundSync", zap.Error(err))
		return
	}

	for _, customer := range customers {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		if !customer.IsActive() || customer.IsOffline() {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer); err != nil {
			logger.Error("subscription.SyncWithProvider", zap.Error(err), zap.String("customer_id", customer.ID))
		}
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}
	logger.Info("subscription.backgroundSync finished", zap.Duration("duration", time.Since(start)))
}

func (s *Service) TriggerSyncByProviderID(ctx context.Context, id string) error {
	subs, err := s.repository.List(ctx, Filter{
		ProviderID: id,
	})
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return ErrNotFound
	}
	customr, err := s.customerService.GetByID(ctx, subs[0].CustomerID)
	if err != nil {
		return err
	}
	return s.SyncWithProvider(ctx, customr)
}

// SyncWithProvider syncs the subscription state with the billing provider
func (s *Service) SyncWithProvider(ctx context.Context, customr customer.Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs, err := s.repository.List(ctx, Filter{
		CustomerID: customr.ID,
	})
	if err != nil {
		return err
	}

	var subErrs []error
	for _, sub := range subs {
		if ctx.Err() != nil {
			break
		}

		if sub.IsCanceled() {
			continue
		}

		if err := s.syncSubscription(ctx, sub, customr); err != nil {
			subErrs = append(subErrs, fmt.Errorf("failed to sync subscription %s: %w", sub.ID, err))
		}
	}

	if len(subErrs) > 0 {
		return fmt.Errorf("failed to sync subscriptions: %w", errors.Join(subErrs...))
	}
	return nil
}

// syncSubscription handles syncing a single subscription with the provider
func (s *Service) syncSubscription(ctx context.Context, sub Subscription, customr customer.Customer) error {
	stripeSubscription, stripeSchedule, err := s.createOrGetSchedule(ctx, sub)
	if err != nil {
		if errors.Is(err, ErrSubscriptionOnProviderNotFound) {
			// if it's a test resource, mark it as canceled
			if val, ok := sub.Metadata[ProviderTestResource].(bool); ok && val {
				sub.State = StateCanceled.String()
				sub.CanceledAt = time.Now().UTC()
				_, err := s.repository.UpdateByID(ctx, sub)
				return err
			}
			return fmt.Errorf("%s: %w", sub.ID, err)
		}
		return err
	}

	if updated, err := s.syncSubscriptionState(ctx, sub, stripeSubscription, stripeSchedule); err != nil {
		return err
	} else if !updated.IsActive() || sub.PlanID == "" {
		return nil
	}

	// Get current plan
	subPlan, err := s.planService.GetByID(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("%w: subscription: %s plan: %s", err, sub.ID, sub.PlanID)
	}

	// Update product quantity if needed
	if err = s.UpdateProductQuantity(ctx, customr.OrgID, subPlan,
		stripeSubscription, stripeSchedule); err != nil {
		return fmt.Errorf("failed to update product quantity: %w", err)
	}

	// Ensure credits for plan
	if err := s.ensureCreditsForPlan(ctx, sub, subPlan); err != nil {
		return fmt.Errorf("ensureCreditsForPlan: %w", err)
	}

	return nil
}

// getPendingInvoiceItemInterval returns the interval for the pending invoice item based on the plan interval
// if it's yearly, it will return a monthly interval else nil
// It ensures if the user adds more members, they are charged for the new members more frequently
// than the natural subscription interval
func getPendingInvoiceItemInterval(p plan.Plan) *stripe.SubscriptionPendingInvoiceItemIntervalParams {
	if p.Interval != "year" {
		return nil
	}
	// TODO(kushsharma): make this configurable as for now every month it will
	// charge the customer for the number of users they have in the org
	// Note: the `pending_invoice_item_interval` must be more frequent than the natural
	// subscription interval.
	return &stripe.SubscriptionPendingInvoiceItemIntervalParams{
		Interval:      stripe.String("month"),
		IntervalCount: stripe.Int64(1),
	}
}

func (s *Service) Cancel(ctx context.Context, id string, immediate bool) (Subscription, error) {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return Subscription{}, err
	}
	if !sub.CanceledAt.IsZero() {
		if !immediate {
			// already canceled, no-op
			return sub, nil
		}
		// already canceled, but now we need to cancel immediately, go ahead
	}

	// check if schedule exists
	_, stripeSchedule, err := s.createOrGetSchedule(ctx, sub)
	if err != nil {
		return sub, err
	}

	if immediate || stripeSchedule == nil {
		stripeSubscription, err := s.stripeClient.Subscriptions.Cancel(sub.ProviderID, &stripe.SubscriptionCancelParams{
			Params: stripe.Params{
				Context: ctx,
			},
			InvoiceNow: stripe.Bool(true),
			Prorate:    stripe.Bool(true),
		})
		if err != nil {
			return Subscription{}, fmt.Errorf("failed to cancel subscription at billing provider: %w", err)
		}
		sub.State = string(stripeSubscription.Status)
		if stripeSubscription.CanceledAt > 0 {
			sub.CanceledAt = utils.AsTimeFromEpoch(stripeSubscription.CanceledAt)
		}
	} else {
		// TODO (Potential bug): We are ending up with Stripe subscriptions where the current phase's start date and the next phase's start date are the same.
		// One place where we saw this was in free trials. Marking it here, since this looks like one of the possible root causes where we set the current phase to be the same as next phase.

		// update schedule to cancel at the end of the current period
		currentPhase, nextPhase := s.getCurrentAndNextPhaseFromSchedule(stripeSchedule)
		if currentPhase == nil {
			// not sure if there could be a case where there is no current phase but if
			// there is, we will cancel the subscription when the next phase ends
			currentPhase = nextPhase
		}

		// update the phases
		updatedSchedule, err := s.stripeClient.SubscriptionSchedules.Update(stripeSchedule.ID, &stripe.SubscriptionScheduleParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Phases: []*stripe.SubscriptionSchedulePhaseParams{
				currentPhase,
			},
			EndBehavior: stripe.String(string(stripe.SubscriptionScheduleEndBehaviorCancel)),
		})
		if err != nil {
			return sub, fmt.Errorf("failed to cancel subscription schedule at billing provider: %w", err)
		}
		sub.Phase.PlanID = ""
		sub.Phase.Reason = SubscriptionCancel.String()
		sub.Phase.EffectiveAt = utils.AsTimeFromEpoch(updatedSchedule.Phases[0].EndDate)
	}

	return s.repository.UpdateByID(ctx, sub)
}

// createOrGetSchedule creates a new stripe schedule if it doesn't exist
func (s *Service) createOrGetSchedule(ctx context.Context, sub Subscription) (*stripe.Subscription, *stripe.SubscriptionSchedule, error) {
	// check if schedule exists
	stripeSubscription, err := s.stripeClient.Subscriptions.Get(sub.ProviderID, &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Expand: []*string{
			stripe.String("schedule"),
		},
	})
	if err != nil {
		// check if it's a subscription not found err
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil, nil, ErrSubscriptionOnProviderNotFound
		}
		return nil, nil, fmt.Errorf("failed to get subscription from billing provider: %w", err)
	}

	if stripeSubscription.Schedule != nil && stripeSubscription.Schedule.ID != "" {
		schedule, err := s.stripeClient.SubscriptionSchedules.Get(stripeSubscription.Schedule.ID, &stripe.SubscriptionScheduleParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Expand: []*string{
				stripe.String("phases.items.price.product"),
			},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get subscription schedule from billing provider: %w", err)
		}
		stripeSubscription.Schedule = schedule
	}

	if stripeSubscription.Status == stripe.SubscriptionStatusCanceled ||
		stripeSubscription.Status == stripe.SubscriptionStatusIncomplete ||
		stripeSubscription.Status == stripe.SubscriptionStatusIncompleteExpired {
		return stripeSubscription, nil, nil
	}

	if stripeSubscription.Schedule == nil {
		// no schedule exists, create a new schedule
		stripeSubscription.Schedule, err = s.stripeClient.SubscriptionSchedules.New(&stripe.SubscriptionScheduleParams{
			Params: stripe.Params{
				Context: ctx,
			},
			FromSubscription: stripe.String(sub.ProviderID),
			Expand: []*string{
				stripe.String("phases.items.price.product"),
			},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription schedule at billing provider: %w", err)
		}
	}
	return stripeSubscription, stripeSubscription.Schedule, nil
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Subscription, error) {
	return s.repository.List(ctx, filter)
}

// UpdateProductQuantity updates the quantity of the product in the subscription
// Note: check if we need to handle subscription schedule
func (s *Service) UpdateProductQuantity(ctx context.Context, orgID string, currentPlan plan.Plan,
	stripeSubscription *stripe.Subscription, stripeSchedule *stripe.SubscriptionSchedule) error {
	var orgMemberCount int64 = 1
	var err error

	// update current subscription
	currentSubscriptionItems := make([]*stripe.SubscriptionItemsParams, 0, len(stripeSubscription.Items.Data))
	for _, item := range stripeSubscription.Items.Data {
		currentSubscriptionItems = append(currentSubscriptionItems, &stripe.SubscriptionItemsParams{
			ID:       &item.ID,
			Quantity: &item.Quantity,
			Price:    &item.Price.ID,
			Metadata: item.Metadata,
		})
	}

	if planFeature, ok := currentPlan.GetUserSeatProduct(); ok {
		var shouldUpdateSubscription = false
		// get the current quantity
		orgMemberCount, err = s.orgService.MemberCount(ctx, orgID)
		if err != nil {
			return fmt.Errorf("failed to get member count: %w", err)
		}

		for _, planProductPrice := range planFeature.Prices {
			// check for changes in subscription
			for idx, subItemData := range currentSubscriptionItems {
				// convert provider price id to system price id and get the product
				if planProductPrice.ProviderID == *subItemData.Price {
					shouldChangeQuantity, err := s.shouldChangeScheduleQuantity(orgMemberCount, subItemData)
					if err != nil {
						return err
					}
					if shouldChangeQuantity {
						shouldUpdateSubscription = true
						currentSubscriptionItems[idx].Quantity = &orgMemberCount
					}
				}
			}
		}

		if shouldUpdateSubscription {
			_, err := s.stripeClient.Subscriptions.Update(stripeSubscription.ID, &stripe.SubscriptionParams{
				Params: stripe.Params{
					Context: ctx,
				},
				Items:                      currentSubscriptionItems,
				PendingInvoiceItemInterval: getPendingInvoiceItemInterval(currentPlan),
			})
			if err != nil {
				return fmt.Errorf("failed to update subscription quantity at billing provider: %w", err)
			}
		}
	}

	// if there is a next phase, we will also update all phases of schedule
	currentPhase, nextPhase := s.getCurrentAndNextPhaseFromSchedule(stripeSchedule)
	if nextPhase == nil {
		// no need to update the phases if there is no next phase
		return nil
	}

	_, nextPlanID, err := s.getPlanFromSchedule(ctx, stripeSchedule)
	if errors.Is(err, ErrNoPhaseActive) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get plan from schedule: %w", err)
	}
	nextPlan, err := s.planService.GetByID(ctx, nextPlanID)
	if err != nil {
		return fmt.Errorf("failed to get next plan: %w", err)
	}
	var shouldUpdateSchedule = false

	if planFeature, ok := currentPlan.GetUserSeatProduct(); ok {
		for _, planProductPrice := range planFeature.Prices {
			// check for changes in schedule
			for idx, subItemData := range currentPhase.Items {
				// convert provider price id to system price id and get the product
				if planProductPrice.ProviderID == *subItemData.Price {
					shouldChangeQuantity, err := s.shouldChangePhaseQuantity(orgMemberCount, subItemData)
					if err != nil {
						return err
					}

					if shouldChangeQuantity {
						shouldUpdateSchedule = true
						currentPhase.Items[idx].Quantity = &orgMemberCount
					}
				}
			}
		}
	}
	if planFeature, ok := nextPlan.GetUserSeatProduct(); ok {
		for _, planProductPrice := range planFeature.Prices {
			// check for changes in schedule
			for idx, subItemData := range nextPhase.Items {
				// convert provider price id to system price id and get the product
				if planProductPrice.ProviderID == *subItemData.Price {
					shouldChangeQuantity, err := s.shouldChangePhaseQuantity(orgMemberCount, subItemData)
					if err != nil {
						return err
					}

					if shouldChangeQuantity {
						shouldUpdateSchedule = true
						nextPhase.Items[idx].Quantity = &orgMemberCount
					}
				}
			}
		}
	}

	if shouldUpdateSchedule {
		updatedPhases := make([]*stripe.SubscriptionSchedulePhaseParams, 0, len(stripeSchedule.Phases))
		if *currentPhase.EndDate > time.Now().Unix() {
			updatedPhases = append(updatedPhases, currentPhase)
		}

		updatedPhases = append(updatedPhases, nextPhase)
		_, err = s.stripeClient.SubscriptionSchedules.Update(stripeSchedule.ID, &stripe.SubscriptionScheduleParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Phases: updatedPhases,
		})
		if err != nil {
			return fmt.Errorf("failed to update subscription schedule at billing provider: %w", err)
		}
	}

	return nil
}

func (s *Service) shouldChangeScheduleQuantity(orgMemberCount int64, subItemData *stripe.SubscriptionItemsParams) (bool, error) {
	shouldChangeQuantity := false
	switch strings.ToLower(s.config.ProductConfig.SeatChangeBehavior) {
	case "exact":
		if orgMemberCount != *subItemData.Quantity {
			shouldChangeQuantity = true
		}
	case "incremental":
		if orgMemberCount > *subItemData.Quantity {
			shouldChangeQuantity = true
		}
	default:
		return false, fmt.Errorf("invalid seat change behavior: %s", s.config.ProductConfig.SeatChangeBehavior)
	}
	return shouldChangeQuantity, nil
}

func (s *Service) shouldChangePhaseQuantity(orgMemberCount int64, subItemData *stripe.SubscriptionSchedulePhaseItemParams) (bool, error) {
	shouldChangeQuantity := false
	switch strings.ToLower(s.config.ProductConfig.SeatChangeBehavior) {
	case "exact":
		if orgMemberCount != *subItemData.Quantity {
			shouldChangeQuantity = true
		}
	case "incremental":
		if orgMemberCount > *subItemData.Quantity {
			shouldChangeQuantity = true
		}
	default:
		return false, fmt.Errorf("invalid seat change behavior: %s", s.config.ProductConfig.SeatChangeBehavior)
	}
	return shouldChangeQuantity, nil
}

// ChangePlan changes the plan of the subscription by creating a subscription schedule
// it first checks if the schedule is already created, if not it creates a new schedule
// using the current subscription as the base and the new plan as the target in upcoming phase.
// Phases can be immediately changed or at the end of the current period.
func (s *Service) ChangePlan(ctx context.Context, id string, changeRequest ChangeRequest) (Phase, error) {
	var change Phase

	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return change, err
	}
	if !sub.IsActive() {
		return change, fmt.Errorf("only active subscriptions can be changed")
	}
	if changeRequest.CancelUpcoming {
		return Phase{}, s.CancelUpcomingPhase(ctx, sub)
	}

	planID := changeRequest.PlanID
	immediate := changeRequest.Immediate

	planObj, err := s.planService.GetByID(ctx, planID)
	if err != nil {
		return change, err
	}

	// check if the plan is already changed
	if sub.PlanID == planObj.ID {
		return change, ErrAlreadyOnSamePlan
	}

	// check if schedule exists
	stripeSubscription, stripeSchedule, err := s.createOrGetSchedule(ctx, sub)
	if err != nil {
		return change, err
	}

	// schedule is active, update the phases
	planByStripeSubscription, err := s.findPlanByStripeSubscription(ctx, stripeSubscription)
	if err != nil {
		return change, err
	}

	// check if the plan is already changed
	if planByStripeSubscription.ID == planObj.ID {
		return change, nil
	}

	customerObj, err := s.customerService.GetByID(ctx, sub.CustomerID)
	if err != nil {
		return change, err
	}
	userCount, err := s.orgService.MemberCount(ctx, customerObj.OrgID)
	if err != nil {
		return change, fmt.Errorf("failed to get member count: %w", err)
	}

	var nextPhaseItems []*stripe.SubscriptionSchedulePhaseItemParams
	for _, planProduct := range planObj.Products {
		// if it's credit, skip
		if planProduct.Behavior == product.CreditBehavior {
			continue
		}

		// if per seat, check if there is a limit of seats, if it breaches limit, fail
		if planProduct.Behavior == product.PerSeatBehavior {
			if planProduct.Config.SeatLimit > 0 && userCount > planProduct.Config.SeatLimit {
				return change, fmt.Errorf("member count exceeds allowed limit of the plan: %w", product.ErrPerSeatLimitReached)
			}
		}
		for _, planProductPrice := range planProduct.Prices {
			// only work with plan interval prices
			if planProductPrice.Interval != planObj.Interval {
				continue
			}

			var quantity int64 = 1
			if planProduct.Behavior == product.PerSeatBehavior {
				quantity = userCount
			}
			nextPhaseItems = append(nextPhaseItems, &stripe.SubscriptionSchedulePhaseItemParams{
				Price:    stripe.String(planProductPrice.ProviderID),
				Quantity: stripe.Int64(quantity),
				Metadata: map[string]string{
					"price_id":   planProductPrice.ID,
					"managed_by": "frontier",
				},
			})
		}
	}

	// find current phase out of list of phases
	currentPhaseItems, err := s.getCurrentPhaseItemsFromSchedule(stripeSchedule)
	if err != nil && !errors.Is(err, ErrPhaseIsUpdating) {
		return change, err
	}

	var endDate *int64
	var endDateNow *bool
	if immediate {
		endDateNow = stripe.Bool(true)
	} else {
		endDate = stripe.Int64(stripeSchedule.CurrentPhase.EndDate)
	}
	var prorationBehavior = s.config.PlanChangeConfig.ProrationBehavior
	if immediate {
		prorationBehavior = s.config.PlanChangeConfig.ImmediateProrationBehavior
	}
	currentAutoTaxStatus := false
	if stripeSubscription.AutomaticTax != nil {
		currentAutoTaxStatus = stripeSubscription.AutomaticTax.Enabled
	}

	var updatePhases []*stripe.SubscriptionSchedulePhaseParams
	if currentPhaseItems != nil {
		updatePhases = append(updatePhases, &stripe.SubscriptionSchedulePhaseParams{
			Items:      currentPhaseItems,
			Currency:   stripe.String(customerObj.Currency),
			StartDate:  stripe.Int64(stripeSchedule.CurrentPhase.StartDate),
			EndDate:    endDate,
			EndDateNow: endDateNow,
			Metadata: map[string]string{
				"plan_id":    planByStripeSubscription.ID,
				"managed_by": "frontier",
			},
			AutomaticTax: &stripe.SubscriptionSchedulePhaseAutomaticTaxParams{
				Enabled: stripe.Bool(currentAutoTaxStatus),
			},
		})
	}
	if len(nextPhaseItems) > 0 {
		updatePhases = append(updatePhases, &stripe.SubscriptionSchedulePhaseParams{
			Items:      nextPhaseItems,
			Currency:   stripe.String(customerObj.Currency),
			Iterations: stripe.Int64(1),
			Metadata: map[string]string{
				"plan_id":    planObj.ID,
				"managed_by": "frontier",
			},

			// when changing plan, we will set up autotax based on config
			AutomaticTax: &stripe.SubscriptionSchedulePhaseAutomaticTaxParams{
				Enabled: stripe.Bool(s.config.StripeAutoTax),
			},
		})
	}

	// update the phases
	updatedSchedule, err := s.stripeClient.SubscriptionSchedules.Update(stripeSchedule.ID, &stripe.SubscriptionScheduleParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Phases:            updatePhases,
		EndBehavior:       stripe.String("release"),
		ProrationBehavior: stripe.String(prorationBehavior),
		DefaultSettings: &stripe.SubscriptionScheduleDefaultSettingsParams{
			CollectionMethod: stripe.String(s.config.PlanChangeConfig.CollectionMethod),
		},
	})
	if err != nil {
		return change, fmt.Errorf("failed to update subscription schedule at billing provider: %w", err)
	}

	// update subscription with new phase
	currentPlanID, nextPlanID, err := s.getPlanFromSchedule(ctx, updatedSchedule)
	if err != nil {
		return change, err
	}
	if updatedSchedule.CurrentPhase.EndDate > 0 {
		sub.Phase.EffectiveAt = utils.AsTimeFromEpoch(updatedSchedule.CurrentPhase.EndDate)
	}
	sub.Phase.Reason = SubscriptionChange.String()
	sub.Phase.PlanID = nextPlanID
	if nextPlanID == "" {
		// if there is no next plan, it means the change was instant
		sub.Phase.PlanID = currentPlanID
		sub.Phase.EffectiveAt = utils.AsTimeFromEpoch(updatedSchedule.CurrentPhase.StartDate)
	}

	sub, err = s.repository.UpdateByID(ctx, sub)
	if err != nil {
		return change, err
	}

	return sub.Phase, nil
}

func (s *Service) getCurrentPhaseItemsFromSchedule(stripeSchedule *stripe.SubscriptionSchedule) ([]*stripe.SubscriptionSchedulePhaseItemParams, error) {
	if stripeSchedule == nil || stripeSchedule.CurrentPhase == nil || len(stripeSchedule.Phases) == 0 {
		return nil, ErrNoPhaseActive
	}
	if stripeSchedule.CurrentPhase.EndDate < time.Now().Unix() {
		// current phase has ended
		return nil, ErrPhaseIsUpdating
	}
	var currentPhaseItems []*stripe.SubscriptionSchedulePhaseItemParams
	for _, phase := range stripeSchedule.Phases {
		if phase.StartDate == stripeSchedule.CurrentPhase.StartDate &&
			phase.EndDate == stripeSchedule.CurrentPhase.EndDate {
			currentPhaseItems = make([]*stripe.SubscriptionSchedulePhaseItemParams, 0, len(phase.Items))
			for _, item := range phase.Items {
				currentPhaseItems = append(currentPhaseItems, &stripe.SubscriptionSchedulePhaseItemParams{
					Price:    stripe.String(item.Price.ID),
					Quantity: stripe.Int64(item.Quantity),
					Metadata: item.Metadata,
				})
			}
			break
		}
	}
	return currentPhaseItems, nil
}

func (s *Service) getCurrentAndNextPhaseFromSchedule(stripeSchedule *stripe.SubscriptionSchedule) (*stripe.SubscriptionSchedulePhaseParams, *stripe.SubscriptionSchedulePhaseParams) {
	if stripeSchedule == nil || stripeSchedule.CurrentPhase == nil {
		return nil, nil
	}
	var currentPhase *stripe.SubscriptionSchedulePhaseParams
	var nextPhase *stripe.SubscriptionSchedulePhaseParams

	for _, phase := range stripeSchedule.Phases {
		if phase.StartDate == stripeSchedule.CurrentPhase.StartDate &&
			phase.EndDate == stripeSchedule.CurrentPhase.EndDate {
			currentPhase = createSchedulePhase(phase)
		} else if phase.StartDate >= stripeSchedule.CurrentPhase.EndDate {
			nextPhase = createSchedulePhase(phase)
		}
	}

	return currentPhase, nextPhase
}

func createSchedulePhase(phase *stripe.SubscriptionSchedulePhase) *stripe.SubscriptionSchedulePhaseParams {
	newPhaseItems := make([]*stripe.SubscriptionSchedulePhaseItemParams, 0, len(phase.Items))
	for _, item := range phase.Items {
		newPhaseItems = append(newPhaseItems, &stripe.SubscriptionSchedulePhaseItemParams{
			Price:    stripe.String(item.Price.ID),
			Quantity: stripe.Int64(item.Quantity),
			Metadata: item.Metadata,
		})
	}

	phaseAutoTaxStatus := false
	if phase.AutomaticTax != nil {
		phaseAutoTaxStatus = phase.AutomaticTax.Enabled
	}
	newPhase := &stripe.SubscriptionSchedulePhaseParams{
		Items:     newPhaseItems,
		Currency:  stripe.String(string(phase.Currency)),
		StartDate: stripe.Int64(phase.StartDate),
		EndDate:   stripe.Int64(phase.EndDate),
		Metadata:  phase.Metadata,
		AutomaticTax: &stripe.SubscriptionSchedulePhaseAutomaticTaxParams{
			Enabled: stripe.Bool(phaseAutoTaxStatus),
		},
		Description: stripe.String(phase.Description),
	}
	if phase.TrialEnd > 0 {
		newPhase.TrialEnd = stripe.Int64(phase.TrialEnd)
	}
	if phase.ProrationBehavior != "" {
		newPhase.ProrationBehavior = stripe.String(string(phase.ProrationBehavior))
	}
	if phase.CollectionMethod != nil {
		newPhase.CollectionMethod = stripe.String(string(*phase.CollectionMethod))
	}
	return newPhase
}

// todo(kushsharma): return plan instead of id
func (s *Service) getPlanFromSchedule(ctx context.Context, stripeSchedule *stripe.SubscriptionSchedule) (string, string, error) {
	if stripeSchedule == nil || stripeSchedule.CurrentPhase == nil {
		return "", "", ErrNoPhaseActive
	}
	var currentPlanID string
	var nextPlanID string
	for _, phase := range stripeSchedule.Phases {
		if phase.StartDate == stripeSchedule.CurrentPhase.StartDate {
			if phase.Metadata != nil {
				if planID, ok := phase.Metadata["plan_id"]; ok {
					currentPlanID = planID
					continue
				}
			}
			currentPlan, err := s.findPlanByStripePhase(ctx, phase)
			if err != nil {
				return "", "", err
			}
			currentPlanID = currentPlan.ID
		} else if phase.StartDate >= stripeSchedule.CurrentPhase.EndDate {
			if phase.Metadata != nil {
				if planID, ok := phase.Metadata["plan_id"]; ok {
					nextPlanID = planID
					continue
				}
			}

			nextPlan, err := s.findPlanByStripePhase(ctx, phase)
			if err != nil {
				return "", "", err
			}
			nextPlanID = nextPlan.ID
		}
	}
	return currentPlanID, nextPlanID, nil
}

// CancelUpcomingPhase cancels the scheduled phase of the subscription
func (s *Service) CancelUpcomingPhase(ctx context.Context, sub Subscription) error {
	stripeSub, stripeSchedule, err := s.createOrGetSchedule(ctx, sub)
	if err != nil {
		return err
	}

	currentPhaseItems := make([]*stripe.SubscriptionSchedulePhaseItemParams, 0, len(stripeSchedule.Phases[0].Items))
	for _, item := range stripeSchedule.Phases[0].Items {
		currentPhaseItems = append(currentPhaseItems, &stripe.SubscriptionSchedulePhaseItemParams{
			Price:    stripe.String(item.Price.ID),
			Quantity: stripe.Int64(item.Quantity),
			Metadata: item.Metadata,
		})
	}
	var currency = string(stripeSchedule.Phases[0].Currency)
	var prorationBehavior = s.config.PlanChangeConfig.ProrationBehavior

	var endBehavior = stripe.SubscriptionScheduleEndBehaviorRelease

	if stripeSub.Status == stripe.SubscriptionStatusTrialing && s.config.SubscriptionConfig.BehaviorAfterTrial == "cancel" {
		endBehavior = stripe.SubscriptionScheduleEndBehaviorCancel
	}

	// update the phases
	_, err = s.stripeClient.SubscriptionSchedules.Update(stripeSchedule.ID, &stripe.SubscriptionScheduleParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items:     currentPhaseItems,
				Currency:  stripe.String(currency),
				StartDate: stripe.Int64(stripeSchedule.CurrentPhase.StartDate),
				EndDate:   stripe.Int64(stripeSchedule.CurrentPhase.EndDate),
				Metadata: map[string]string{
					"plan_id":    sub.PlanID,
					"managed_by": "frontier",
				},
			},
		},
		EndBehavior:       stripe.String(string(endBehavior)),
		ProrationBehavior: stripe.String(prorationBehavior),
		DefaultSettings: &stripe.SubscriptionScheduleDefaultSettingsParams{
			CollectionMethod: stripe.String(s.config.PlanChangeConfig.CollectionMethod),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update subscription schedule at billing provider: %w", err)
	}

	sub.Phase.Reason = ""
	sub.Phase.EffectiveAt = time.Time{}
	sub.Phase.PlanID = ""
	_, err = s.repository.UpdateByID(ctx, sub)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) findPlanByStripeSubscription(ctx context.Context, stripeSubscription *stripe.Subscription) (plan.Plan, error) {
	// keep plan id in sync based on what products are attached to the subscription
	// it can change if the user changes the plan using a schedule
	var productPlanIDs []string
	var interval string

	for _, subStripeItem := range stripeSubscription.Items.Data {
		product, err := s.productService.GetByProviderID(ctx, subStripeItem.Price.Product.ID)
		if err != nil {
			return plan.Plan{}, fmt.Errorf("failed to get product from billing provider: %w", err)
		}
		if len(productPlanIDs) == 0 {
			productPlanIDs = append(productPlanIDs, product.PlanIDs...)
			interval = string(subStripeItem.Price.Recurring.Interval)
			continue
		}
		productPlanIDs = utils.Intersection(productPlanIDs, product.PlanIDs)
	}

	plans, err := s.planService.List(ctx, plan.Filter{
		IDs:      productPlanIDs,
		Interval: interval,
	})
	if err != nil {
		return plan.Plan{}, err
	}

	if len(plans) == 0 {
		return plan.Plan{}, fmt.Errorf("no plan found for subscription provider id: %s", stripeSubscription.ID)
	} else if len(plans) > 1 {
		return plan.Plan{}, fmt.Errorf("multiple plans found for products: %v", plans)
	}

	return plans[0], nil
}

func (s *Service) findPlanByStripePhase(ctx context.Context, stripePhase *stripe.SubscriptionSchedulePhase) (plan.Plan, error) {
	// keep plan id in sync based on what products are attached to the subscription
	// it can change if the user changes the plan using a schedule
	var productPlanIDs []string
	var interval string

	for _, subStripeItem := range stripePhase.Items {
		product, err := s.productService.GetByProviderID(ctx, subStripeItem.Price.Product.ID)
		if err != nil {
			return plan.Plan{}, fmt.Errorf("failed to get product from billing provider: %w", err)
		}
		if len(productPlanIDs) == 0 {
			productPlanIDs = append(productPlanIDs, product.PlanIDs...)
			interval = string(subStripeItem.Price.Recurring.Interval)
			continue
		}
		productPlanIDs = utils.Intersection(productPlanIDs, product.PlanIDs)
	}

	plans, err := s.planService.List(ctx, plan.Filter{
		IDs:      productPlanIDs,
		Interval: interval,
	})
	if err != nil {
		return plan.Plan{}, err
	}

	if len(plans) == 0 {
		return plan.Plan{}, fmt.Errorf("no plan found for phase products: %v, interval: %s", productPlanIDs, interval)
	} else if len(plans) > 1 {
		return plan.Plan{}, fmt.Errorf("multiple plans found for products: %v", plans)
	}

	return plans[0], nil
}

func (s *Service) ensureCreditsForPlan(ctx context.Context, sub Subscription, subPlan plan.Plan) error {
	customerID := sub.CustomerID
	txID := uuid.NewSHA1(credit.TxNamespaceUUID, []byte(fmt.Sprintf("%s:%s", subPlan.ID, customerID))).String()
	if subPlan.OnStartCredits == 0 {
		// no such product
		return nil
	}

	// if already subscribed to the plan before, don't provide starter credits
	// a plan's on start credits gets awarded only once, we should make it configurable
	tx, err := s.creditService.GetByID(ctx, txID)
	if err == nil && tx.CustomerID == customerID {
		return nil
	}

	initiatorID := ""
	if id, ok := sub.Metadata[InitiatorIDMetadataKey].(string); ok {
		initiatorID = id
	}

	description := fmt.Sprintf("addition of %d credits for %s", subPlan.OnStartCredits, subPlan.Title)
	if err := s.creditService.Add(ctx, credit.Credit{
		ID:          txID,
		CustomerID:  customerID,
		Amount:      subPlan.OnStartCredits,
		Source:      credit.SourceSystemOnboardEvent,
		Metadata:    subPlan.Metadata,
		Description: description,
		UserID:      initiatorID,
	}); err != nil && !errors.Is(err, credit.ErrAlreadyApplied) {
		return err
	}
	return nil
}

func (s *Service) DeleteByCustomer(ctx context.Context, customr customer.Customer) error {
	subs, err := s.List(ctx, Filter{
		CustomerID: customr.ID,
	})
	if err != nil {
		return err
	}
	if err := s.SyncWithProvider(ctx, customr); err != nil {
		return err
	}
	for _, sub := range subs {
		if sub.IsActive() {
			if _, err := s.Cancel(ctx, sub.ID, true); err != nil {
				return err
			}
		}
		if err := s.repository.Delete(ctx, sub.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HasUserSubscribedBefore(ctx context.Context, customerID string, planID string) (bool, error) {
	subs, err := s.List(ctx, Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return false, err
	}
	for _, sub := range subs {
		isPlanUsedBefore := false
		if sub.PlanID == planID {
			isPlanUsedBefore = true
		}
		for _, history := range sub.PlanHistory {
			if history.PlanID == planID {
				isPlanUsedBefore = true
			}
		}

		if isPlanUsedBefore {
			return true, nil
		}
	}
	return false, nil
}

// syncSubscriptionState syncs the subscription state with the provider and returns the updated subscription
func (s *Service) syncSubscriptionState(ctx context.Context, sub Subscription,
	stripeSubscription *stripe.Subscription,
	stripeSchedule *stripe.SubscriptionSchedule) (Subscription, error) {
	updateNeeded := false

	// Sync basic subscription state
	if sub.State != string(stripeSubscription.Status) {
		updateNeeded = true
		sub.State = string(stripeSubscription.Status)
	}

	// Sync timestamps
	timestamps := []struct {
		current *time.Time
		new     int64
	}{
		{&sub.CanceledAt, stripeSubscription.CanceledAt},
		{&sub.EndedAt, stripeSubscription.EndedAt},
		{&sub.TrialEndsAt, stripeSubscription.TrialEnd},
		{&sub.CurrentPeriodStartAt, stripeSubscription.CurrentPeriodStart},
		{&sub.CurrentPeriodEndAt, stripeSubscription.CurrentPeriodEnd},
		{&sub.BillingCycleAnchorAt, stripeSubscription.BillingCycleAnchor},
	}

	for _, ts := range timestamps {
		if ts.new > 0 && ts.current.Unix() != ts.new {
			updateNeeded = true
			*ts.current = utils.AsTimeFromEpoch(ts.new)
		}
	}

	// Update plan IDs
	currentPlanID, nextPlanID, err := s.getPlanFromSchedule(ctx, stripeSchedule)
	if errors.Is(err, ErrNoPhaseActive) {
		currentPlan, err := s.findPlanByStripeSubscription(ctx, stripeSubscription)
		if err != nil {
			return sub, fmt.Errorf("failed to find plan from stripe subscription: %w", err)
		}
		currentPlanID = currentPlan.ID
	} else if err != nil {
		return sub, fmt.Errorf("failed to find plan from stripe schedule: %w", err)
	}

	if sub.PlanID != currentPlanID {
		updateNeeded = true
		if sub.PlanID != "" {
			sub.PlanHistory = append(sub.PlanHistory, Phase{
				EndsAt: time.Now().UTC(),
				PlanID: sub.PlanID,
			})
		}
		sub.PlanID = currentPlanID
	}

	// Update phase
	if sub.Phase.PlanID != nextPlanID {
		updateNeeded = true
		sub.Phase.PlanID = nextPlanID
		sub.Phase.Reason = SubscriptionChange.String()

		if stripeSchedule != nil && stripeSchedule.EndBehavior == stripe.SubscriptionScheduleEndBehaviorCancel {
			sub.Phase.Reason = SubscriptionCancel.String()
		}
	}

	// Update phase effective date
	if stripeSubscription.Schedule != nil {
		if stripeSubscription.Schedule.CurrentPhase == nil && sub.Phase.EffectiveAt.Unix() > 0 {
			updateNeeded = true
			sub.Phase.EffectiveAt = time.Time{}
		} else if stripeSubscription.Schedule.CurrentPhase != nil &&
			sub.Phase.EffectiveAt.Unix() != stripeSubscription.Schedule.CurrentPhase.EndDate {
			updateNeeded = true
			sub.Phase.EffectiveAt = utils.AsTimeFromEpoch(stripeSubscription.Schedule.CurrentPhase.EndDate)
		}
	}

	if updateNeeded {
		return s.repository.UpdateByID(ctx, sub)
	}
	return sub, nil
}
