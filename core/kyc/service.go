package kyc

import "context"

type Repository interface {
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s Service) GetKyc(ctx context.Context, idOrSlug string) (KYC, error) {
	return KYC{OrgId: "blah",
		Status: true,
		Link:   "abcd"}, nil
}
func (s Service) SetKyc(ctx context.Context, idOrSlug string) (KYC, error) {

}
