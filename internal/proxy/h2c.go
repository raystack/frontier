package proxy

import (
	"net/http"
	"net/http/httputil"
	"time"
)

type RequestDirector interface {
	// Build prepares a request that will be sent to backend service
	Direct(*http.Request)
}

// H2c
type H2c struct {
	proxy        *httputil.ReverseProxy
	bufferPool   httputil.BufferPool
	roundTripper http.RoundTripper

	director RequestDirector
}

func NewH2c(roundTripper http.RoundTripper, director RequestDirector) *H2c {
	return &H2c{
		proxy:        &httputil.ReverseProxy{},
		bufferPool:   newBufferPool(),
		roundTripper: roundTripper,
		director:     director,
	}
}

func (p *H2c) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Forwarded-For", r.Host)

	p.proxy.Transport = p.roundTripper
	p.proxy.FlushInterval = 100 * time.Millisecond
	p.proxy.BufferPool = p.bufferPool
	p.proxy.Director = p.director.Direct
	p.proxy.ServeHTTP(w, r)
}
