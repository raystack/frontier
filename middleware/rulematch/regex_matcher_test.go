package rulematch

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/odpf/shield/store/mocks"
	"github.com/odpf/shield/structs"
	"github.com/stretchr/testify/suite"
)

type RegexMatcherTest struct {
	suite.Suite
	regexMatcher RegexMatcher
	ruleRepo     mocks.RuleRepository
}

func (s *RegexMatcherTest) SetupTest() {
	s.ruleRepo = mocks.RuleRepository{}
	s.regexMatcher = RegexMatcher{
		ruleRepo: &s.ruleRepo,
	}
}

func (s *RegexMatcherTest) TestRegexMatcher_Match_ErrorFetchingRules() {
	ErrorFetchingRules := errors.New("error fetching rules")
	s.ruleRepo.On("GetAll", context.Background()).Return(nil, ErrorFetchingRules) // error fetching rules

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/foo"},
	}

	rule, err := s.regexMatcher.Match(context.Background(), req.Method, req.URL)
	s.ErrorIs(err, ErrorFetchingRules)
	s.Nil(rule)
}

func (s *RegexMatcherTest) TestRegexMatcher_Match_NoRules() {
	s.ruleRepo.On("GetAll", context.Background()).Return([]structs.Ruleset{}, nil) // no rules

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/foo"},
	}

	rule, err := s.regexMatcher.Match(context.Background(), req.Method, req.URL)
	s.ErrorIs(err, ErrUnknownRule)
	s.Nil(rule)
}

func (s *RegexMatcherTest) TestRegexMatcher_Match_NoMatchingRules() {
	ruleSet := []structs.Ruleset{{
		Rules: []structs.Rule{
			{
				Frontend:    structs.Frontend{URL: "/foo", Method: "GET", URLRx: regexp.MustCompile("^/foo$")},
				Backend:     structs.Backend{},
				Middlewares: nil,
				Hooks:       nil,
			},
		},
	}}
	s.ruleRepo.On("GetAll", context.Background()).Return(ruleSet, nil) // no matching rules

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/bar"}, // no match
	}

	rule, err := s.regexMatcher.Match(context.Background(), req.Method, req.URL)
	s.ErrorIs(err, ErrUnknownRule)
	s.Nil(rule)
}

func (s *RegexMatcherTest) TestRegexMatcher_Match_MatchingRules() {
	ruleSet := []structs.Ruleset{{
		Rules: []structs.Rule{
			{
				Frontend:    structs.Frontend{URL: "/foo", Method: "POST", URLRx: regexp.MustCompile("^/foo$")},
				Backend:     structs.Backend{},
				Middlewares: nil,
				Hooks:       nil,
			},
			{
				Frontend:    structs.Frontend{URL: "/foo", Method: "GET", URLRx: regexp.MustCompile("^/foo$")},
				Backend:     structs.Backend{},
				Middlewares: nil,
				Hooks:       nil,
			},
		},
	}}
	s.ruleRepo.On("GetAll", context.Background()).Return(ruleSet, nil) // matching rules

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/foo"}, // match
	}

	rule, err := s.regexMatcher.Match(context.Background(), req.Method, req.URL)
	s.NoError(err)
	s.NotNil(rule)
	s.Equal(ruleSet[0].Rules[1], *rule)
}

func TestNewRegexMatcher(t *testing.T) {
	suite.Run(t, new(RegexMatcherTest))
}
