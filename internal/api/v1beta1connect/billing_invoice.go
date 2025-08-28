package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/pkg/pagination"
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
