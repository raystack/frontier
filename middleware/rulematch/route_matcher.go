package rulematch

import (
	"net/http"

	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"

	"github.com/gorilla/mux"
)

type RouteMatcher struct {
	ruleRepo store.RuleRepository
}

func (r RouteMatcher) Match(req *http.Request) (*structs.Rule, error) {
	ruleset, err := r.ruleRepo.GetAll(req.Context())
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
	return nil, ErrUnknownRule
}

func NewRouteMatcher(ruleRepo store.RuleRepository) *RouteMatcher {
	return &RouteMatcher{
		ruleRepo: ruleRepo,
	}
}
