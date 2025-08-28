package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/customer"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
	Create(ctx context.Context, customer customer.Customer, offline bool) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
	UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (customer.Details, error)
	GetDetails(ctx context.Context, customerID string) (customer.Details, error)
	ListPaymentMethods(ctx context.Context, id string) ([]customer.PaymentMethod, error)
}

func (h *ConnectHandler) ListAllBillingAccounts(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllBillingAccountsRequest]) (*connect.Response[frontierv1beta1.ListAllBillingAccountsResponse], error) {
	var customers []*frontierv1beta1.BillingAccount
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range customerList {
		customerPB, err := transformCustomerToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.UpdateBillingAccountLimitsResponse{}), nil
}

func (h *ConnectHandler) GetBillingAccountDetails(ctx context.Context, request *connect.Request[frontierv1beta1.GetBillingAccountDetailsRequest]) (*connect.Response[frontierv1beta1.GetBillingAccountDetailsResponse], error) {
	details, err := h.customerService.GetDetails(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var paymentMethodsPbs []*frontierv1beta1.PaymentMethod
	if request.Msg.GetWithPaymentMethods() {
		pms, err := h.customerService.ListPaymentMethods(ctx, request.Msg.GetId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		for _, v := range pms {
			pmPB, err := transformPaymentMethodToPB(v)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
			paymentMethodsPbs = append(paymentMethodsPbs, pmPB)
		}
	}

	var billingDetailsPb *frontierv1beta1.BillingAccountDetails
	if request.Msg.GetWithBillingDetails() {
		billingDetails, err := h.customerService.GetDetails(ctx, request.Msg.GetId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		billingDetailsPb = &frontierv1beta1.BillingAccountDetails{
			CreditMin: billingDetails.CreditMin,
			DueInDays: billingDetails.DueInDays,
		}
	}

	customerPB, err := transformCustomerToPB(customerOb)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GetBillingAccountResponse{
		BillingAccount: customerPB,
		PaymentMethods: paymentMethodsPbs,
		BillingDetails: billingDetailsPb,
	}), nil
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
