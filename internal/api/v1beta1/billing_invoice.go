package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/billing/invoice"
	"google.golang.org/protobuf/types/known/timestamppb"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type InvoiceService interface {
	List(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	ListAll(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	GetUpcoming(ctx context.Context, customerID string) (invoice.Invoice, error)
	TriggerCreditOverdraftInvoices(ctx context.Context) error
	SearchInvoices(ctx context.Context, rqlQuery *rql.Query) ([]invoice.InvoiceWithOrganization, error)
}

func (h Handler) ListAllInvoices(ctx context.Context, request *frontierv1beta1.ListAllInvoicesRequest) (*frontierv1beta1.ListAllInvoicesResponse, error) {
	paginate := pagination.NewPagination(request.GetPageNum(), request.GetPageSize())

	invoices, err := h.invoiceService.ListAll(ctx, invoice.Filter{
		Pagination: paginate,
	})
	if err != nil {
		return nil, err
	}
	var invoicePBs []*frontierv1beta1.Invoice
	for _, v := range invoices {
		invoicePB, err := transformInvoiceToPB(v)
		if err != nil {
			return nil, err
		}
		invoicePBs = append(invoicePBs, invoicePB)
	}

	return &frontierv1beta1.ListAllInvoicesResponse{
		Invoices: invoicePBs,
		Count:    paginate.Count,
	}, nil
}

func (h Handler) ListInvoices(ctx context.Context, request *frontierv1beta1.ListInvoicesRequest) (*frontierv1beta1.ListInvoicesResponse, error) {
	invoices, err := h.invoiceService.List(ctx, invoice.Filter{
		CustomerID:  request.GetBillingId(),
		NonZeroOnly: request.GetNonzeroAmountOnly(),
	})
	if err != nil {
		return nil, err
	}
	var invoicePBs []*frontierv1beta1.Invoice
	for _, v := range invoices {
		invoicePB, err := transformInvoiceToPB(v)
		if err != nil {
			return nil, err
		}
		invoicePBs = append(invoicePBs, invoicePB)
	}

	return &frontierv1beta1.ListInvoicesResponse{
		Invoices: invoicePBs,
	}, nil
}

func (h Handler) GetUpcomingInvoice(ctx context.Context, request *frontierv1beta1.GetUpcomingInvoiceRequest) (*frontierv1beta1.GetUpcomingInvoiceResponse, error) {
	invoice, err := h.invoiceService.GetUpcoming(ctx, request.GetBillingId())
	if err != nil {
		return nil, err
	}
	invoicePB, err := transformInvoiceToPB(invoice)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetUpcomingInvoiceResponse{
		Invoice: invoicePB,
	}, nil
}

func (h Handler) GenerateInvoices(ctx context.Context, request *frontierv1beta1.GenerateInvoicesRequest) (*frontierv1beta1.GenerateInvoicesResponse, error) {
	err := h.invoiceService.TriggerCreditOverdraftInvoices(ctx)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GenerateInvoicesResponse{}, nil
}

func (h Handler) SearchInvoices(ctx context.Context, request *frontierv1beta1.SearchInvoicesRequest) (*frontierv1beta1.SearchInvoicesResponse, error) {
	var invoices []*frontierv1beta1.SearchInvoicesResponse_Invoice

	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), invoice.InvoiceWithOrganization{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, invoice.InvoiceWithOrganization{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	invoicesData, err := h.invoiceService.SearchInvoices(ctx, rqlQuery)
	if err != nil {
		if errors.Is(err, invoice.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	for _, v := range invoicesData {
		invoices = append(invoices, transformInvoiceToSearchPB(v))
	}

	return &frontierv1beta1.SearchInvoicesResponse{
		Invoices: invoices,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(rqlQuery.Offset),
			Limit:  uint32(rqlQuery.Limit),
		},
	}, nil
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
