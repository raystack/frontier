package interceptors

import (
	"context"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/pkg/server/consts"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/internal/api/v1beta1"
	"google.golang.org/grpc"
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
		return handler(ctx, req)
	}
}

func APIRequestEnrich(ctx context.Context, handler *v1beta1.Handler, methodName string, req any) (any, error) {
	// convert project ids to org ids for billing endpoints:
	// - CreateBillingUsage
	// - CheckFeatureEntitlement
	switch methodName {
	case "/raystack.frontier.v1beta1.FrontierService/CheckFeatureEntitlement":
		req := req.(*frontierv1beta1.CheckFeatureEntitlementRequest)
		if req.GetProjectId() != "" && req.GetOrgId() != "" {
			return req, status.Error(codes.InvalidArgument, "both project_id and org_id cannot be provided")
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
			return req, status.Error(codes.InvalidArgument, "both project_id and org_id cannot be provided")
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
	// - ListInvoices
	// - CreateBillingUsage
	// - GetUpcomingInvoice
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
	case "/raystack.frontier.v1beta1.FrontierService/ListInvoices":
		req := req.(*frontierv1beta1.ListInvoicesRequest)
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
	case "/raystack.frontier.v1beta1.FrontierService/GetUpcomingInvoice":
		req := req.(*frontierv1beta1.GetUpcomingInvoiceRequest)
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
	}
	return ctx
}
