package proxy

import (
	"context"
	"github.com/odpf/shield/structs"
	"net/http"
	"net/url"
	"strings"
)

const (
	RequestErrorKey = "proxy_err"
)

type Handler struct {
	ruleMatcher structs.RuleMatcher
	authorizers []structs.Authorizer
}

func NewHandler(ruleMatcher structs.RuleMatcher, authz []structs.Authorizer) *Handler {
	return &Handler{
		ruleMatcher: ruleMatcher,
		authorizers: authz,
	}
}

func (b Handler) Direct(req *http.Request)  {
	// find matched rule
	matchedRule, err := b.ruleMatcher.Match(req.Context(), req.Method, req.URL)
	if err != nil {
		// we failed to match any rule, don't apply the request any more
		*req = *req.WithContext(context.WithValue(req.Context(), RequestErrorKey, err))
		return
	}

	// apply rules
	if err := b.apply(req, matchedRule); err != nil {
		// we failed to apply matched rule, don't apply the request any more
		*req = *req.WithContext(context.WithValue(req.Context(), RequestErrorKey, err))
		return
	}

	// update backend request to match rules
	target, err := url.Parse(matchedRule.Backend.URL)
	if err != nil {
		// backend is not configured properly
		*req = *req.WithContext(context.WithValue(req.Context(), RequestErrorKey, err))
		return
	}
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
	if target.RawQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
	req.Header.Set("proxy-by", "shield")
}

// apply make sure the request is allowed & ready to be sent to backend
func (h Handler) apply(req *http.Request, rule *structs.Rule) (error) {
	for _, auth := range h.authorizers {
		if err := auth.Do(req, rule); err != nil {
			return err
		}
	}
	return nil
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}