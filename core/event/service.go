package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v79"

	"github.com/google/uuid"

	"github.com/raystack/frontier/billing/credit"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/core/user"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/core/organization"
)

type CheckoutService interface {
	Apply(ctx context.Context, ch checkout.Checkout) (*subscription.Subscription, *product.Product, error)
	TriggerSyncByProviderID(ctx context.Context, id string) error
}

type CustomerService interface {
	Create(ctx context.Context, customer customer.Customer, offline bool) (customer.Customer, error)
	TriggerSyncByProviderID(ctx context.Context, id string) error
	List(ctx context.Context, flt customer.Filter) ([]customer.Customer, error)
}

type SubscriptionService interface {
	TriggerSyncByProviderID(ctx context.Context, id string) error
}

type OrganizationService interface {
	GetRaw(ctx context.Context, id string) (organization.Organization, error)
}

type PlanService interface {
	GetByID(ctx context.Context, id string) (plan.Plan, error)
}

type UserService interface {
	ListByOrg(ctx context.Context, orgID string, roleFilter string) ([]user.User, error)
}

type CreditService interface {
	Add(ctx context.Context, ch credit.Credit) error
}

type Service struct {
	billingConf     billing.Config
	checkoutService CheckoutService
	customerService CustomerService
	orgService      OrganizationService
	planService     PlanService
	userService     UserService
	subsService     SubscriptionService
	creditService   CreditService

	sf singleflight.Group
}

func NewService(billingConf billing.Config, organizationService OrganizationService,
	checkoutService CheckoutService, customerService CustomerService,
	planService PlanService, userService UserService,
	subsService SubscriptionService, creditService CreditService) *Service {
	return &Service{
		billingConf:     billingConf,
		orgService:      organizationService,
		checkoutService: checkoutService,
		customerService: customerService,
		planService:     planService,
		userService:     userService,
		subsService:     subsService,
		creditService:   creditService,

		sf: singleflight.Group{},
	}
}

// EnsureDefaultPlan create a new customer account and subscribe to the default plan if configured
func (p *Service) EnsureDefaultPlan(ctx context.Context, orgID string) error {
	if p.billingConf.DefaultCurrency != "" && p.billingConf.AccountConfig.AutoCreateWithOrg {
		// only create if there is no customer account already
		customers, err := p.customerService.List(ctx, customer.Filter{
			OrgID: orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to list customers: %w", err)
		}
		if len(customers) > 0 {
			return nil
		}

		org, err := p.orgService.GetRaw(ctx, orgID)
		if err != nil {
			return fmt.Errorf("failed to get organization: %w", err)
		}
		users, err := p.userService.ListByOrg(ctx, org.ID, organization.AdminRole)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		emailID := ""
		if len(users) > 0 {
			emailID = users[0].Email
		}
		customr, err := p.customerService.Create(ctx, customer.Customer{
			OrgID:    org.ID,
			Name:     getCustomerName(org),
			Email:    emailID,
			Currency: p.billingConf.DefaultCurrency,
			Metadata: map[string]any{
				"auto_created": "true",
			},
		}, p.billingConf.AccountConfig.DefaultOffline)
		if err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}

		if p.billingConf.AccountConfig.DefaultPlan != "" {
			// validate the plan requested is free
			defaultPlan, err := p.planService.GetByID(ctx, p.billingConf.AccountConfig.DefaultPlan)
			if err != nil {
				return fmt.Errorf("failed to get default plan: %w", err)
			}

			for _, prod := range defaultPlan.Products {
				for _, price := range prod.Prices {
					if price.Amount > 0 {
						return fmt.Errorf("default plan is not free")
					}
				}
			}
			_, _, err = p.checkoutService.Apply(ctx, checkout.Checkout{
				CustomerID: customr.ID,
				PlanID:     defaultPlan.ID,
				SkipTrial:  true,
			})
			if err != nil {
				return fmt.Errorf("failed to apply default plan: %w", err)
			}
		}

		if amount := p.billingConf.AccountConfig.OnboardCreditsWithOrg; amount > 0 {
			txID := uuid.NewSHA1(credit.TxNamespaceUUID, []byte(fmt.Sprintf("%s", customr.OrgID))).String()
			if err := p.creditService.Add(ctx, credit.Credit{
				ID:          txID,
				CustomerID:  customr.ID,
				Amount:      p.billingConf.AccountConfig.OnboardCreditsWithOrg,
				Metadata:    map[string]any{"auto_created": "true"},
				Source:      credit.SourceSystemAwardedEvent,
				Description: fmt.Sprintf("Awarded %d credits for onboarding", amount),
			}); err != nil {
				// credit is only awarded once for an org
				if !errors.Is(err, credit.ErrAlreadyApplied) {
					return err
				}
			}
		}
	}
	return nil
}

