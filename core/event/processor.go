package event

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/user"

	"github.com/raystack/frontier/billing/plan"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/core/organization"
)

type CheckoutService interface {
	Apply(ctx context.Context, ch checkout.Checkout) (*subscription.Subscription, *product.Product, error)
}

type CustomerService interface {
	Create(ctx context.Context, customer customer.Customer) (customer.Customer, error)
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

type Processor struct {
	billingConf     billing.Config
	checkoutService CheckoutService
	customerService CustomerService
	orgService      OrganizationService
	planService     PlanService
	userService     UserService
}

func NewProcessor(billingConf billing.Config, organizationService OrganizationService,
	checkoutService CheckoutService, customerService CustomerService,
	planService PlanService, userService UserService) *Processor {
	return &Processor{
		billingConf:     billingConf,
		orgService:      organizationService,
		checkoutService: checkoutService,
		customerService: customerService,
		planService:     planService,
		userService:     userService,
	}
}

// EnsureDefaultPlan create a new customer account and subscribe to the default plan if configured
func (p *Processor) EnsureDefaultPlan(ctx context.Context, orgID string) error {
	if p.billingConf.DefaultPlan != "" && p.billingConf.DefaultCurrency != "" {
		// validate the plan requested is free
		defaultPlan, err := p.planService.GetByID(ctx, p.billingConf.DefaultPlan)
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
			Name:     org.Name,
			Email:    emailID,
			Currency: p.billingConf.DefaultCurrency,
			Metadata: map[string]any{
				"auto_created": "true",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
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
	return nil
}
