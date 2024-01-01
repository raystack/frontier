package v1beta1

import (
	"context"

	"github.com/raystack/frontier/billing/invoice"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type InvoiceService interface {
	List(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
}

func (h Handler) ListInvoices(ctx context.Context, request *frontierv1beta1.ListInvoicesRequest) (*frontierv1beta1.ListInvoicesResponse, error) {
	logger := grpczap.Extract(ctx)

	invoices, err := h.invoiceService.List(ctx, invoice.Filter{
		CustomerID: request.GetBillingId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	var invoicePBs []*frontierv1beta1.Invoice
	for _, v := range invoices {
		invoicePB, err := transformInvoiceToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		invoicePBs = append(invoicePBs, invoicePB)
	}

	return &frontierv1beta1.ListInvoicesResponse{
		Invoices: invoicePBs,
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
	if !i.DueDate.IsZero() {
		pb.DueDate = timestamppb.New(i.DueDate)
	}
	if !i.EffectiveAt.IsZero() {
		pb.EffectiveAt = timestamppb.New(i.EffectiveAt)
	}
	if !i.CreatedAt.IsZero() {
		pb.CreatedAt = timestamppb.New(i.CreatedAt)
	}
	return pb, nil
}
