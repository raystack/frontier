package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.CreateBillingAccountRequest]) (*connect.Response[frontierv1beta1.CreateBillingAccountResponse], error) {
	var stripeTestClockID *string
	if val, ok := customer.GetStripeTestClockFromContext(ctx); ok {
		stripeTestClockID = &val
	}

	var customerAddress customer.Address
	if request.Msg.GetBody().GetAddress() != nil {
		customerAddress = customer.Address{
			City:       request.Msg.GetBody().GetAddress().GetCity(),
			Country:    request.Msg.GetBody().GetAddress().GetCountry(),
			Line1:      request.Msg.GetBody().GetAddress().GetLine1(),
			Line2:      request.Msg.GetBody().GetAddress().GetLine2(),
			PostalCode: request.Msg.GetBody().GetAddress().GetPostalCode(),
			State:      request.Msg.GetBody().GetAddress().GetState(),
		}
	}
	var customerTaxes []customer.Tax
	if len(request.Msg.GetBody().GetTaxData()) > 0 {
		for _, tax := range request.Msg.GetBody().GetTaxData() {
			customerTaxes = append(customerTaxes, customer.Tax{
				Type: tax.GetType(),
				ID:   tax.GetId(),
			})
		}
	}
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	newCustomer, err := h.customerService.Create(ctx, customer.Customer{
		OrgID:             request.Msg.GetOrgId(),
		Name:              request.Msg.GetBody().GetName(),
		Email:             request.Msg.GetBody().GetEmail(),
		Phone:             request.Msg.GetBody().GetPhone(),
		Address:           customerAddress,
		Currency:          request.Msg.GetBody().GetCurrency(),
		Metadata:          metaDataMap,
		StripeTestClockID: stripeTestClockID,
		TaxData:           customerTaxes,
	}, request.Msg.GetOffline())
	if err != nil {
		if errors.Is(err, customer.ErrActiveConflict) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateBillingAccount.Create: org_id=%s customer_name=%s customer_email=%s currency=%s offline=%v: %w", request.Msg.GetOrgId(), request.Msg.GetBody().GetName(), request.Msg.GetBody().GetEmail(), request.Msg.GetBody().GetCurrency(), request.Msg.GetOffline(), err))
	}

	customerPB, err := transformCustomerToPB(newCustomer)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateBillingAccount: customer_id=%s: %w", newCustomer.ID, err))
	}
	return connect.NewResponse(&frontierv1beta1.CreateBillingAccountResponse{
		BillingAccount: customerPB,
	}), nil
}

func (h *ConnectHandler) UpdateBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateBillingAccountRequest]) (*connect.Response[frontierv1beta1.UpdateBillingAccountResponse], error) {
	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}
	var customerAddress customer.Address
	if request.Msg.GetBody().GetAddress() != nil {
		customerAddress = customer.Address{
			City:       request.Msg.GetBody().GetAddress().GetCity(),
			Country:    request.Msg.GetBody().GetAddress().GetCountry(),
			Line1:      request.Msg.GetBody().GetAddress().GetLine1(),
			Line2:      request.Msg.GetBody().GetAddress().GetLine2(),
			PostalCode: request.Msg.GetBody().GetAddress().GetPostalCode(),
			State:      request.Msg.GetBody().GetAddress().GetState(),
		}
	}
	var customerTaxes []customer.Tax
	if len(request.Msg.GetBody().GetTaxData()) > 0 {
		for _, tax := range request.Msg.GetBody().GetTaxData() {
			customerTaxes = append(customerTaxes, customer.Tax{
				Type: tax.GetType(),
				ID:   tax.GetId(),
			})
		}
	}

	// Ignore org_id from request - it will be inferred from billing account ID
	updatedCustomer, err := h.customerService.Update(ctx, customer.Customer{
		ID:       request.Msg.GetId(),
		Name:     request.Msg.GetBody().GetName(),
		Email:    request.Msg.GetBody().GetEmail(),
		Phone:    request.Msg.GetBody().GetPhone(),
		Currency: request.Msg.GetBody().GetCurrency(),
		Address:  customerAddress,
		Metadata: metaDataMap,
		TaxData:  customerTaxes,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateBillingAccount.Update: customer_id=%s customer_name=%s customer_email=%s currency=%s: %w", request.Msg.GetId(), request.Msg.GetBody().GetName(), request.Msg.GetBody().GetEmail(), request.Msg.GetBody().GetCurrency(), err))
	}

	customerPB, err := transformCustomerToPB(updatedCustomer)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateBillingAccount: customer_id=%s: %w", updatedCustomer.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.UpdateBillingAccountResponse{
		BillingAccount: customerPB,
	}), nil
}

