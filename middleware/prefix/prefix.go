package prefix

import (
	"net/http"
	"strings"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/middleware"
	"github.com/odpf/shield/structs"
)

type Ware struct {
	next http.Handler
	log  log.Logger
}

type Config struct {
	Strip string `mapstructure:"strip"`
}

func New(log log.Logger, next http.Handler) *Ware {
	return &Ware{
		next: next,
		log:  log,
	}
}

func (w Ware) Info() *structs.MiddlewareInfo {
	return &structs.MiddlewareInfo{
		Name:        "prefix",
		Description: "strip prefix from request path",
	}
}

func (w *Ware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wareSpec, ok := middleware.ExtractMiddleware(req, w.Info().Name)
	if !ok {
		w.next.ServeHTTP(rw, req)
		return
	}

	var prefixStr string
	if raw, ok := wareSpec.Config["strip"]; ok {
		prefixStr = raw.(string)
	} else {
		w.log.Debug("middleware: prefix config missing")
		w.next.ServeHTTP(rw, req)
		return
	}

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
