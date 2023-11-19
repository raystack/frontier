package credit

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/billing/usage"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type TransactionRepository interface {
	CreateEntry(ctx context.Context, debit, credit Transaction) ([]Transaction, error)
	GetBalance(ctx context.Context, id string) (int64, error)
	List(ctx context.Context, flt Filter) ([]Transaction, error)
	GetByID(ctx context.Context, id string) (Transaction, error)
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
	// check if already credited
	if t, err := s.transactionRepository.GetByID(ctx, cred.ID); err == nil && t.ID != "" {
		return ErrAlreadyApplied
	}

	if cred.ID == "" {
		return fmt.Errorf("credit id is empty, it is required to create a transaction")
	}
	description := cred.Description
	if description == "" {
		description = "addition of credits"
	}
	_, err := s.transactionRepository.CreateEntry(ctx, Transaction{
		ID:          cred.ID,
		AccountID:   schema.PlatformOrgID.String(),
		Type:        TypeDebit,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      "system",
		Metadata:    cred.Metadata,
	}, Transaction{
		Type:        TypeCredit,
		AccountID:   cred.AccountID,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      "system",
		Metadata:    cred.Metadata,
	})
	if err != nil {
		return fmt.Errorf("transactionRepository.CreateEntry: %w", err)
	}
	return nil
}

func (s Service) Deduct(ctx context.Context, u usage.Usage) error {
	if u.ID == "" {
		return fmt.Errorf("usage id is empty, it is required to create a transaction")
	}
	if u.Type != usage.TypeCredit {
		return fmt.Errorf("usage is not of credit type")
	}

	// check balance, if enough, sub credits
	currentBalance, err := s.GetBalance(ctx, u.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to apply transaction: %w", err)
	}
	if currentBalance < u.Amount {
		return ErrNotEnough
	}

	description := u.Description
	if description == "" {
		description = "utilization of credits"
	}

	if _, err := s.transactionRepository.CreateEntry(ctx, Transaction{
		ID:          u.ID,
		AccountID:   u.CustomerID,
		Type:        TypeDebit,
		Amount:      u.Amount,
		Description: u.Description,
		Source:      u.Source,
		Metadata:    u.Metadata,
	}, Transaction{
		Type:        TypeCredit,
		AccountID:   schema.PlatformOrgID.String(),
		Amount:      u.Amount,
		Description: u.Description,
		Source:      u.Source,
		Metadata:    u.Metadata,
	}); err != nil {
		return fmt.Errorf("failed to sub credits: %w", err)
	}
	return nil
}

func (s Service) List(ctx context.Context, flt Filter) ([]Transaction, error) {
	return s.transactionRepository.List(ctx, flt)
}

func (s Service) GetBalance(ctx context.Context, accountID string) (int64, error) {
	return s.transactionRepository.GetBalance(ctx, accountID)
}