func (h *ConnectHandler) RegisterBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.RegisterBillingAccountRequest]) (*connect.Response[frontierv1beta1.RegisterBillingAccountResponse], error) {
	_, err := h.customerService.RegisterToProviderIfRequired(ctx, request.Msg.GetId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, ErrCustomerNotFound)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RegisterBillingAccount.RegisterToProviderIfRequired: customer_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.RegisterBillingAccountResponse{}), nil
}

func (h *ConnectHandler) ListBillingAccounts(ctx context.Context, request *connect.Request[frontierv1beta1.ListBillingAccountsRequest]) (*connect.Response[frontierv1beta1.ListBillingAccountsResponse], error) {
	if request.Msg.GetOrgId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	var customers []*frontierv1beta1.BillingAccount
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListBillingAccounts.List: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}
	for _, v := range customerList {
		customerPB, err := transformCustomerToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListBillingAccounts: customer_id=%s: %w", v.ID, err))
		}
		customers = append(customers, customerPB)
	}

	response := &frontierv1beta1.ListBillingAccountsResponse{
		BillingAccounts: customers,
	}

	// Handle response enrichment based on expand field
	response = h.enrichListBillingAccountsResponse(ctx, request.Msg, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) DeleteBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteBillingAccountRequest]) (*connect.Response[frontierv1beta1.DeleteBillingAccountResponse], error) {
	err := h.customerService.Delete(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteBillingAccount.Delete: customer_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.DeleteBillingAccountResponse{}), nil
}

func (h *ConnectHandler) EnableBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.EnableBillingAccountRequest]) (*connect.Response[frontierv1beta1.EnableBillingAccountResponse], error) {
	err := h.customerService.Enable(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("EnableBillingAccount.Enable: customer_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.EnableBillingAccountResponse{}), nil
}

func (h *ConnectHandler) DisableBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.DisableBillingAccountRequest]) (*connect.Response[frontierv1beta1.DisableBillingAccountResponse], error) {
	err := h.customerService.Disable(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DisableBillingAccount.Disable: customer_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.DisableBillingAccountResponse{}), nil
}

func (h *ConnectHandler) GetBillingBalance(ctx context.Context, request *connect.Request[frontierv1beta1.GetBillingBalanceRequest]) (*connect.Response[frontierv1beta1.GetBillingBalanceResponse], error) {
	balanceAmount, err := h.creditService.GetBalance(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingBalance.GetBalance: customer_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.GetBillingBalanceResponse{
		Balance: &frontierv1beta1.BillingAccount_Balance{
			Amount:   balanceAmount,
			Currency: "VC",
		},
	}), nil
}

func (h *ConnectHandler) HasTrialed(ctx context.Context, request *connect.Request[frontierv1beta1.HasTrialedRequest]) (*connect.Response[frontierv1beta1.HasTrialedResponse], error) {
	hasTrialed, err := h.subscriptionService.HasUserSubscribedBefore(ctx, request.Msg.GetId(), request.Msg.GetPlanId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("HasTrialed.HasUserSubscribedBefore: customer_id=%s plan_id=%s: %w", request.Msg.GetId(), request.Msg.GetPlanId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.HasTrialedResponse{
		Trialed: hasTrialed,
	}), nil
}

func (h *ConnectHandler) ListAllBillingAccounts(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllBillingAccountsRequest]) (*connect.Response[frontierv1beta1.ListAllBillingAccountsResponse], error) {
	var customers []*frontierv1beta1.BillingAccount
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListAllBillingAccounts.List: org_id=%s: %w", request.Msg.GetOrgId(), err))
	}
	for _, v := range customerList {
		customerPB, err := transformCustomerToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListAllBillingAccounts: customer_id=%s: %w", v.ID, err))
		}
		customers = append(customers, customerPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListAllBillingAccountsResponse{
		BillingAccounts: customers,
	}), nil
}

