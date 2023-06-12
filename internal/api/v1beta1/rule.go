package v1beta1

import (
	"context"

	"github.com/raystack/shield/core/rule"
)

//go:generate mockery --name=RuleService -r --case underscore --with-expecter --structname RuleService --filename rule_service.go --output=./mocks
type RuleService interface {
	GetAllConfigs(ctx context.Context) ([]rule.Ruleset, error)
}
