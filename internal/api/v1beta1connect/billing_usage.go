package v1beta1connect

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/usage"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UsageService interface {
	Report(ctx context.Context, usages []usage.Usage) error
	Revert(ctx context.Context, accountID, usageID string, amount int64) error
}

func (h *ConnectHandler) CreateBillingUsage(ctx context.Context, request *connect.Request[frontierv1beta1.CreateBillingUsageRequest]) (*connect.Response[frontierv1beta1.CreateBillingUsageResponse], error) {
	createRequests := make([]usage.Usage, 0, len(request.Msg.GetUsages()))
	for _, v := range request.Msg.GetUsages() {
		usageType := usage.CreditType
		if len(v.GetType()) > 0 {
			usageType = usage.Type(v.GetType())
		}

		createRequests = append(createRequests, usage.Usage{
			ID:          v.GetId(),
			CustomerID:  request.Msg.GetBillingId(),
			Type:        usageType,
			Amount:      v.GetAmount(),
			Source:      strings.ToLower(v.GetSource()), // source in lower case looks nicer
			Description: v.GetDescription(),
			UserID:      v.GetUserId(),
			Metadata:    v.GetMetadata().AsMap(),
		})
	}

	if err := h.usageService.Report(ctx, createRequests); err != nil {
		if errors.Is(err, credit.ErrInsufficientCredits) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if errors.Is(err, credit.ErrAlreadyApplied) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}), nil
}

func (h *ConnectHandler) ListBillingTransactions(ctx context.Context, request *connect.Request[frontierv1beta1.ListBillingTransactionsRequest]) (*connect.Response[frontierv1beta1.ListBillingTransactionsResponse], error) {
	if request.Msg.GetBillingId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	var transactions []*frontierv1beta1.BillingTransaction
	var startRange time.Time
	if request.Msg.GetSince() != nil {
		startRange = request.Msg.GetSince().AsTime()
	}
	if request.Msg.GetStartRange() != nil {
		startRange = request.Msg.GetStartRange().AsTime()
	}
	var endRange time.Time
	if request.Msg.GetEndRange() != nil {
		endRange = request.Msg.GetEndRange().AsTime()
	}

	transactionsList, err := h.creditService.List(ctx, credit.Filter{
		CustomerID: request.Msg.GetBillingId(),
		StartRange: startRange,
		EndRange:   endRange,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range transactionsList {
		transactionPB, err := transformTransactionToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		transactions = append(transactions, transactionPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
		Transactions: transactions,
	}), nil
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

func (h *ConnectHandler) RevertBillingUsage(ctx context.Context, request *connect.Request[frontierv1beta1.RevertBillingUsageRequest]) (*connect.Response[frontierv1beta1.RevertBillingUsageResponse], error) {
	if err := h.usageService.Revert(ctx, request.Msg.GetBillingId(),
		request.Msg.GetUsageId(), request.Msg.GetAmount()); err != nil {
		if errors.Is(err, usage.ErrRevertAmountExceeds) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		} else if errors.Is(err, usage.ErrExistingRevertedUsage) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		} else if errors.Is(err, credit.ErrNotFound) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		} else if errors.Is(err, credit.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		} else if errors.Is(err, credit.ErrAlreadyApplied) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.RevertBillingUsageResponse{}), nil
}
