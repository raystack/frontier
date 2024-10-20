package usage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/billing/credit"
)

type CreditService interface {
	Add(ctx context.Context, cred credit.Credit) error
	Deduct(ctx context.Context, cred credit.Credit) error
	GetByID(ctx context.Context, id string) (credit.Transaction, error)
}

type Service struct {
	creditService CreditService
}

func NewService(transactionService CreditService) *Service {
	return &Service{
		creditService: transactionService,
	}
}

func (s Service) Report(ctx context.Context, usages []Usage) error {
	var errs []error
	for _, u := range usages {
		switch u.Type {
		case CreditType:
			if err := s.creditService.Deduct(ctx, credit.Credit{
				ID:          u.ID,
				CustomerID:  u.CustomerID,
				Amount:      u.Amount,
				UserID:      u.UserID,
				Source:      u.Source,
				Description: u.Description,
				Metadata:    u.Metadata,
			}); err != nil {
				errs = append(errs, fmt.Errorf("failed to deduct usage: %w", err))
			}
		default:
			errs = append(errs, fmt.Errorf("unsupported usage type: %s for usage %s", u.Type, u.ID))
		}
	}
	return errors.Join(errs...)
}

func (s Service) Revert(ctx context.Context, customerID, usageID string, amount int64) error {
	creditTx, err := s.creditService.GetByID(ctx, usageID)
	if err != nil {
		return fmt.Errorf("creditService.GetByID: %w", err)
	}
	if creditTx.CustomerID != customerID {
		return fmt.Errorf("creditService.GetByID: accountID mismatch")
	}
	// check amount
	if amount > creditTx.Amount {
		return ErrRevertAmountExceeds
	}
	// a revert can't be reverted
	if strings.HasPrefix(creditTx.Source, credit.SourceSystemRevertEvent) {
		return ErrExistingRevertedUsage
	}
	revertMeta := creditTx.Metadata
	revertMeta["revert_request_using"] = creditTx.ID

	// Revert the usage
	if err := s.creditService.Add(ctx, credit.Credit{
		ID:          credit.TxUUID(usageID, customerID),
		CustomerID:  customerID,
		Amount:      amount,
		Description: fmt.Sprintf("Revert: %s", creditTx.Description),
		Source:      fmt.Sprintf("%s.%s", credit.SourceSystemRevertEvent, creditTx.Source),
		UserID:      creditTx.UserID,
		Metadata:    revertMeta,
	}); err != nil {
		return fmt.Errorf("creditService.Add: %w", err)
	}
	return nil
}
