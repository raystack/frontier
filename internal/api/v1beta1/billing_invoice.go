package v1beta1

import (
	"context"
	"github.com/raystack/frontier/pkg/pagination"

	"github.com/raystack/frontier/billing/invoice"
	"google.golang.org/protobuf/types/known/timestamppb"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type InvoiceService interface {
	List(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	ListAll(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	GetUpcoming(ctx context.Context, customerID string) (invoice.Invoice, error)
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

func transformInvoiceToPB(i invoice.Invoice) (*frontierv1beta1.Invoice, error) {
	metaData, err := i.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Invoice{}, err
	}

	pb := &frontierv1beta1.Invoice{
		Id:         i.ID,
		CustomerId: i.CustomerID,
		ProviderId: i.ProviderID,
		State:      i.State,
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
