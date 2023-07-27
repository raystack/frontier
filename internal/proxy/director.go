package proxy

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/raystack/frontier/internal/proxy/middleware"
	"github.com/raystack/frontier/pkg/httputil"
)

var ctxRequestErrorKey = struct{}{}

type Director struct {
}

func NewDirector() *Director {
	return &Director{}
}

func (h Director) Direct(req *http.Request) {
	matchedRule, _ := middleware.ExtractRule(req)

	// update backend request to match rules
	target, err := url.Parse(matchedRule.Backend.URL)
	if err != nil {
		// backend is not configured properly
		*req = *req.WithContext(context.WithValue(req.Context(), ctxRequestErrorKey, err))
		return
	}
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
	req.Host = target.Host
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}

	if target.RawQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header[httputil.HeaderUserAgent]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set(httputil.HeaderUserAgent, "")
	}
	req.Header.Set("proxy-by", "frontier")
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
