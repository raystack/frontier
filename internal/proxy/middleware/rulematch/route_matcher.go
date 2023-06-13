package rulematch

import (
	"net/http"

	"github.com/raystack/shield/core/rule"
	"github.com/raystack/shield/internal/proxy/middleware"

	"github.com/gorilla/mux"
)

type RouteMatcher struct {
	ruleService RuleService
}

func (r RouteMatcher) Match(req *http.Request) (*rule.Rule, error) {
	ruleset, err := r.ruleService.GetAllConfigs(req.Context())
	if err != nil {
		return nil, err
	}

	for _, set := range ruleset {
		for _, rule := range set.Rules {
			router := mux.NewRouter()
			router.StrictSlash(true)
			route := router.NewRoute().Path(rule.Frontend.URL).Methods(rule.Frontend.Method)
			routeMatcher := mux.RouteMatch{}
			if route.Match(req, &routeMatcher) {
				middleware.EnrichPathParams(req, routeMatcher.Vars)
				return &rule, nil
			}
		}
	}
	return nil, rule.ErrUnknown
}

func NewRouteMatcher(ruleService RuleService) *RouteMatcher {
	return &RouteMatcher{
		ruleService: ruleService,
	}
}
