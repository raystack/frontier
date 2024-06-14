package credit

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

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
	if cred.ID == "" {
		return fmt.Errorf("credit id is empty, it is required to create a transaction")
	}
	if cred.Amount < 0 {
		return fmt.Errorf("credit amount is negative")
	}
	// check if already credited
	t, err := s.transactionRepository.GetByID(ctx, cred.ID)

	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if err == nil && t.ID != "" {
		return ErrAlreadyApplied
	}

	txSource := "system"
	if cred.Source != "" {
		txSource = cred.Source
	}

	_, err = s.transactionRepository.CreateEntry(ctx, Transaction{
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
		return fmt.Errorf("transactionRepository.CreateEntry: %w", err)
	}
	return nil
}

func (s Service) Deduct(ctx context.Context, cred Credit) error {
	if cred.ID == "" {
		return fmt.Errorf("credit id is empty, it is required to create a transaction")
	}
	if cred.Amount < 0 {
		return fmt.Errorf("credit amount is negative")
	}

	// check balance, if enough, sub credits
	currentBalance, err := s.GetBalance(ctx, cred.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to apply transaction: %w", err)
	}
	// TODO(kushsharma): this is prone to timing attacks and better we do this
	// in a transaction
	if currentBalance < cred.Amount {
		return ErrInsufficientCredits
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

func (s Service) GetByID(ctx context.Context, id string) (Transaction, error) {
	return s.transactionRepository.GetByID(ctx, id)
}
