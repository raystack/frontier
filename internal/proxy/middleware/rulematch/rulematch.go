package rulematch

import (
	"context"
	"net/http"

	"github.com/raystack/shield/core/rule"
	"github.com/raystack/shield/internal/proxy/middleware"

	"github.com/raystack/salt/log"
)

type RuleService interface {
	GetAllConfigs(ctx context.Context) ([]rule.Ruleset, error)
}

type RuleMatcher interface {
	Match(req *http.Request) (*rule.Rule, error)
}

type Ware struct {
	log         log.Logger
	next        http.Handler
	ruleMatcher RuleMatcher
}

func New(log log.Logger, next http.Handler, matcher RuleMatcher) *Ware {
	return &Ware{
		log:         log,
		next:        next,
		ruleMatcher: matcher,
	}
}

func (m Ware) Info() *middleware.MiddlewareInfo {
	return &middleware.MiddlewareInfo{
		Name:        "_rulematch",
		Description: "match request with service rule set and enrich context",
	}
}

func (m *Ware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// find matched rule
	matchedRule, err := m.ruleMatcher.Match(req)
	if err != nil {
		m.log.Error("middleware", "rulematch", "error_matching_rule", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	middleware.EnrichRule(req, matchedRule)

	// enriching context with request body to use it in hooks
	if err = middleware.EnrichRequestBody(req); err != nil {
		m.log.Error("middleware", "rulematch", "error_enriching_request_body", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	m.next.ServeHTTP(rw, req)
}
