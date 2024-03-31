package v1beta1

import (
	"context"
	"errors"
	"strings"

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
	Revert(ctx context.Context, accountID, usageID string, amount int64) error
}

func (h Handler) CreateBillingUsage(ctx context.Context, request *frontierv1beta1.CreateBillingUsageRequest) (*frontierv1beta1.CreateBillingUsageResponse, error) {
	logger := grpczap.Extract(ctx)

	createRequests := make([]usage.Usage, 0, len(request.GetUsages()))
	for _, v := range request.GetUsages() {
		usageType := usage.CreditType
		if len(v.GetType()) > 0 {
			usageType = usage.Type(v.GetType())
		}

		createRequests = append(createRequests, usage.Usage{
			ID:          v.GetId(),
			CustomerID:  request.GetBillingId(),
			Type:        usageType,
			Amount:      v.GetAmount(),
			Source:      strings.ToLower(v.GetSource()), // source in lower case looks nicer
			Description: v.GetDescription(),
			UserID:      v.GetUserId(),
			Metadata:    v.GetMetadata().AsMap(),
		})
	}

	if err := h.usageService.Report(ctx, createRequests); err != nil {
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
		CustomerID: request.GetBillingId(),
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

func (h Handler) RevertBillingUsage(ctx context.Context, request *frontierv1beta1.RevertBillingUsageRequest) (*frontierv1beta1.RevertBillingUsageResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.usageService.Revert(ctx, request.GetBillingId(),
		request.GetUsageId(), request.GetAmount()); err != nil {
		logger.Error(err.Error())
		if errors.Is(err, usage.ErrRevertAmountExceeds) {
			return nil, ErrInvalidInput(err.Error())
		} else if errors.Is(err, usage.ErrExistingRevertedUsage) {
			return nil, ErrInvalidInput(err.Error())
		}
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.RevertBillingUsageResponse{}, nil
}

func transformTransactionToPB(t credit.Transaction) (*frontierv1beta1.BillingTransaction, error) {
	metaData, err := t.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.BillingTransaction{}, err
	}
	return &frontierv1beta1.BillingTransaction{
		Id:          t.ID,
		CustomerId:  t.CustomerID,
		Amount:      t.Amount,
		Type:        string(t.Type),
		Source:      t.Source,
		Description: t.Description,
		UserId:      t.UserID,
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}, nil
}
