package pipeline

import (
	"context"
	"errors"
	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
	"net/url"
	"regexp"
)

var (
	ErrUnknownRule = errors.New("undefined proxy rule")
)

type RegexMatcher struct {

	// TODO: should be a factory instead
	ruleRepo store.RuleRepository
}

func (m RegexMatcher) Match(ctx context.Context, reqMethod string, reqURL *url.URL) (*structs.Rule, error) {

	// TODO: make sure this call is properly cached
	services, err := m.ruleRepo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		for _, rule := range service.Rules {
			// TODO: we should check if the frontend url is a valid regex when reading
			// the spec
			exp, err := regexp.Compile(rule.Frontend.URL)
			if err != nil {
				return nil, err
			}

			//reqFriendlyUrl := fmt.Sprintf("%s://%s%s", reqURL.Scheme, reqURL.Host, reqURL.Path)
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



