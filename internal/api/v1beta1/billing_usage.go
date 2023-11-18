package v1beta1

import (
	"context"
	"time"

	"github.com/raystack/frontier/billing/usage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/raystack/frontier/billing/credit"

	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type CreditService interface {
	List(ctx context.Context, filter credit.Filter) ([]credit.Transaction, error)
	GetBalance(ctx context.Context, accountID string) (int64, error)
}

type UsageService interface {
	Report(ctx context.Context, usages []usage.Usage) error
}

func (h Handler) CreateBillingUsage(ctx context.Context, request *frontierv1beta1.CreateBillingUsageRequest) (*frontierv1beta1.CreateBillingUsageResponse, error) {
	logger := grpczap.Extract(ctx)

	createRequests := []usage.Usage{}
	for _, v := range request.GetUsages() {
		createdAt := v.GetCreatedAt().AsTime()
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		createRequests = append(createRequests, usage.Usage{
			ID:          v.GetId(),
			CustomerID:  v.GetCustomerId(),
			Type:        usage.Type(v.GetType()),
			Amount:      v.GetAmount(),
			Source:      v.GetSource(),
			Description: v.GetDescription(),
			Metadata:    v.GetMetadata().AsMap(),
			CreatedAt:   createdAt,
		})
	}
	err := h.usageService.Report(ctx, createRequests)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &frontierv1beta1.CreateBillingUsageResponse{}, nil
}

func (h Handler) ListBillingTransactions(ctx context.Context, request *frontierv1beta1.ListBillingTransactionsRequest) (*frontierv1beta1.ListBillingTransactionsResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}
	var transactions []*frontierv1beta1.BillingTransaction
	transactionsList, err := h.creditService.List(ctx, credit.Filter{
		AccountID: request.GetBillingId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range transactionsList {
		transactionPB, err := transformTransactionToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		transactions = append(transactions, transactionPB)
	}

	return &frontierv1beta1.ListBillingTransactionsResponse{
		Transactions: transactions,
	}, nil
}

func transformTransactionToPB(t credit.Transaction) (*frontierv1beta1.BillingTransaction, error) {
	metaData, err := t.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.BillingTransaction{}, err
	}
	return &frontierv1beta1.BillingTransaction{
		Id:          t.ID,
		CustomerId:  t.AccountID,
		Amount:      t.Amount,
		Type:        string(t.Type),
		Source:      t.Source,
		Description: t.Description,
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}, nil
}
