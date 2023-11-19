package usage

import (
	"context"
	"errors"
	"fmt"
)

type CreditService interface {
	Deduct(ctx context.Context, u Usage) error
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
		case TypeCredit:
			if err := s.creditService.Deduct(ctx, u); err != nil {
				errs = append(errs, fmt.Errorf("failed to deduct usage: %w", err))
			}
		default:
			errs = append(errs, fmt.Errorf("unsupported usage type: %s for usage %s", u.Type, u.ID))
		}
	}
	return errors.Join(errs...)
}
