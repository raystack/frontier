package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/raystack/frontier/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	grpcCustomerNotFoundErr = status.Errorf(codes.NotFound, "customer doesn't exist")
)

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
	Create(ctx context.Context, customer customer.Customer, offline bool) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
	Delete(ctx context.Context, id string) error
	ListPaymentMethods(ctx context.Context, id string) ([]customer.PaymentMethod, error)
	Update(ctx context.Context, customer customer.Customer) (customer.Customer, error)
	RegisterToProviderIfRequired(ctx context.Context, customerID string) (customer.Customer, error)
	Disable(ctx context.Context, id string) error
	Enable(ctx context.Context, id string) error
	UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (customer.Details, error)
	GetDetails(ctx context.Context, customerID string) (customer.Details, error)
	UpdateDetails(ctx context.Context, customerID string, details customer.Details) (customer.Details, error)
}

func (h Handler) CreateBillingAccount(ctx context.Context, request *frontierv1beta1.CreateBillingAccountRequest) (*frontierv1beta1.CreateBillingAccountResponse, error) {
	var stripeTestClockID *string
	if val, ok := customer.GetStripeTestClockFromContext(ctx); ok {
		stripeTestClockID = &val
	}

	var customerAddress customer.Address
	if request.GetBody().GetAddress() != nil {
		customerAddress = customer.Address{
			City:       request.GetBody().GetAddress().GetCity(),
			Country:    request.GetBody().GetAddress().GetCountry(),
			Line1:      request.GetBody().GetAddress().GetLine1(),
			Line2:      request.GetBody().GetAddress().GetLine2(),
			PostalCode: request.GetBody().GetAddress().GetPostalCode(),
			State:      request.GetBody().GetAddress().GetState(),
		}
	}
	var customerTaxes []customer.Tax
	if len(request.GetBody().GetTaxData()) > 0 {
		for _, tax := range request.GetBody().GetTaxData() {
			customerTaxes = append(customerTaxes, customer.Tax{
				Type: tax.GetType(),
				ID:   tax.GetId(),
			})
		}
	}
	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())
	newCustomer, err := h.customerService.Create(ctx, customer.Customer{
		OrgID:             request.GetOrgId(),
		Name:              request.GetBody().GetName(),
		Email:             request.GetBody().GetEmail(),
		Phone:             request.GetBody().GetPhone(),
		Address:           customerAddress,
		Currency:          request.GetBody().GetCurrency(),
		Metadata:          metaDataMap,
		StripeTestClockID: stripeTestClockID,
		TaxData:           customerTaxes,
	}, request.GetOffline())
	if err != nil {
		if errors.Is(err, customer.ErrActiveConflict) {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		return nil, err
	}

	customerPB, err := transformCustomerToPB(newCustomer)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateBillingAccountResponse{
		BillingAccount: customerPB,
	}, nil
}

func (h Handler) ListAllBillingAccounts(ctx context.Context, request *frontierv1beta1.ListAllBillingAccountsRequest) (*frontierv1beta1.ListAllBillingAccountsResponse, error) {
	var customers []*frontierv1beta1.BillingAccount
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.GetOrgId(),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range customerList {
		customerPB, err := transformCustomerToPB(v)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customerPB)
	}

	return &frontierv1beta1.ListAllBillingAccountsResponse{
		BillingAccounts: customers,
	}, nil
}

func (h Handler) ListBillingAccounts(ctx context.Context, request *frontierv1beta1.ListBillingAccountsRequest) (*frontierv1beta1.ListBillingAccountsResponse, error) {
	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}
	var customers []*frontierv1beta1.BillingAccount
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.GetOrgId(),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range customerList {
		customerPB, err := transformCustomerToPB(v)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customerPB)
	}

	return &frontierv1beta1.ListBillingAccountsResponse{
		BillingAccounts: customers,
	}, nil
}

