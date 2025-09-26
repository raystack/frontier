package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InvoiceService interface {
	List(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	ListAll(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	GetUpcoming(ctx context.Context, customerID string) (invoice.Invoice, error)
	TriggerCreditOverdraftInvoices(ctx context.Context) error
	SearchInvoices(ctx context.Context, rqlQuery *rql.Query) ([]invoice.InvoiceWithOrganization, error)
}

func (h *ConnectHandler) ListAllInvoices(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllInvoicesRequest]) (*connect.Response[frontierv1beta1.ListAllInvoicesResponse], error) {
	paginate := pagination.NewPagination(request.Msg.GetPageNum(), request.Msg.GetPageSize())

	invoices, err := h.invoiceService.ListAll(ctx, invoice.Filter{
		Pagination: paginate,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	var invoicePBs []*frontierv1beta1.Invoice
	for _, v := range invoices {
		invoicePB, err := transformInvoiceToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		invoicePBs = append(invoicePBs, invoicePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListAllInvoicesResponse{
		Invoices: invoicePBs,
		Count:    paginate.Count,
	}), nil
}

func (h *ConnectHandler) ListInvoices(ctx context.Context, request *connect.Request[frontierv1beta1.ListInvoicesRequest]) (*connect.Response[frontierv1beta1.ListInvoicesResponse], error) {
	invoices, err := h.invoiceService.List(ctx, invoice.Filter{
		CustomerID:  request.Msg.GetBillingId(),
		NonZeroOnly: request.Msg.GetNonzeroAmountOnly(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	var invoicePBs []*frontierv1beta1.Invoice
	for _, v := range invoices {
		invoicePB, err := transformInvoiceToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		invoicePBs = append(invoicePBs, invoicePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListInvoicesResponse{
		Invoices: invoicePBs,
	}), nil
}

func transformInvoiceToPB(i invoice.Invoice) (*frontierv1beta1.Invoice, error) {
	metaData, err := i.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Invoice{}, err
	}

	pb := &frontierv1beta1.Invoice{
		Id:         i.ID,
		CustomerId: i.CustomerID,
		ProviderId: i.ProviderID,
		State:      i.State.String(),
		Currency:   i.Currency,
		Amount:     i.Amount,
		HostedUrl:  i.HostedURL,
		Metadata:   metaData,
	}
	if !i.DueAt.IsZero() {
		pb.DueDate = timestamppb.New(i.DueAt)
	}
	if !i.EffectiveAt.IsZero() {
		pb.EffectiveAt = timestamppb.New(i.EffectiveAt)
	}
	if !i.CreatedAt.IsZero() {
		pb.CreatedAt = timestamppb.New(i.CreatedAt)
	}
	if !i.PeriodStartAt.IsZero() {
		pb.PeriodStartAt = timestamppb.New(i.PeriodStartAt)
	}
	if !i.PeriodEndAt.IsZero() {
		pb.PeriodEndAt = timestamppb.New(i.PeriodEndAt)
	}
	return pb, nil
}

func (h *ConnectHandler) GenerateInvoices(ctx context.Context, request *connect.Request[frontierv1beta1.GenerateInvoicesRequest]) (*connect.Response[frontierv1beta1.GenerateInvoicesResponse], error) {
	err := h.invoiceService.TriggerCreditOverdraftInvoices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GenerateInvoicesResponse{}), nil
}

func (h *ConnectHandler) SearchInvoices(ctx context.Context, request *connect.Request[frontierv1beta1.SearchInvoicesRequest]) (*connect.Response[frontierv1beta1.SearchInvoicesResponse], error) {
	var invoices []*frontierv1beta1.SearchInvoicesResponse_Invoice

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), invoice.InvoiceWithOrganization{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, invoice.InvoiceWithOrganization{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	invoicesData, err := h.invoiceService.SearchInvoices(ctx, rqlQuery)
	if err != nil {
		if errors.Is(err, invoice.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range invoicesData {
		invoices = append(invoices, transformInvoiceToSearchPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.SearchInvoicesResponse{
		Invoices: invoices,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(rqlQuery.Offset),
			Limit:  uint32(rqlQuery.Limit),
		},
	}), nil
}

func transformInvoiceToSearchPB(v invoice.InvoiceWithOrganization) *frontierv1beta1.SearchInvoicesResponse_Invoice {
	return &frontierv1beta1.SearchInvoicesResponse_Invoice{
		Id:          v.ID,
		Amount:      v.Amount,
		Currency:    v.Currency,
		State:       v.State.String(),
		InvoiceLink: v.InvoiceLink,
		CreatedAt:   timestamppb.New(v.CreatedAt),
		OrgId:       v.OrgID,
		OrgName:     v.OrgName,
		OrgTitle:    v.OrgTitle,
	}
}