func getCustomerName(org organization.Organization) string {
	if org.Title != "" {
		return org.Title
	}
	return org.Name
}

func (p *Service) BillingWebhook(ctx context.Context, payload ProviderWebhookEvent) error {
	stdLogger := grpczap.Extract(ctx).With(zap.String("provider", payload.Name))
	if payload.Name != "stripe" {
		return fmt.Errorf("provider not supported")
	}
	if len(p.billingConf.StripeWebhookSecrets) == 0 {
		return fmt.Errorf("no stripe webhook secrets configured")
	}

	webhookSignature, ok := customer.GetStripeWebhookSignatureFromContext(ctx)
	if !ok {
		return fmt.Errorf("missing billing provider webhook signature")
	}

	// try all secrets to parse the event, useful for rotating secrets
	var parseErrs []error
	var evt stripe.Event
	for _, secret := range p.billingConf.StripeWebhookSecrets {
		var err error
		evt, err = webhook.ConstructEvent(payload.Body, webhookSignature, secret)
		if err != nil {
			parseErrs = append(parseErrs, err)
			continue
		}
		break
	}
	if len(parseErrs) > 0 {
		return fmt.Errorf("failed to construct event: %w", errors.Join(parseErrs...))
	}
	ctx = context.WithoutCancel(ctx)

	// limit all executions to 1 per second per event type
	currentExecutionUnit := time.Now().Second()
	providerID := evt.GetObjectValue("id")

	go func() {
		// don't block the webhook and process it in the background
		switch evt.Type {
		case "checkout.session.completed", "checkout.session.async_payment_succeeded":
			// trigger checkout sync
			deDupKey := fmt.Sprintf("checkout-%s-%d", providerID, currentExecutionUnit)
			_, err, _ := p.sf.Do(deDupKey, func() (interface{}, error) {
				return nil, p.checkoutService.TriggerSyncByProviderID(ctx, providerID)
			})
			if err != nil {
				stdLogger.Error("error syncing checkout", zap.Error(err), zap.String("provider_id", providerID))
			}
		case "customer.created", "customer.updated", "customer.source.created", "customer.source.updated":
			// trigger customer sync
			deDupKey := fmt.Sprintf("customer-%s-%d", providerID, currentExecutionUnit)
			_, err, _ := p.sf.Do(deDupKey, func() (interface{}, error) {
				return nil, p.customerService.TriggerSyncByProviderID(ctx, providerID)
			})
			if err != nil {
				stdLogger.Error("error syncing customer", zap.Error(err), zap.String("provider_id", providerID))
			}
		case "customer.subscription.created",
			"customer.subscription.updated", "customer.subscription.deleted":
			// trigger subscriptions sync
			deDupKey := fmt.Sprintf("subscription-%s-%d", providerID, currentExecutionUnit)
			_, err, _ := p.sf.Do(deDupKey, func() (interface{}, error) {
				return nil, p.subsService.TriggerSyncByProviderID(ctx, providerID)
			})
			if err != nil {
				stdLogger.Error("error syncing subscription", zap.Error(err), zap.String("provider_id", providerID))
			}
		}
	}()
	return nil
}
