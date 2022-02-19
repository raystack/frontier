package rulematch

import (
	"context"
	"errors"
	"github.com/odpf/shield/lib/mocks"
	"github.com/odpf/shield/structs"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/url"
	"testing"
)

type RouteMatcherTestSuite struct {
	suite.Suite
	routeMatcher RouteMatcher
	ruleRepo     mocks.RuleRepository
}

func (s *RouteMatcherTestSuite) SetupTest() {
	s.ruleRepo = mocks.RuleRepository{}
	s.routeMatcher = RouteMatcher{
		ruleRepo: &s.ruleRepo,
	}
}

func (s *RouteMatcherTestSuite) TestRouteMatcher_Match_ErrorFetchingRules() {
	ErrorFetchingRules := errors.New("error fetching rules")
	s.ruleRepo.On("GetAll", context.Background()).Return(nil, ErrorFetchingRules) // error fetching rules

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/foo"},
	}
	rule, err := s.routeMatcher.Match(req)
	s.ErrorIs(err, ErrorFetchingRules)
	s.Nil(rule)
}

func (s *RouteMatcherTestSuite) TestRouteMatcher_Match_NoRules() {
	s.ruleRepo.On("GetAll", context.Background()).Return([]structs.Ruleset{}, nil) // empty ruleset

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/foo"},
	}
	rule, err := s.routeMatcher.Match(req)
	s.ErrorIs(err, ErrUnknownRule)
	s.Nil(rule)
}

func (s *RouteMatcherTestSuite) TestRouteMatcher_Match_NoMatchingRules() {
	ruleSet := []structs.Ruleset{{
		Rules: []structs.Rule{
			{
				Frontend:    structs.Frontend{URL: "/foo", Method: "GET"},
				Backend:     structs.Backend{},
				Middlewares: nil,
				Hooks:       nil,
			},
		},
	}}
	s.ruleRepo.On("GetAll", context.Background()).Return(ruleSet, nil)

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/bar"}, // no match
	}
	rule, err := s.routeMatcher.Match(req)
	s.ErrorIs(err, ErrUnknownRule)
	s.Nil(rule)
}

func (s *RouteMatcherTestSuite) TestRouteMatcher_Match_MatchingRule() {
	ruleSet := []structs.Ruleset{{
		Rules: []structs.Rule{
			{
				Frontend:    structs.Frontend{URL: "/foo", Method: "GET"},
				Backend:     structs.Backend{},
				Middlewares: nil,
				Hooks:       nil,
			},
			{
				Frontend:    structs.Frontend{URL: "/foo", Method: "POST"},
				Backend:     structs.Backend{},
				Middlewares: nil,
				Hooks:       nil,
			},
		},
	}}
	s.ruleRepo.On("GetAll", context.Background()).Return(ruleSet, nil)

	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/foo"}, // match
	}
	rule, err := s.routeMatcher.Match(req)
	s.NoError(err)
	s.Equal(ruleSet[0].Rules[0], *rule)
}

func TestNewRouteMatcher(t *testing.T) {
	suite.Run(t, new(RouteMatcherTestSuite))
}
