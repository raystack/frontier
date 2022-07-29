package rulematch

import (
	"context"
	"net/http"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/rule"
	"github.com/odpf/shield/internal/proxy/middleware"
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
		m.log.Info("middleware: failed to match rule", "path", req.URL.String(), "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	middleware.EnrichRule(req, matchedRule)

	// enriching context with request body to use it in hooks
	if err := middleware.EnrichRequestBody(req); err != nil {
		m.log.Info("middleware: failed to enrich ctx with request body", "err", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	m.next.ServeHTTP(rw, req)
}
