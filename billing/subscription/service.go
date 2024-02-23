package subscription

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/raystack/frontier/billing"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/robfig/cron/v3"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/pkg/debounce"
	"go.uber.org/zap"

	"github.com/raystack/frontier/billing/plan"

	"github.com/raystack/frontier/billing/customer"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

const (
	SyncDelay = time.Second * 60
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Subscription, error)
	Create(ctx context.Context, subs Subscription) (Subscription, error)
	UpdateByID(ctx context.Context, subs Subscription) (Subscription, error)
	List(ctx context.Context, filter Filter) ([]Subscription, error)
	GetByProviderID(ctx context.Context, id string) (Subscription, error)
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

type Service struct {
	repository      Repository
	stripeClient    *client.API
	customerService CustomerService
	planService     PlanService
	orgService      OrganizationService
	productService  ProductService

	syncLimiter *debounce.Limiter
	syncJob     *cron.Cron
	mu          sync.Mutex
	config      billing.Config
}

func NewService(stripeClient *client.API, config billing.Config, repository Repository,
	customerService CustomerService, planService PlanService,
	orgService OrganizationService, productService ProductService) *Service {
	return &Service{
		stripeClient:    stripeClient,
		repository:      repository,
		customerService: customerService,
		planService:     planService,
		orgService:      orgService,
		productService:  productService,
		syncLimiter:     debounce.New(2 * time.Second),
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

func (s *Service) Init(ctx context.Context) {
	if s.syncJob != nil {
		s.syncJob.Stop()
	}

	s.syncJob = cron.New()
	s.syncJob.AddFunc(fmt.Sprintf("@every %s", SyncDelay.String()), func() {
		s.backgroundSync(ctx)
	})
	s.syncJob.Start()
}

func (s *Service) Close() error {
	if s.syncJob != nil {
		return s.syncJob.Stop().Err()
	}
	return nil
}

func (s *Service) backgroundSync(ctx context.Context) {
	logger := grpczap.Extract(ctx)
	customers, err := s.customerService.List(ctx, customer.Filter{})
	if err != nil {
		logger.Error("subscription.backgroundSync", zap.Error(err))
		return
	}

	for _, customer := range customers {
		if customer.DeletedAt != nil || customer.ProviderID == "" {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer); err != nil {
			logger.Error("subscription.SyncWithProvider", zap.Error(err))
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
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

	for _, sub := range subs {
		stripeSubscription, err := s.stripeClient.Subscriptions.Get(sub.ProviderID, &stripe.SubscriptionParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Expand: []*string{
				stripe.String("schedule"),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to get subscription from billing provider: %w", err)
		}

		updateNeeded := false
		if sub.State != string(stripeSubscription.Status) {
			updateNeeded = true
			sub.State = string(stripeSubscription.Status)
		}
		if stripeSubscription.CanceledAt > 0 && sub.CanceledAt.Unix() != stripeSubscription.CanceledAt {
			updateNeeded = true
			sub.CanceledAt = time.Unix(stripeSubscription.CanceledAt, 0)
		}
		if stripeSubscription.EndedAt > 0 && sub.EndedAt.Unix() != stripeSubscription.EndedAt {
			updateNeeded = true
			sub.EndedAt = time.Unix(stripeSubscription.EndedAt, 0)
		}
		if stripeSubscription.TrialEnd > 0 && sub.TrialEndsAt.Unix() != stripeSubscription.TrialEnd {
			updateNeeded = true
			sub.TrialEndsAt = time.Unix(stripeSubscription.TrialEnd, 0)
		}
		if stripeSubscription.CurrentPeriodStart > 0 && sub.CurrentPeriodStartAt.Unix() != stripeSubscription.CurrentPeriodStart {
			updateNeeded = true
			sub.CurrentPeriodStartAt = time.Unix(stripeSubscription.CurrentPeriodStart, 0)
		}
		if stripeSubscription.CurrentPeriodEnd > 0 && sub.CurrentPeriodEndAt.Unix() != stripeSubscription.CurrentPeriodEnd {
			updateNeeded = true
			sub.CurrentPeriodEndAt = time.Unix(stripeSubscription.CurrentPeriodEnd, 0)
		}
		if stripeSubscription.BillingCycleAnchor > 0 && sub.BillingCycleAnchorAt.Unix() != stripeSubscription.BillingCycleAnchor {
			updateNeeded = true
			sub.BillingCycleAnchorAt = time.Unix(stripeSubscription.BillingCycleAnchor, 0)
		}

		// update plan id if it's changed
		planByStripeSubscription, err := s.findPlanByStripeSubscription(ctx, stripeSubscription)
		if err != nil {
			return err
		}
		if sub.PlanID != planByStripeSubscription.ID {
			sub.PlanID = planByStripeSubscription.ID
			updateNeeded = true
		}

		// update sub change if it's changed
		if stripeSubscription.Schedule != nil &&
			stripeSubscription.Schedule.CurrentPhase != nil {
			if sub.Phase.EffectiveAt.IsZero() {
				sub.Phase.EffectiveAt = time.Unix(stripeSubscription.Schedule.CurrentPhase.EndDate, 0)
				updateNeeded = true
			}
		}

		if updateNeeded {
			if _, err := s.repository.UpdateByID(ctx, sub); err != nil {
				return err
			}
		}

		// if subscription is active, and per seat pricing is enabled, update the quantity
		if sub.State == string(stripe.SubscriptionStatusActive) {
			plan, err := s.planService.GetByID(ctx, sub.PlanID)
			if err != nil {
				return err
			}

			if err = s.UpdateProductQuantity(ctx, customr.OrgID, plan, stripeSubscription); err != nil {
				return err
			}
		}
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
		// already canceled, no-op
		return sub, nil
	}

	if immediate {
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
			sub.CanceledAt = time.Unix(stripeSubscription.CanceledAt, 0)
		}
	} else {
		// check if schedule exists
		_, stripeSchedule, err := s.createOrGetSchedule(ctx, sub)
		if err != nil {
			return sub, err
		}

		// update schedule to cancel at the end of the current period
		currentPhaseItems := make([]*stripe.SubscriptionSchedulePhaseItemParams, 0, len(stripeSchedule.Phases[0].Items))
		for _, item := range stripeSchedule.Phases[0].Items {
			currentPhaseItems = append(currentPhaseItems, &stripe.SubscriptionSchedulePhaseItemParams{
				Price:    stripe.String(item.Price.ID),
				Quantity: stripe.Int64(item.Quantity),
				Metadata: item.Metadata,
			})
		}

		var currency = string(stripeSchedule.Phases[0].Currency)
		var endDate = stripe.Int64(stripeSchedule.CurrentPhase.EndDate)

		// update the phases
		updatedSchedule, err := s.stripeClient.SubscriptionSchedules.Update(stripeSchedule.ID, &stripe.SubscriptionScheduleParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Phases: []*stripe.SubscriptionSchedulePhaseParams{
				{
					Items:     currentPhaseItems,
					Currency:  stripe.String(currency),
					StartDate: stripe.Int64(stripeSchedule.CurrentPhase.StartDate),
					EndDate:   endDate,
				},
			},
			EndBehavior: stripe.String("cancel"),
		})
		if err != nil {
			return sub, fmt.Errorf("failed to cancel subscription schedule at billing provider: %w", err)
		}
		sub.Phase.EffectiveAt = time.Unix(updatedSchedule.Phases[0].EndDate, 0)
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
		Expand: []*string{stripe.String("schedule")},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subscription from billing provider: %w", err)
	}

	var stripeSchedule = stripeSubscription.Schedule
	if stripeSchedule == nil || stripeScheduleCreateRequired(stripeSchedule) {
		// no schedule exists, create a new schedule
		stripeSchedule, err = s.stripeClient.SubscriptionSchedules.New(&stripe.SubscriptionScheduleParams{
			Params: stripe.Params{
				Context: ctx,
			},
			FromSubscription: stripe.String(sub.ProviderID),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription schedule at billing provider: %w", err)
		}
	}

	return stripeSubscription, stripeSchedule, nil
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Subscription, error) {
	logger := grpczap.Extract(ctx)
	customr, err := s.customerService.GetByID(ctx, filter.CustomerID)
	if err != nil {
		return nil, err
	}
	s.syncLimiter.Call(func() {
		// fix context as the List ctx will get cancelled after call finishes
		if err := s.SyncWithProvider(context.Background(), customr); err != nil {
			logger.Error("subscription.SyncWithProvider", zap.Error(err))
		}
	})

	return s.repository.List(ctx, filter)
}

// UpdateProductQuantity updates the quantity of the product in the subscription
// Note: check if we need to handle subscription schedule
func (s *Service) UpdateProductQuantity(ctx context.Context, orgID string, plan plan.Plan, stripeSubscription *stripe.Subscription) error {
	if planFeature, ok := plan.GetUserSeatProduct(); ok {
		// get the current quantity
		count, err := s.orgService.MemberCount(ctx, orgID)
		if err != nil {
			return fmt.Errorf("failed to get member count: %w", err)
		}

		for _, subItemData := range stripeSubscription.Items.Data {
			shouldChangeQuantity := false
			switch strings.ToLower(s.config.ProductConfig.SeatChangeBehavior) {
			case "exact":
				if count != subItemData.Quantity {
					shouldChangeQuantity = true
				}
			case "incremental":
				if count > subItemData.Quantity {
					shouldChangeQuantity = true
				}
			default:
				return fmt.Errorf("invalid seat change behavior: %s", s.config.ProductConfig.SeatChangeBehavior)
			}

			// convert provider price id to system price id and get the feature
			for _, planProductPrice := range planFeature.Prices {
				if planProductPrice.ProviderID == subItemData.Price.ID {
					if shouldChangeQuantity {
						_, err = s.stripeClient.Subscriptions.Update(stripeSubscription.ID, &stripe.SubscriptionParams{
							Params: stripe.Params{
								Context: ctx,
							},
							// TODO(kushsharma): check if it removes the items we don't pass
							// in update call
							Items: []*stripe.SubscriptionItemsParams{
								{
									ID:       stripe.String(subItemData.ID),
									Quantity: stripe.Int64(count),
									Metadata: map[string]string{
										"price_id":   planProductPrice.ID,
										"managed_by": "frontier",
									},
								},
							},
							PendingInvoiceItemInterval: getPendingInvoiceItemInterval(plan),
						})
						if err != nil {
							return fmt.Errorf("failed to update subscription quantity at billing provider: %w", err)
						}
					}
				}
			}
		}
	}

	return nil
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
	if sub.State != string(stripe.SubscriptionStatusActive) {
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
		return change, nil
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

	currentPhaseItems := make([]*stripe.SubscriptionSchedulePhaseItemParams, 0, len(stripeSchedule.Phases[0].Items))
	for _, item := range stripeSchedule.Phases[0].Items {
		currentPhaseItems = append(currentPhaseItems, &stripe.SubscriptionSchedulePhaseItemParams{
			Price:    stripe.String(item.Price.ID),
			Quantity: stripe.Int64(item.Quantity),
			Metadata: item.Metadata,
		})
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

	var endDate *int64
	var endDateNow *bool
	if immediate {
		endDateNow = stripe.Bool(true)
	} else {
		endDate = stripe.Int64(stripeSchedule.CurrentPhase.EndDate)
	}
	var currency = string(stripeSchedule.Phases[0].Currency)
	var prorationBehavior = s.config.PlanChangeConfig.ProrationBehavior
	if immediate {
		prorationBehavior = s.config.PlanChangeConfig.ImmediateProrationBehavior
	}

	// update the phases
	updatedSchedule, err := s.stripeClient.SubscriptionSchedules.Update(stripeSchedule.ID, &stripe.SubscriptionScheduleParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				Items:      currentPhaseItems,
				Currency:   stripe.String(currency),
				StartDate:  stripe.Int64(stripeSchedule.CurrentPhase.StartDate),
				EndDate:    endDate,
				EndDateNow: endDateNow,
			},
			{
				Items:      nextPhaseItems,
				Currency:   stripe.String(currency),
				Iterations: stripe.Int64(1),
			},
		},
		EndBehavior:       stripe.String("release"),
		ProrationBehavior: stripe.String(prorationBehavior),
		DefaultSettings: &stripe.SubscriptionScheduleDefaultSettingsParams{
			CollectionMethod: stripe.String(s.config.PlanChangeConfig.CollectionMethod),
		},
	})
	if err != nil {
		return change, fmt.Errorf("failed to update subscription schedule at billing provider: %w", err)
	}

	sub.Phase.EffectiveAt = time.Unix(updatedSchedule.Phases[1].StartDate, 0)
	sub.Phase.PlanID = planObj.ID
	sub, err = s.repository.UpdateByID(ctx, sub)
	if err != nil {
		return change, err
	}

	return sub.Phase, nil
}

// CancelUpcomingPhase cancels the scheduled phase of the subscription
func (s *Service) CancelUpcomingPhase(ctx context.Context, sub Subscription) error {
	_, stripeSchedule, err := s.createOrGetSchedule(ctx, sub)
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
			},
		},
		EndBehavior:       stripe.String("release"),
		ProrationBehavior: stripe.String(prorationBehavior),
		DefaultSettings: &stripe.SubscriptionScheduleDefaultSettingsParams{
			CollectionMethod: stripe.String(s.config.PlanChangeConfig.CollectionMethod),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update subscription schedule at billing provider: %w", err)
	}

	sub.Phase.EffectiveAt = time.Time{}
	sub.Phase.PlanID = ""
	sub, err = s.repository.UpdateByID(ctx, sub)
	if err != nil {
		return err
	}

	return nil
}

func stripeScheduleCreateRequired(stripeSchedule *stripe.SubscriptionSchedule) bool {
	return stripeSchedule != nil &&
		(stripeSchedule.Status == stripe.SubscriptionScheduleStatusCanceled ||
			stripeSchedule.Status == stripe.SubscriptionScheduleStatusReleased)
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