func (h Handler) GetBillingAccount(ctx context.Context, request *frontierv1beta1.GetBillingAccountRequest) (*frontierv1beta1.GetBillingAccountResponse, error) {
	customerOb, err := h.customerService.GetByID(ctx, request.GetId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return nil, grpcCustomerNotFoundErr
		}
		return nil, err
	}

	var paymentMethodsPbs []*frontierv1beta1.PaymentMethod
	if request.GetWithPaymentMethods() {
		pms, err := h.customerService.ListPaymentMethods(ctx, request.GetId())
		if err != nil {
			return nil, err
		}
		for _, v := range pms {
			pmPB, err := transformPaymentMethodToPB(v)
			if err != nil {
				return nil, err
			}
			paymentMethodsPbs = append(paymentMethodsPbs, pmPB)
		}
	}

	customerPB, err := transformCustomerToPB(customerOb)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetBillingAccountResponse{
		BillingAccount: customerPB,
		PaymentMethods: paymentMethodsPbs,
	}, nil
}

func (h Handler) DeleteBillingAccount(ctx context.Context, request *frontierv1beta1.DeleteBillingAccountRequest) (*frontierv1beta1.DeleteBillingAccountResponse, error) {
	err := h.customerService.Delete(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.DeleteBillingAccountResponse{}, nil
}

func (h Handler) GetBillingBalance(ctx context.Context, request *frontierv1beta1.GetBillingBalanceRequest) (*frontierv1beta1.GetBillingBalanceResponse, error) {
	balanceAmount, err := h.creditService.GetBalance(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetBillingBalanceResponse{
		Balance: &frontierv1beta1.BillingAccount_Balance{
			Amount:   balanceAmount,
			Currency: "VC",
		},
	}, nil
}

func (h Handler) UpdateBillingAccount(ctx context.Context, request *frontierv1beta1.UpdateBillingAccountRequest) (*frontierv1beta1.UpdateBillingAccountResponse, error) {
	var metaDataMap metadata.Metadata
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())
	}
	var customerAddress customer.Address
	if request.GetBody().GetAddress() != nil {
		customerAddress = customer.Address{
			City:       request.GetBody().GetAddress().GetCity(),
			Country:    request.GetBody().GetAddress().GetCountry(),
			Line1:      request.GetBody().GetAddress().GetLine1(),
			Line2:      request.GetBody().GetAddress().GetLine2(),
			PostalCode: request.GetBody().GetAddress().GetPostalCode(),
			State:      request.GetBody().GetAddress().GetState(),
		}
	}
	var customerTaxes []customer.Tax
	if len(request.GetBody().GetTaxData()) > 0 {
		for _, tax := range request.GetBody().GetTaxData() {
			customerTaxes = append(customerTaxes, customer.Tax{
				Type: tax.GetType(),
				ID:   tax.GetId(),
			})
		}
	}

	updatedCustomer, err := h.customerService.Update(ctx, customer.Customer{
		ID:       request.GetId(),
		OrgID:    request.GetOrgId(),
		Name:     request.GetBody().GetName(),
		Email:    request.GetBody().GetEmail(),
		Phone:    request.GetBody().GetPhone(),
		Currency: request.GetBody().GetCurrency(),
		Address:  customerAddress,
		Metadata: metaDataMap,
		TaxData:  customerTaxes,
	})
	if err != nil {
		return nil, err
	}

	customerPB, err := transformCustomerToPB(updatedCustomer)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.UpdateBillingAccountResponse{
		BillingAccount: customerPB,
	}, nil
}

func (h Handler) RegisterBillingAccount(ctx context.Context, request *frontierv1beta1.RegisterBillingAccountRequest) (*frontierv1beta1.RegisterBillingAccountResponse, error) {
	_, err := h.customerService.RegisterToProviderIfRequired(ctx, request.GetId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return nil, grpcCustomerNotFoundErr
		}
		return nil, err
	}
	return &frontierv1beta1.RegisterBillingAccountResponse{}, nil
}

