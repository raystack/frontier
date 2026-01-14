package v1beta1connect

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/usage"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateBillingUsage(ctx context.Context, request *connect.Request[frontierv1beta1.CreateBillingUsageRequest]) (*connect.Response[frontierv1beta1.CreateBillingUsageResponse], error) {
	errorLogger := NewErrorLogger()

	// Always infer billing_id from org_id
	cust, err := h.customerService.GetByOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		errorLogger.LogServiceError(ctx, request, "CreateBillingUsage.GetByOrgID", err,
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	createRequests := make([]usage.Usage, 0, len(request.Msg.GetUsages()))
	for _, v := range request.Msg.GetUsages() {
		usageType := usage.CreditType
		if len(v.GetType()) > 0 {
			usageType = usage.Type(v.GetType())
		}

		createRequests = append(createRequests, usage.Usage{
			ID:          v.GetId(),
			CustomerID:  cust.ID,
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
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInsufficientCredits)
		}
		if errors.Is(err, credit.ErrAlreadyApplied) {
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrAlreadyApplied)
		}
		errorLogger.LogServiceError(ctx, request, "CreateBillingUsage.Report", err,
			zap.String("billing_id", cust.ID),
			zap.String("org_id", request.Msg.GetOrgId()),
			zap.Int("usage_count", len(createRequests)))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}), nil
}

func (h *ConnectHandler) ListBillingTransactions(ctx context.Context, request *connect.Request[frontierv1beta1.ListBillingTransactionsRequest]) (*connect.Response[frontierv1beta1.ListBillingTransactionsResponse], error) {
	errorLogger := NewErrorLogger()

	// Always infer billing_id from org_id
	cust, err := h.customerService.GetByOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		// Return empty list if billing account doesn't exist
		if errors.Is(err, customer.ErrNotFound) {
			return connect.NewResponse(&frontierv1beta1.ListBillingTransactionsResponse{
				Transactions: []*frontierv1beta1.BillingTransaction{},
			}), nil
		}
		// Return bad request for invalid org_id
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		errorLogger.LogServiceError(ctx, request, "ListBillingTransactions.GetByOrgID", err,
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	billingID := cust.ID

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
		CustomerID: billingID,
		StartRange: startRange,
		EndRange:   endRange,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListBillingTransactions.List", err,
			zap.String("org_id", request.Msg.GetOrgId()),
			zap.String("billing_id", billingID),
			zap.Time("start_range", startRange),
			zap.Time("end_range", endRange))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range transactionsList {
		transactionPB, err := transformTransactionToPB(v)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListBillingTransactions", v.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		transactions = append(transactions, transactionPB)
	}

	response := &frontierv1beta1.ListBillingTransactionsResponse{
		Transactions: transactions,
	}

	// Handle response enrichment based on expand field
	response = h.enrichListBillingTransactionsResponse(ctx, request.Msg, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) TotalDebitedTransactions(ctx context.Context, request *connect.Request[frontierv1beta1.TotalDebitedTransactionsRequest]) (*connect.Response[frontierv1beta1.TotalDebitedTransactionsResponse], error) {
	errorLogger := NewErrorLogger()

	// Always infer billing_id from org_id
	cust, err := h.customerService.GetByOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		// Return zero amount if billing account doesn't exist
		if errors.Is(err, customer.ErrNotFound) {
			return connect.NewResponse(&frontierv1beta1.TotalDebitedTransactionsResponse{
				Debited: &frontierv1beta1.BillingAccount_Balance{
					Amount:   0,
					Currency: "VC",
				},
			}), nil
		}
		// Return bad request for invalid org_id
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		errorLogger.LogServiceError(ctx, request, "TotalDebitedTransactions.GetByOrgID", err,
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	billingID := cust.ID

	debitAmount, err := h.creditService.GetTotalDebitedAmount(ctx, billingID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "TotalDebitedTransactions.GetTotalDebitedAmount", err,
			zap.String("org_id", request.Msg.GetOrgId()),
			zap.String("billing_id", billingID))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.TotalDebitedTransactionsResponse{
		Debited: &frontierv1beta1.BillingAccount_Balance{
			Amount:   debitAmount,
			Currency: "VC",
		},
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
	errorLogger := NewErrorLogger()

	// Always infer billing_id from org_id
	cust, err := h.customerService.GetByOrgID(ctx, request.Msg.GetOrgId())
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		errorLogger.LogServiceError(ctx, request, "RevertBillingUsage.GetByOrgID", err,
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if err := h.usageService.Revert(ctx, cust.ID,
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
		errorLogger.LogServiceError(ctx, request, "RevertBillingUsage.Revert", err,
			zap.String("billing_id", cust.ID),
			zap.String("org_id", request.Msg.GetOrgId()),
			zap.String("usage_id", request.Msg.GetUsageId()),
			zap.Int64("amount", request.Msg.GetAmount()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.RevertBillingUsageResponse{}), nil
}

// IsUserIDSuperUser returns true if the user ID is a super user
func (h *ConnectHandler) IsUserIDSuperUser(ctx context.Context, userID string) (bool, error) {
	return h.userService.IsSudo(ctx, userID, schema.PlatformSudoPermission)
}

// parseExpandModels extracts expand field values from any request using reflection
func parseExpandModels(req any) map[string]bool {
	expandModels := map[string]bool{}
	expandReflect := reflect.ValueOf(req).Elem().FieldByName("Expand")
	if expandReflect.IsValid() && expandReflect.Len() > 0 {
		for i := 0; i < expandReflect.Len(); i++ {
			expandModels[strings.ToLower(expandReflect.Index(i).String())] = true
		}
	}
	return expandModels
}

// enrichListBillingTransactionsResponse enriches the response with expanded fields
func (h *ConnectHandler) enrichListBillingTransactionsResponse(ctx context.Context, req *frontierv1beta1.ListBillingTransactionsRequest, resp *frontierv1beta1.ListBillingTransactionsResponse) *frontierv1beta1.ListBillingTransactionsResponse {
	expandModels := parseExpandModels(req)
	if len(expandModels) == 0 {
		// no need to enrich the response
		return resp
	}

	if len(resp.GetTransactions()) > 0 {
		for tIdx, t := range resp.GetTransactions() {
			if expandModels["customer"] && len(t.GetCustomerId()) > 0 {
				ba, _ := h.GetBillingAccount(ctx, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
					Id: t.GetCustomerId(),
				}))
				if ba != nil && ba.Msg != nil {
					resp.Transactions[tIdx].Customer = ba.Msg.GetBillingAccount()
				}
			}

			if expandModels["user"] && len(t.GetUserId()) > 0 {
				// if we allowed anyone to report usage with a user id, a bad actor can pass any user id
				// and retrieve user information.
				user, _ := h.GetUser(ctx, connect.NewRequest(&frontierv1beta1.GetUserRequest{
					Id: t.GetUserId(),
				}))
				if user != nil && user.Msg != nil {
					if isSuper, err := h.IsUserIDSuperUser(ctx, user.Msg.GetUser().GetId()); err == nil {
						if !isSuper {
							resp.Transactions[tIdx].User = user.Msg.GetUser()
						}
					}
				}
			}
		}
	}

	return resp
}
