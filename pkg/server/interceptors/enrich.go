package interceptors

import (
	"context"
	"reflect"
	"strings"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/pkg/server/consts"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/internal/api/v1beta1"
	"google.golang.org/grpc"
)

var (
	ErrStatusOrgProjectMismatch = status.Error(codes.InvalidArgument, "both project_id and org_id cannot be provided")
)

// UnaryAPIRequestEnrich is a unary server interceptor that enriches the request context with additional information
func UnaryAPIRequestEnrich() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		serverHandler, ok := info.Server.(*v1beta1.Handler)
		if !ok {
			// no need to capture the request
			return handler(ctx, req)
		}

		req, err = APIRequestEnrich(ctx, serverHandler, info.FullMethod, req)
		if err != nil {
			return nil, err
		}

		// populate stripe key if applicable
		ctx = UnaryCtxWithStripeTestClock(ctx, serverHandler, info.FullMethod)
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}

		resp = APIResponseEnrich(ctx, serverHandler, info.FullMethod, req, resp)
		return resp, nil
	}
}

func APIRequestEnrich(ctx context.Context, handler *v1beta1.Handler, methodName string, req any) (any, error) {
	// convert project ids to org ids for billing endpoints:
	// - CheckFeatureEntitlement
	// - CreateBillingUsage
	// - RevertBillingUsage
	switch methodName {
	case "/raystack.frontier.v1beta1.FrontierService/CheckFeatureEntitlement":
		req := req.(*frontierv1beta1.CheckFeatureEntitlementRequest)
		if req.GetProjectId() != "" && req.GetOrgId() != "" {
			return req, ErrStatusOrgProjectMismatch
		}

		if req.GetProjectId() != "" {
			proj, err := handler.GetProject(ctx, &frontierv1beta1.GetProjectRequest{
				Id: req.GetProjectId(),
			})
			if err != nil {
				return req, err
			}
			req.OrgId = proj.GetProject().GetOrgId()
		}
	case "/raystack.frontier.v1beta1.FrontierService/CreateBillingUsage":
		req := req.(*frontierv1beta1.CreateBillingUsageRequest)
		if req.GetProjectId() != "" && req.GetOrgId() != "" {
			return req, ErrStatusOrgProjectMismatch
		}

		if req.GetProjectId() != "" {
			proj, err := handler.GetProject(ctx, &frontierv1beta1.GetProjectRequest{
				Id: req.GetProjectId(),
			})
			if err != nil {
				return req, err
			}
			req.OrgId = proj.GetProject().GetOrgId()
		}
	case "/raystack.frontier.v1beta1.AdminService/RevertBillingUsage":
		req := req.(*frontierv1beta1.RevertBillingUsageRequest)
		if req.GetProjectId() != "" && req.GetOrgId() != "" {
			return req, ErrStatusOrgProjectMismatch
		}

		if req.GetProjectId() != "" {
			proj, err := handler.GetProject(ctx, &frontierv1beta1.GetProjectRequest{
				Id: req.GetProjectId(),
			})
			if err != nil {
				return req, err
			}
			req.OrgId = proj.GetProject().GetOrgId()
		}
	}

	// find default billing account for the customer if needed
	// - CheckFeatureEntitlement
	// - CreateBillingUsage
	// - ListInvoices
	// - GetUpcomingInvoice
	// - RevertBillingUsage
	switch methodName {
	case "/raystack.frontier.v1beta1.FrontierService/CheckFeatureEntitlement":
		req := req.(*frontierv1beta1.CheckFeatureEntitlementRequest)
		if req.GetBillingId() == "" {
			customerID, err := handler.GetRequestCustomerID(ctx, req)
			if err != nil {
				return req, status.Error(codes.InvalidArgument, err.Error())
			}
			req.BillingId = customerID
		}
	case "/raystack.frontier.v1beta1.FrontierService/CreateBillingUsage":
		req := req.(*frontierv1beta1.CreateBillingUsageRequest)
		if req.GetBillingId() == "" {
			customerID, err := handler.GetRequestCustomerID(ctx, req)
			if err != nil {
				return req, status.Error(codes.InvalidArgument, err.Error())
			}
			req.BillingId = customerID
		}
	case "/raystack.frontier.v1beta1.FrontierService/ListInvoices":
		req := req.(*frontierv1beta1.ListInvoicesRequest)
		if req.GetBillingId() == "" {
			customerID, err := handler.GetRequestCustomerID(ctx, req)
			if err != nil {
				return req, status.Error(codes.InvalidArgument, err.Error())
			}
			req.BillingId = customerID
		}
	case "/raystack.frontier.v1beta1.FrontierService/GetUpcomingInvoice":
		req := req.(*frontierv1beta1.GetUpcomingInvoiceRequest)
		if req.GetBillingId() == "" {
			customerID, err := handler.GetRequestCustomerID(ctx, req)
			if err != nil {
				return req, status.Error(codes.InvalidArgument, err.Error())
			}
			req.BillingId = customerID
		}
	case "/raystack.frontier.v1beta1.AdminService/RevertBillingUsage":
		req := req.(*frontierv1beta1.RevertBillingUsageRequest)
		if req.GetBillingId() == "" {
			customerID, err := handler.GetRequestCustomerID(ctx, req)
			if err != nil {
				return req, status.Error(codes.InvalidArgument, err.Error())
			}
			req.BillingId = customerID
		}
	}
	return req, nil
}

