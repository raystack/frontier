package prefix

import (
	"github.com/odpf/shield/internal/proxy/middleware/attributes"
	"net/http"
	"strings"

	"github.com/odpf/salt/log"
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
	permissionAttributes, _ := attributes.GetAttributesFromContext(req.Context())
	prefixInterface := permissionAttributes["prefix"]
	if prefixInterface == nil {
		w.next.ServeHTTP(rw, req)
		return
	}

	prefixStr := prefixInterface.(string)
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
