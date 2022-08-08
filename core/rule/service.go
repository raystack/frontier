package rule

import "context"

type Service struct {
	configRepository ConfigRepository
}

func NewService(configRepository ConfigRepository) *Service {
	return &Service{
		configRepository: configRepository,
	}
}

func (s Service) GetAllConfigs(ctx context.Context) ([]Ruleset, error) {
	return s.configRepository.GetAll(ctx)
}
