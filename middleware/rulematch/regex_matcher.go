package rulematch

import (
	"context"
	"net/url"
	"regexp"

	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
)

type RegexMatcher struct {
	ruleRepo store.RuleRepository
}

func (m RegexMatcher) Match(ctx context.Context, reqMethod string, reqURL *url.URL) (*structs.Rule, error) {
	// TODO: make sure this call is properly cached
	ruleset, err := m.ruleRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, set := range ruleset {
		for _, rule := range set.Rules {
			var isMethodMatch bool
			for _, ruleMethod := range rule.Frontend.Methods {
				if reqMethod == ruleMethod {
					isMethodMatch = true
					break
				}
			}
			if !isMethodMatch {
				continue
			}

			// TODO: we should check if the frontend url is a valid regex when reading
			// the spec so it fails early
			exp, err := regexp.Compile(rule.Frontend.URL)
			if err != nil {
				return nil, err
			}

			if exp.MatchString(reqURL.String()) {
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
