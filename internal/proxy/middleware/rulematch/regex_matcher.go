package rulematch

import (
	"context"
	"net/url"

	"github.com/odpf/shield/core/rule"
)

type RegexMatcher struct {
	ruleService RuleService
}

func (m RegexMatcher) Match(ctx context.Context, reqMethod string, reqURL *url.URL) (*rule.Rule, error) {
	ruleset, err := m.ruleService.GetAllConfigs(ctx)
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
	return nil, rule.ErrUnknown
}

func NewRegexMatcher(ruleService RuleService) *RegexMatcher {
	return &RegexMatcher{
		ruleService: ruleService,
	}
}
