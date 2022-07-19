package rule

import "context"

type Service struct {
	ruleRepo RuleRepository
}

func NewService(ruleRepo RuleRepository) *Service {
	return &Service{
		ruleRepo: ruleRepo,
	}
}

func (s Service) GetAll(ctx context.Context) ([]Ruleset, error) {
	return s.ruleRepo.GetAll(ctx)
}