func transformCustomerToPB(customer customer.Customer) (*frontierv1beta1.BillingAccount, error) {
	metaData, err := customer.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.BillingAccount{}, err
	}
	taxData := make([]*frontierv1beta1.BillingAccount_Tax, 0, len(customer.TaxData))
	for _, tax := range customer.TaxData {
		taxData = append(taxData, &frontierv1beta1.BillingAccount_Tax{
			Type: tax.Type,
			Id:   tax.ID,
		})
	}
	return &frontierv1beta1.BillingAccount{
		Id:         customer.ID,
		OrgId:      customer.OrgID,
		Name:       customer.Name,
		Email:      customer.Email,
		Phone:      customer.Phone,
		Currency:   customer.Currency,
		ProviderId: customer.ProviderID,
		Address: &frontierv1beta1.BillingAccount_Address{
			City:       customer.Address.City,
			Country:    customer.Address.Country,
			Line1:      customer.Address.Line1,
			Line2:      customer.Address.Line2,
			PostalCode: customer.Address.PostalCode,
			State:      customer.Address.State,
		},
		TaxData:   taxData,
		State:     customer.State.String(),
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
		Metadata:  metaData,
	}, nil
}

func (h *ConnectHandler) UpdateBillingAccountLimits(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateBillingAccountLimitsRequest]) (*connect.Response[frontierv1beta1.UpdateBillingAccountLimitsResponse], error) {
	_, err := h.customerService.UpdateCreditMinByID(ctx, request.Msg.GetId(), request.Msg.GetCreditMin())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateBillingAccountLimits.UpdateCreditMinByID: customer_id=%s credit_min=%d: %w", request.Msg.GetId(), request.Msg.GetCreditMin(), err))
	}

	return connect.NewResponse(&frontierv1beta1.UpdateBillingAccountLimitsResponse{}), nil
}

func (h *ConnectHandler) GetBillingAccountDetails(ctx context.Context, request *connect.Request[frontierv1beta1.GetBillingAccountDetailsRequest]) (*connect.Response[frontierv1beta1.GetBillingAccountDetailsResponse], error) {
	details, err := h.customerService.GetDetails(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingAccountDetails.GetDetails: customer_id=%s: %w", request.Msg.GetId(), err))
	}

	return connect.NewResponse(&frontierv1beta1.GetBillingAccountDetailsResponse{
		CreditMin: details.CreditMin,
		DueInDays: details.DueInDays,
	}), nil
}

func (h *ConnectHandler) GetBillingAccount(ctx context.Context, request *connect.Request[frontierv1beta1.GetBillingAccountRequest]) (*connect.Response[frontierv1beta1.GetBillingAccountResponse], error) {
	customerOb, err := h.customerService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingAccount.GetByID: customer_id=%s: %w", request.Msg.GetId(), err))
	}

	var paymentMethodsPbs []*frontierv1beta1.PaymentMethod
	if request.Msg.GetWithPaymentMethods() {
		pms, err := h.customerService.ListPaymentMethods(ctx, request.Msg.GetId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingAccount.ListPaymentMethods: customer_id=%s: %w", request.Msg.GetId(), err))
		}
		for _, v := range pms {
			pmPB, err := transformPaymentMethodToPB(v)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingAccount: payment_method_id=%s: %w", v.ID, err))
			}
			paymentMethodsPbs = append(paymentMethodsPbs, pmPB)
		}
	}

	var billingDetailsPb *frontierv1beta1.BillingAccountDetails
	if request.Msg.GetWithBillingDetails() {
		billingDetails, err := h.customerService.GetDetails(ctx, request.Msg.GetId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingAccount.GetDetails: customer_id=%s: %w", request.Msg.GetId(), err))
		}
		billingDetailsPb = &frontierv1beta1.BillingAccountDetails{
			CreditMin: billingDetails.CreditMin,
			DueInDays: billingDetails.DueInDays,
		}
	}

	customerPB, err := transformCustomerToPB(customerOb)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetBillingAccount: customer_id=%s: %w", customerOb.ID, err))
	}

	response := &frontierv1beta1.GetBillingAccountResponse{
		BillingAccount: customerPB,
		PaymentMethods: paymentMethodsPbs,
		BillingDetails: billingDetailsPb,
	}

	// Handle response enrichment based on expand field
	response = h.enrichGetBillingAccountResponse(ctx, request.Msg, response)

	return connect.NewResponse(response), nil
}

