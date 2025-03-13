package credit

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type TransactionRepository interface {
	CreateEntry(ctx context.Context, debit, credit Transaction) ([]Transaction, error)
	GetBalance(ctx context.Context, id string) (int64, error)
	GetTotalDebitedAmount(ctx context.Context, id string) (int64, error)
	List(ctx context.Context, flt Filter) ([]Transaction, error)
	GetByID(ctx context.Context, id string) (Transaction, error)
	GetBalanceForRange(ctx context.Context, accountID string, start time.Time, end time.Time) (int64, error)
}

type Service struct {
	transactionRepository TransactionRepository
}

func NewService(repository TransactionRepository) *Service {
	return &Service{
		transactionRepository: repository,
	}
}

func (s Service) Add(ctx context.Context, cred Credit) error {
	if cred.ID == "" {
		return errors.New("credit id is empty, it is required to create a transaction")
	}
	if cred.Amount < 0 {
		return errors.New("credit amount is negative")
	}

	txSource := "system"
	if cred.Source != "" {
		txSource = cred.Source
	}

	_, err := s.transactionRepository.CreateEntry(ctx, Transaction{
		CustomerID:  schema.PlatformOrgID.String(),
		Type:        DebitType,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}, Transaction{
		ID:          cred.ID,
		Type:        CreditType,
		CustomerID:  cred.CustomerID,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	})
	if err != nil {
		if errors.Is(err, ErrAlreadyApplied) {
			return ErrAlreadyApplied
		}
		return fmt.Errorf("transactionRepository.CreateEntry: %w", err)
	}
	return nil
}

func (s Service) Deduct(ctx context.Context, cred Credit) error {
	if cred.ID == "" {
		return errors.New("credit id is empty, it is required to create a transaction")
	}
	if cred.Amount < 0 {
		return errors.New("credit amount is negative")
	}

	txSource := "system"
	if cred.Source != "" {
		txSource = cred.Source
	}

	if _, err := s.transactionRepository.CreateEntry(ctx, Transaction{
		ID:          cred.ID,
		CustomerID:  cred.CustomerID,
		Type:        DebitType,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}, Transaction{
		Type:        CreditType,
		CustomerID:  schema.PlatformOrgID.String(),
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}); err != nil {
		if errors.Is(err, ErrAlreadyApplied) {
			return ErrAlreadyApplied
		} else if errors.Is(err, ErrInsufficientCredits) {
			return ErrInsufficientCredits
		}
		return fmt.Errorf("failed to deduct credits: %w", err)
	}
	return nil
}

func (s Service) List(ctx context.Context, flt Filter) ([]Transaction, error) {
	return s.transactionRepository.List(ctx, flt)
}

func (s Service) GetBalance(ctx context.Context, accountID string) (int64, error) {
	return s.transactionRepository.GetBalance(ctx, accountID)
}

func (s Service) GetTotalDebitedAmount(ctx context.Context, accountID string) (int64, error) {
	return s.transactionRepository.GetTotalDebitedAmount(ctx, accountID)
}

// GetBalanceForRange returns the balance for the given accountID within the given time range
// start time is inclusive, end time is exclusive
func (s Service) GetBalanceForRange(ctx context.Context, accountID string, start time.Time, end time.Time) (int64, error) {
	return s.transactionRepository.GetBalanceForRange(ctx, accountID, start, end)
}

func (s Service) GetByID(ctx context.Context, id string) (Transaction, error) {
	return s.transactionRepository.GetByID(ctx, id)
}
