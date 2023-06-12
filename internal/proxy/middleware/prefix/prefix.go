package prefix

import (
	"net/http"
	"strings"

	"github.com/raystack/salt/log"
	"github.com/raystack/shield/internal/proxy/middleware"
)

type Ware struct {
	next http.Handler
	log  log.Logger
}

func New(log log.Logger, next http.Handler) *Ware {
	return &Ware{
		next: next,
		log:  log,
	}
}

func (w *Ware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rules, ok := middleware.ExtractRule(req)
	if !ok {
		w.next.ServeHTTP(rw, req)
		return
	}
	prefixStr := rules.Backend.Prefix
	req.URL.Path = w.getPrefixStripped(req.URL.Path, prefixStr)
	if req.URL.RawPath != "" {
		req.URL.RawPath = w.getPrefixStripped(req.URL.RawPath, prefixStr)
	}
	w.next.ServeHTTP(rw, req)
}

func (w *Ware) getPrefixStripped(urlPath, prefix string) string {
	str := strings.TrimPrefix(urlPath, prefix)
	// ensure leading slash
	if str == "" {
		return str
	}
	if str[0] == '/' {
		return str
	}
	return "/" + str
}