func (h Handler) DisableBillingAccount(ctx context.Context, request *frontierv1beta1.DisableBillingAccountRequest) (*frontierv1beta1.DisableBillingAccountResponse, error) {
	err := h.customerService.Disable(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.DisableBillingAccountResponse{}, nil
}

func (h Handler) EnableBillingAccount(ctx context.Context, request *frontierv1beta1.EnableBillingAccountRequest) (*frontierv1beta1.EnableBillingAccountResponse, error) {
	err := h.customerService.Enable(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.EnableBillingAccountResponse{}, nil
}

// GetRequestCustomerID returns the customer id from the request via reflection(billing_id/id)
// or defaults to first customer id present in the org
func (h Handler) GetRequestCustomerID(ctx context.Context, request any) (string, error) {
	requestValue := reflect.Indirect(reflect.ValueOf(request))
	reqBillingIDValue := requestValue.FieldByName("BillingId")
	if reqBillingIDValue.IsValid() {
		reqBillingID := reqBillingIDValue.String()
		if utils.IsValidUUID(reqBillingID) {
			return reqBillingID, nil
		}
	} else {
		reqIDValue := requestValue.FieldByName("Id")
		if reqIDValue.IsValid() {
			reqID := reqIDValue.String()
			if utils.IsValidUUID(reqID) {
				return reqID, nil
			}
		}
	}
	reqOrgID := requestValue.FieldByName("OrgId")
	if reqOrgID.IsValid() {
		org, err := h.orgService.Get(ctx, reqOrgID.String())
		if err != nil {
			return "", fmt.Errorf("no org found with id %s", reqOrgID.String())
		}
		// no id found, find default customer id
		customers, err := h.customerService.List(ctx, customer.Filter{
			OrgID: org.ID,
			State: customer.ActiveState,
		})
		if err != nil {
			return "", fmt.Errorf("error listing customers for org %s: %w", org.ID, err)
		}
		if len(customers) == 0 {
			return "", fmt.Errorf("no customer found for org %s", org.ID)
		}
		return customers[0].ID, nil
	}
	return "", fmt.Errorf("no billing id or org id found in request")
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

func (h Handler) HasTrialed(ctx context.Context, request *frontierv1beta1.HasTrialedRequest) (*frontierv1beta1.HasTrialedResponse, error) {
	hasTrialed, err := h.subscriptionService.HasUserSubscribedBefore(ctx, request.GetId(), request.GetPlanId())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.HasTrialedResponse{
		Trialed: hasTrialed,
	}, nil
}

func (h Handler) UpdateBillingAccountLimits(ctx context.Context,
	request *frontierv1beta1.UpdateBillingAccountLimitsRequest) (*frontierv1beta1.UpdateBillingAccountLimitsResponse, error) {
	_, err := h.customerService.UpdateCreditMinByID(ctx, request.GetId(), request.GetCreditMin())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.UpdateBillingAccountLimitsResponse{}, nil
}

func (h Handler) GetBillingAccountDetails(ctx context.Context,
	request *frontierv1beta1.GetBillingAccountDetailsRequest) (*frontierv1beta1.GetBillingAccountDetailsResponse, error) {
	details, err := h.customerService.GetDetails(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetBillingAccountDetailsResponse{
		CreditMin: details.CreditMin,
		DueInDays: details.DueInDays,
	}, nil
}

func (h Handler) UpdateBillingAccountDetails(ctx context.Context,
	request *frontierv1beta1.UpdateBillingAccountDetailsRequest) (*frontierv1beta1.UpdateBillingAccountDetailsResponse, error) {
	if request.GetDueInDays() < 0 {
		return nil, status.Errorf(codes.FailedPrecondition,
			"cannot create predated invoices: due in days shoule be greated than 0")
	}

	_, err := h.customerService.UpdateDetails(ctx, request.GetId(), customer.Details{
		CreditMin: request.GetCreditMin(),
		DueInDays: request.GetDueInDays(),
	})
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.UpdateBillingAccountDetailsResponse{}, nil
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
