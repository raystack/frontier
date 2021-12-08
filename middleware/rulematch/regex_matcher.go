package rulematch

import (
	"context"
	"net/url"

	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
)

type RegexMatcher struct {
	ruleRepo store.RuleRepository
}

func (m RegexMatcher) Match(ctx context.Context, reqMethod string, reqURL *url.URL) (*structs.Rule, error) {
	ruleset, err := m.ruleRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, set := range ruleset {
		for _, rule := range set.Rules {
			if reqMethod != rule.Frontend.Method {
				continue
			}

			if rule.Frontend.URLRx.MatchString(reqURL.String()) {
				return &rule, nil
			}
		}
	}
	return nil, ErrUnknownRule
}

func NewRegexMatcher(ruleRepo store.RuleRepository) *RegexMatcher {
	return &RegexMatcher{
		ruleRepo: ruleRepo,
	}
}
