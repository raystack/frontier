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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UsageService interface {
	Report(ctx context.Context, usages []usage.Usage) error
	Revert(ctx context.Context, accountID, usageID string, amount int64) error
}

func (h *ConnectHandler) CreateBillingUsage(ctx context.Context, request *connect.Request[frontierv1beta1.CreateBillingUsageRequest]) (*connect.Response[frontierv1beta1.CreateBillingUsageResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichCreateBillingUsageRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	createRequests := make([]usage.Usage, 0, len(enrichedReq.GetUsages()))
	for _, v := range enrichedReq.GetUsages() {
		usageType := usage.CreditType
		if len(v.GetType()) > 0 {
			usageType = usage.Type(v.GetType())
		}

		createRequests = append(createRequests, usage.Usage{
			ID:          v.GetId(),
			CustomerID:  enrichedReq.GetBillingId(),
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateBillingUsageResponse{}), nil
}

func (h *ConnectHandler) ListBillingTransactions(ctx context.Context, request *connect.Request[frontierv1beta1.ListBillingTransactionsRequest]) (*connect.Response[frontierv1beta1.ListBillingTransactionsResponse], error) {
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichListBillingTransactionsRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	var transactions []*frontierv1beta1.BillingTransaction
	var startRange time.Time
	if enrichedReq.GetSince() != nil {
		startRange = enrichedReq.GetSince().AsTime()
	}
	if enrichedReq.GetStartRange() != nil {
		startRange = enrichedReq.GetStartRange().AsTime()
	}
	var endRange time.Time
	if enrichedReq.GetEndRange() != nil {
		endRange = enrichedReq.GetEndRange().AsTime()
	}

	transactionsList, err := h.creditService.List(ctx, credit.Filter{
		CustomerID: enrichedReq.GetBillingId(),
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

	response := &frontierv1beta1.ListBillingTransactionsResponse{
		Transactions: transactions,
	}

	// Handle response enrichment based on expand field
	response = h.enrichListBillingTransactionsResponse(ctx, enrichedReq, response)

	return connect.NewResponse(response), nil
}

func (h *ConnectHandler) TotalDebitedTransactions(ctx context.Context, request *connect.Request[frontierv1beta1.TotalDebitedTransactionsRequest]) (*connect.Response[frontierv1beta1.TotalDebitedTransactionsResponse], error) {
	if request.Msg.GetBillingId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	debitAmount, err := h.creditService.GetTotalDebitedAmount(ctx, request.Msg.GetBillingId())
	if err != nil {
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
	// Handle request enrichment similar to gRPC interceptor
	enrichedReq, err := h.enrichRevertBillingUsageRequest(ctx, request.Msg)
	if err != nil {
		return nil, err
	}

	if err := h.usageService.Revert(ctx, enrichedReq.GetBillingId(),
		enrichedReq.GetUsageId(), enrichedReq.GetAmount()); err != nil {
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

// enrichCreateBillingUsageRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichCreateBillingUsageRequest(ctx context.Context, req *frontierv1beta1.CreateBillingUsageRequest) (*frontierv1beta1.CreateBillingUsageRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.CreateBillingUsageRequest{
		ProjectId: req.GetProjectId(),
		OrgId:     req.GetOrgId(),
		BillingId: req.GetBillingId(),
		Usages:    req.GetUsages(),
	}

	// Step 1: Convert project ID to org ID if needed
	if enrichedReq.GetProjectId() != "" && enrichedReq.GetOrgId() != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrStatusOrgProjectMismatch)
	}

	if enrichedReq.GetProjectId() != "" {
		proj, err := h.GetProject(ctx, connect.NewRequest(&frontierv1beta1.GetProjectRequest{
			Id: enrichedReq.GetProjectId(),
		}))
		if err != nil {
			return nil, err
		}
		if proj != nil && proj.Msg != nil && proj.Msg.GetProject() != nil {
			enrichedReq.OrgId = proj.Msg.GetProject().GetOrgId()
		}
	}

	// Step 2: Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}

// enrichRevertBillingUsageRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichRevertBillingUsageRequest(ctx context.Context, req *frontierv1beta1.RevertBillingUsageRequest) (*frontierv1beta1.RevertBillingUsageRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.RevertBillingUsageRequest{
		ProjectId: req.GetProjectId(),
		OrgId:     req.GetOrgId(),
		BillingId: req.GetBillingId(),
		UsageId:   req.GetUsageId(),
		Amount:    req.GetAmount(),
	}

	// Step 1: Convert project ID to org ID if needed
	if enrichedReq.GetProjectId() != "" && enrichedReq.GetOrgId() != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrStatusOrgProjectMismatch)
	}

	if enrichedReq.GetProjectId() != "" {
		proj, err := h.GetProject(ctx, connect.NewRequest(&frontierv1beta1.GetProjectRequest{
			Id: enrichedReq.GetProjectId(),
		}))
		if err != nil {
			return nil, err
		}
		if proj != nil && proj.Msg != nil && proj.Msg.GetProject() != nil {
			enrichedReq.OrgId = proj.Msg.GetProject().GetOrgId()
		}
	}

	// Step 2: Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		// Find default customer id for the org
		customers, err := h.customerService.List(ctx, customer.Filter{
			OrgID: enrichedReq.GetOrgId(),
			State: customer.ActiveState,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		if len(customers) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrCustomerNotFound)
		}
		enrichedReq.BillingId = customers[0].ID
	}

	return enrichedReq, nil
}

// findDefaultBillingAccount finds the default billing account for an organization
func (h *ConnectHandler) findDefaultBillingAccount(ctx context.Context, orgID string) (string, error) {
	if orgID == "" {
		return "", connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	customers, err := h.customerService.List(ctx, customer.Filter{
		OrgID: orgID,
		State: customer.ActiveState,
	})
	if err != nil {
		return "", connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	if len(customers) == 0 {
		return "", connect.NewError(connect.CodeInvalidArgument, ErrCustomerNotFound)
	}
	return customers[0].ID, nil
}

// enrichListBillingTransactionsRequest enriches the request similar to gRPC interceptor
func (h *ConnectHandler) enrichListBillingTransactionsRequest(ctx context.Context, req *frontierv1beta1.ListBillingTransactionsRequest) (*frontierv1beta1.ListBillingTransactionsRequest, error) {
	// Create a copy of the request to avoid modifying the original
	enrichedReq := &frontierv1beta1.ListBillingTransactionsRequest{
		BillingId:  req.GetBillingId(),
		OrgId:      req.GetOrgId(),
		Since:      req.GetSince(),
		StartRange: req.GetStartRange(),
		EndRange:   req.GetEndRange(),
		Expand:     req.GetExpand(),
	}

	// Find default billing account if billing_id is empty
	if enrichedReq.GetBillingId() == "" && enrichedReq.GetOrgId() != "" {
		billingID, err := h.findDefaultBillingAccount(ctx, enrichedReq.GetOrgId())
		if err != nil {
			return nil, err
		}
		enrichedReq.BillingId = billingID
	}

	return enrichedReq, nil
}
