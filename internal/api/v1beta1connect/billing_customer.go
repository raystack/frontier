package v1beta1connect

import (
	"context"

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