// APIResponseEnrich is a unary server interceptor that enriches the response context with additional information
// based on the request expand field
func APIResponseEnrich(ctx context.Context, handler *v1beta1.Handler, methodName string, req, resp any) any {
	// find the strings mentioned in the request under Expand field
	// we will only enrich the response if the Expand field is set for the specific model
	expandModels := map[string]bool{}
	expandReflect := reflect.ValueOf(req).Elem().FieldByName("Expand")
	if expandReflect.IsValid() && expandReflect.Len() > 0 {
		for i := 0; i < expandReflect.Len(); i++ {
			expandModels[strings.ToLower(expandReflect.Index(i).String())] = true
		}
	}
	if len(expandModels) == 0 {
		// no need to enrich the response
		return resp
	}

	switch methodName {
	case "/raystack.frontier.v1beta1.FrontierService/ListBillingTransactions":
		resp := resp.(*frontierv1beta1.ListBillingTransactionsResponse)
		if len(resp.GetTransactions()) > 0 {
			for tIdx, t := range resp.GetTransactions() {
				if expandModels["customer"] && len(t.GetCustomerId()) > 0 {
					ba, _ := handler.GetBillingAccount(ctx, &frontierv1beta1.GetBillingAccountRequest{
						Id: t.GetCustomerId(),
					})
					resp.Transactions[tIdx].Customer = ba.GetBillingAccount()
				}

				if expandModels["user"] && len(t.GetUserId()) > 0 {
					// if we allowed anyone to report usage with a user id, a bad actor can pass any user id
					// and retrieve user information.
					user, _ := handler.GetUser(ctx, &frontierv1beta1.GetUserRequest{
						Id: t.GetUserId(),
					})
					resp.Transactions[tIdx].User = user.GetUser()
				}
			}
		}
	case "/raystack.frontier.v1beta1.FrontierService/GetBillingAccount":
		resp := resp.(*frontierv1beta1.GetBillingAccountResponse)
		if (expandModels["organization"] || expandModels["org"]) && resp.GetBillingAccount() != nil {
			org, _ := handler.GetOrganization(ctx, &frontierv1beta1.GetOrganizationRequest{
				Id: resp.GetBillingAccount().GetOrgId(),
			})
			resp.BillingAccount.Organization = org.GetOrganization()
		}
	case "/raystack.frontier.v1beta1.FrontierService/ListBillingAccounts":
		resp := resp.(*frontierv1beta1.ListBillingAccountsResponse)
		if len(resp.GetBillingAccounts()) > 0 {
			for baIdx, ba := range resp.GetBillingAccounts() {
				if expandModels["organization"] || expandModels["org"] {
					org, _ := handler.GetOrganization(ctx, &frontierv1beta1.GetOrganizationRequest{
						Id: ba.GetOrgId(),
					})
					resp.BillingAccounts[baIdx].Organization = org.GetOrganization()
				}
			}
		}
	case "/raystack.frontier.v1beta1.FrontierService/GetSubscription":
		resp := resp.(*frontierv1beta1.GetSubscriptionResponse)
		if resp.GetSubscription() == nil {
			if expandModels["customer"] {
				ba, _ := handler.GetBillingAccount(ctx, &frontierv1beta1.GetBillingAccountRequest{
					Id: resp.GetSubscription().GetCustomerId(),
				})
				resp.Subscription.Customer = ba.GetBillingAccount()
			}
			if expandModels["plan"] {
				plan, _ := handler.GetPlan(ctx, &frontierv1beta1.GetPlanRequest{
					Id: resp.GetSubscription().GetPlanId(),
				})
				resp.Subscription.Plan = plan.GetPlan()
			}
		}
	case "/raystack.frontier.v1beta1.FrontierService/ListSubscriptions":
		resp := resp.(*frontierv1beta1.ListSubscriptionsResponse)
		if len(resp.GetSubscriptions()) > 0 {
			for sIdx, s := range resp.GetSubscriptions() {
				if expandModels["customer"] {
					ba, _ := handler.GetBillingAccount(ctx, &frontierv1beta1.GetBillingAccountRequest{
						Id: s.GetCustomerId(),
					})
					resp.Subscriptions[sIdx].Customer = ba.GetBillingAccount()
				}
				if expandModels["plan"] {
					plan, _ := handler.GetPlan(ctx, &frontierv1beta1.GetPlanRequest{
						Id: s.GetPlanId(),
					})
					resp.Subscriptions[sIdx].Plan = plan.GetPlan()
				}
			}
		}
	case "/raystack.frontier.v1beta1.FrontierService/ListInvoices":
		resp := resp.(*frontierv1beta1.ListInvoicesResponse)
		if len(resp.GetInvoices()) > 0 {
			for iIdx, i := range resp.GetInvoices() {
				if expandModels["customer"] {
					ba, _ := handler.GetBillingAccount(ctx, &frontierv1beta1.GetBillingAccountRequest{
						Id: i.GetCustomerId(),
					})
					resp.Invoices[iIdx].Customer = ba.GetBillingAccount()
				}
			}
		}
	}
	return resp
}

// UnaryCtxWithStripeTestClock adds stripe test clock id to context
func UnaryCtxWithStripeTestClock(ctx context.Context, handler *v1beta1.Handler, methodName string) context.Context {
	switch methodName {
	// stripeTestClockEnabledEndpoints is a map of endpoints that are allowed to use stripe test clock
	case "/raystack.frontier.v1beta1.FrontierService/CreateBillingAccount":
		if handler.IsSuperUser(ctx) == nil {
			// superuser can simulate stripe test clock if needed
			values := metadata.ValueFromIncomingContext(ctx, consts.StripeTestClockRequestKey)
			if len(values) > 0 {
				ctx = customer.SetStripeTestClockInContext(ctx, values[0])
			}
		}
	case "/raystack.frontier.v1beta1.FrontierService/BillingWebhookCallback":
		values := metadata.ValueFromIncomingContext(ctx, consts.StripeWebhookSignature)
		if len(values) > 0 {
			ctx = customer.SetStripeWebhookSignatureInContext(ctx, values[0])
		}
	}
	return ctx
}
