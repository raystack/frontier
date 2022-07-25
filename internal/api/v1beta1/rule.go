package v1beta1

import (
	"context"

	"github.com/odpf/shield/core/rule"
)

type RuleService interface {
	GetAllConfigs(ctx context.Context) ([]rule.Ruleset, error)
}
