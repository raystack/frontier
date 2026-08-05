package kyc

import "context"

type Repository interface {
	GetByOrgID(context.Context, string) (KYC, error)
	List(context.Context) ([]KYC, error)
	Upsert(context.Context, KYC) (KYC, error)
	Delete(context.Context, string) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s Service) GetKyc(ctx context.Context, orgID string) (KYC, error) {
	return s.repository.GetByOrgID(ctx, orgID)
}

func (s Service) SetKyc(ctx context.Context, kyc KYC) (KYC, error) {
	return s.repository.Upsert(ctx, kyc)
}

func (s Service) ListKycs(ctx context.Context) ([]KYC, error) {
	return s.repository.List(ctx)
}

func (s Service) DeleteKyc(ctx context.Context, orgID string) error {
	return s.repository.Delete(ctx, orgID)
}
