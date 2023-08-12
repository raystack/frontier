package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/rule"
)

type RuleService interface {
	GetAllConfigs(ctx context.Context) ([]rule.Ruleset, error)
}