func transformPaymentMethodToPB(pm customer.PaymentMethod) (*frontierv1beta1.PaymentMethod, error) {
	metaData, err := pm.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.PaymentMethod{}, err
	}
	return &frontierv1beta1.PaymentMethod{
		Id:              pm.ID,
		CustomerId:      pm.CustomerID,
		ProviderId:      pm.ProviderID,
		Type:            pm.Type,
		CardLast4:       pm.CardLast4,
		CardBrand:       pm.CardBrand,
		CardExpiryMonth: pm.CardExpiryMonth,
		CardExpiryYear:  pm.CardExpiryYear,
		Metadata:        metaData,
		CreatedAt:       timestamppb.New(pm.CreatedAt),
	}, nil
}

func (h *ConnectHandler) UpdateBillingAccountDetails(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateBillingAccountDetailsRequest]) (*connect.Response[frontierv1beta1.UpdateBillingAccountDetailsResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.GetDueInDays() < 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot create predated invoices: due in days should be greater than 0"))
	}

	details, err := h.customerService.UpdateDetails(ctx, request.Msg.GetId(), customer.Details{
		CreditMin: request.Msg.GetCreditMin(),
		DueInDays: request.Msg.GetDueInDays(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateBillingAccountDetails.UpdateDetails: customer_id=%s credit_min=%d due_in_days=%d: %w", request.Msg.GetId(), request.Msg.GetCreditMin(), request.Msg.GetDueInDays(), err))
	}

	// Add audit log - infer org_id from billing account
	customerOb, err := h.customerService.GetByID(ctx, request.Msg.GetId())
	if err == nil {
		if err := audit.GetAuditor(ctx, customerOb.OrgID).LogWithAttrs(audit.BillingAccountDetailsUpdatedEvent, audit.Target{
			ID:   request.Msg.GetId(),
			Type: "billing_account",
		}, map[string]string{
			"credit_min":  fmt.Sprintf("%d", details.CreditMin),
			"due_in_days": fmt.Sprintf("%d", details.DueInDays),
		}); err != nil {
			errorLogger.LogServiceError(ctx, request, "UpdateBillingAccountDetails.AuditLog", err,
				"customer_id", request.Msg.GetId())
		}
	} else {
		errorLogger.LogServiceError(ctx, request, "UpdateBillingAccountDetails.GetByID", err,
			"customer_id", request.Msg.GetId())
	}

	return connect.NewResponse(&frontierv1beta1.UpdateBillingAccountDetailsResponse{}), nil
}

// enrichGetBillingAccountResponse enriches the response with expanded fields
func (h *ConnectHandler) enrichGetBillingAccountResponse(ctx context.Context, req *frontierv1beta1.GetBillingAccountRequest, resp *frontierv1beta1.GetBillingAccountResponse) *frontierv1beta1.GetBillingAccountResponse {
	expandModels := parseExpandModels(req)
	if len(expandModels) == 0 {
		// no need to enrich the response
		return resp
	}

	if (expandModels["organization"] || expandModels["org"]) && resp.GetBillingAccount() != nil {
		org, _ := h.GetOrganization(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{
			Id: resp.GetBillingAccount().GetOrgId(),
		}))
		if org != nil && org.Msg != nil {
			resp.BillingAccount.Organization = org.Msg.GetOrganization()
		}
	}

	return resp
}

// enrichListBillingAccountsResponse enriches the response with expanded fields
func (h *ConnectHandler) enrichListBillingAccountsResponse(ctx context.Context, req *frontierv1beta1.ListBillingAccountsRequest, resp *frontierv1beta1.ListBillingAccountsResponse) *frontierv1beta1.ListBillingAccountsResponse {
	expandModels := parseExpandModels(req)
	if len(expandModels) == 0 {
		// no need to enrich the response
		return resp
	}

	if len(resp.GetBillingAccounts()) > 0 {
		for baIdx, ba := range resp.GetBillingAccounts() {
			if expandModels["organization"] || expandModels["org"] {
				org, _ := h.GetOrganization(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{
					Id: ba.GetOrgId(),
				}))
				if org != nil && org.Msg != nil {
					resp.BillingAccounts[baIdx].Organization = org.Msg.GetOrganization()
				}
			}
		}
	}

	return resp
}
