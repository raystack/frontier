package rulematch

import (
	"errors"
	"net/http"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/structs"
)

var (
	ErrUnknownRule = errors.New("undefined proxy rule")
)

type Ware struct {
	log         log.Logger
	next        http.Handler
	ruleMatcher structs.RuleMatcher
}

func New(log log.Logger, next http.Handler, matcher structs.RuleMatcher) *Ware {
	return &Ware{
		log:         log,
		next:        next,
		ruleMatcher: matcher,
	}
}

func (m Ware) Info() *structs.MiddlewareInfo {
	return &structs.MiddlewareInfo{
		Name:        "_rulematch",
		Description: "match request with service rule set and enrich context",
	}
}

func (m *Ware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// find matched rule
	matchedRule, err := m.ruleMatcher.Match(req.Context(), req.Method, req.URL)
	if err != nil {
		m.log.Info("middleware: failed to match rule", "path", req.URL.String(), "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	middleware.EnrichRule(req, matchedRule)
	m.next.ServeHTTP(rw, req)
}
