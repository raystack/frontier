package credit

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
)

type TransactionRepository interface {
	CreateEntry(ctx context.Context, debit, credit Transaction) ([]Transaction, error)
	GetBalance(ctx context.Context, id string) (int64, error)
	GetTotalDebitedAmount(ctx context.Context, id string) (int64, error)
	List(ctx context.Context, flt Filter) ([]Transaction, error)
	GetByID(ctx context.Context, id string) (Transaction, error)
	GetBalanceForRange(ctx context.Context, accountID string, start time.Time, end time.Time) (int64, error)
}

type CustomerRepository interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, record auditrecord.AuditRecord) (auditrecord.AuditRecord, error)
}

type Service struct {
	transactionRepository TransactionRepository
	customerRepository    CustomerRepository
	auditRepository       AuditRecordRepository
}

func NewService(repository TransactionRepository, customerRepo CustomerRepository, auditRepo AuditRecordRepository) *Service {
	return &Service{
		transactionRepository: repository,
		customerRepository:    customerRepo,
		auditRepository:       auditRepo,
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

	debitEntry := Transaction{
		CustomerID:  schema.PlatformOrgID.String(),
		Type:        DebitType,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}
	creditEntry := Transaction{
		ID:          cred.ID,
		Type:        CreditType,
		CustomerID:  cred.CustomerID,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}

	_, err := s.transactionRepository.CreateEntry(ctx, debitEntry, creditEntry)
	if err != nil {
		if errors.Is(err, ErrAlreadyApplied) {
			return ErrAlreadyApplied
		}
		return fmt.Errorf("transactionRepository.CreateEntry: %w", err)
	}

	if creditEntry.CustomerID != schema.PlatformOrgID.String() {
		if err := s.createAuditRecord(ctx, creditEntry.CustomerID, pkgAuditRecord.BillingTransactionCreditEvent, creditEntry.ID, creditEntry); err != nil {
			return err
		}
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

	debitEntry := Transaction{
		ID:          cred.ID,
		CustomerID:  cred.CustomerID,
		Type:        DebitType,
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}
	creditEntry := Transaction{
		Type:        CreditType,
		CustomerID:  schema.PlatformOrgID.String(),
		Amount:      cred.Amount,
		Description: cred.Description,
		Source:      txSource,
		UserID:      cred.UserID,
		Metadata:    cred.Metadata,
	}

	_, err := s.transactionRepository.CreateEntry(ctx, debitEntry, creditEntry)
	if err != nil {
		if errors.Is(err, ErrAlreadyApplied) {
			return ErrAlreadyApplied
		} else if errors.Is(err, ErrInsufficientCredits) {
			return ErrInsufficientCredits
		}
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	// Create audit record after transaction succeeds
	if debitEntry.CustomerID != schema.PlatformOrgID.String() {
		if err := s.createAuditRecord(ctx, debitEntry.CustomerID, pkgAuditRecord.BillingTransactionDebitEvent, debitEntry.ID, debitEntry); err != nil {
			return err
		}
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

// createAuditRecord creates an audit record for billing transaction events.
func (s Service) createAuditRecord(ctx context.Context, customerID string, eventType pkgAuditRecord.Event, txID string, txEntry Transaction) error {
	customerAcc, err := s.customerRepository.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	_, err = s.auditRepository.Create(ctx, auditrecord.AuditRecord{
		Event: eventType,
		Resource: auditrecord.Resource{
			ID:   customerID,
			Type: pkgAuditRecord.BillingCustomerType,
			Name: customerAcc.Name,
		},
		Target: &auditrecord.Target{
			ID:   txID,
			Type: pkgAuditRecord.BillingTransactionType,
			Metadata: map[string]interface{}{
				"amount":      txEntry.Amount,
				"source":      txEntry.Source,
				"description": txEntry.Description,
				"user_id":     txEntry.UserID,
			},
		},
		OccurredAt: time.Now(),
		OrgID:      customerAcc.OrgID,
	})
	return err
}
