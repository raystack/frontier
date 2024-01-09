package v1beta1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
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
	Create(ctx context.Context, customer customer.Customer) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
	Delete(ctx context.Context, id string) error
	ListPaymentMethods(ctx context.Context, id string) ([]customer.PaymentMethod, error)
	Update(ctx context.Context, customer customer.Customer) (customer.Customer, error)
}

func (h Handler) CreateBillingAccount(ctx context.Context, request *frontierv1beta1.CreateBillingAccountRequest) (*frontierv1beta1.CreateBillingAccountResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())
	newCustomer, err := h.customerService.Create(ctx, customer.Customer{
		OrgID: request.GetOrgId(),
		Name:  request.GetBody().GetName(),
		Email: request.GetBody().GetEmail(),
		Phone: request.GetBody().GetPhone(),
		Address: customer.Address{
			City:       request.GetBody().GetAddress().GetCity(),
			Country:    request.GetBody().GetAddress().GetCountry(),
			Line1:      request.GetBody().GetAddress().GetLine1(),
			Line2:      request.GetBody().GetAddress().GetLine2(),
			PostalCode: request.GetBody().GetAddress().GetPostalCode(),
			State:      request.GetBody().GetAddress().GetState(),
		},
		Currency: request.GetBody().GetCurrency(),
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	customerPB, err := transformCustomerToPB(newCustomer)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.CreateBillingAccountResponse{
		BillingAccount: customerPB,
	}, nil
}

func (h Handler) ListBillingAccounts(ctx context.Context, request *frontierv1beta1.ListBillingAccountsRequest) (*frontierv1beta1.ListBillingAccountsResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}
	var customers []*frontierv1beta1.BillingAccount
	customerList, err := h.customerService.List(ctx, customer.Filter{
		OrgID: request.GetOrgId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range customerList {
		customerPB, err := transformCustomerToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		customers = append(customers, customerPB)
	}

	return &frontierv1beta1.ListBillingAccountsResponse{
		BillingAccounts: customers,
	}, nil
}

func (h Handler) GetBillingAccount(ctx context.Context, request *frontierv1beta1.GetBillingAccountRequest) (*frontierv1beta1.GetBillingAccountResponse, error) {
	logger := grpczap.Extract(ctx)

	customerOb, err := h.customerService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		if errors.Is(err, customer.ErrNotFound) {
			return nil, grpcCustomerNotFoundErr
		}
		return nil, grpcInternalServerError
	}

	var paymentMethodsPbs []*frontierv1beta1.PaymentMethod
	if request.GetWithPaymentMethods() {
		pms, err := h.customerService.ListPaymentMethods(ctx, request.GetId())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		for _, v := range pms {
			pmPB, err := transformPaymentMethodToPB(v)
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcInternalServerError
			}
			paymentMethodsPbs = append(paymentMethodsPbs, pmPB)
		}
	}

	customerPB, err := transformCustomerToPB(customerOb)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.GetBillingAccountResponse{
		BillingAccount: customerPB,
		PaymentMethods: paymentMethodsPbs,
	}, nil
}

func (h Handler) DeleteBillingAccount(ctx context.Context, request *frontierv1beta1.DeleteBillingAccountRequest) (*frontierv1beta1.DeleteBillingAccountResponse, error) {
	logger := grpczap.Extract(ctx)

	err := h.customerService.Delete(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.DeleteBillingAccountResponse{}, nil
}

func (h Handler) GetBillingBalance(ctx context.Context, request *frontierv1beta1.GetBillingBalanceRequest) (*frontierv1beta1.GetBillingBalanceResponse, error) {
	logger := grpczap.Extract(ctx)

	balanceAmount, err := h.creditService.GetBalance(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.GetBillingBalanceResponse{
		Balance: &frontierv1beta1.BillingAccount_Balance{
			Amount:    balanceAmount,
			Currency:  "VC",
			UpdatedAt: nil,
		},
	}, nil
}

func (h Handler) UpdateBillingAccount(ctx context.Context, request *frontierv1beta1.UpdateBillingAccountRequest) (*frontierv1beta1.UpdateBillingAccountResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())
	updatedCustomer, err := h.customerService.Update(ctx, customer.Customer{
		ID:       request.GetId(),
		OrgID:    request.GetOrgId(),
		Name:     request.GetBody().GetName(),
		Email:    request.GetBody().GetEmail(),
		Phone:    request.GetBody().GetPhone(),
		Currency: request.GetBody().GetCurrency(),
		Address: customer.Address{
			City:       request.GetBody().GetAddress().GetCity(),
			Country:    request.GetBody().GetAddress().GetCountry(),
			Line1:      request.GetBody().GetAddress().GetLine1(),
			Line2:      request.GetBody().GetAddress().GetLine2(),
			PostalCode: request.GetBody().GetAddress().GetPostalCode(),
			State:      request.GetBody().GetAddress().GetState(),
		},
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	customerPB, err := transformCustomerToPB(updatedCustomer)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.UpdateBillingAccountResponse{
		BillingAccount: customerPB,
	}, nil
}

func transformCustomerToPB(customer customer.Customer) (*frontierv1beta1.BillingAccount, error) {
	metaData, err := customer.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.BillingAccount{}, err
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
		State:     customer.State,
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
		Metadata:  metaData,
	}, nil
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
