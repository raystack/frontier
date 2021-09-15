package proxy

import (
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

// H2c
type H2c struct {
	proxy        *httputil.ReverseProxy
	bufferPool   httputil.BufferPool
	roundTripper http.RoundTripper

	director RequestDirector
}

func NewH2c(roundTripper http.RoundTripper, director RequestDirector) *H2c {
	return &H2c{
		proxy: &httputil.ReverseProxy{},
		bufferPool: newBufferPool(),
		roundTripper: roundTripper,
		director: director,
	}
}

func (p *H2c) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Forwarded-For", r.Host)

	p.proxy.Transport = p.roundTripper
	p.proxy.FlushInterval = 100 * time.Millisecond
	p.proxy.BufferPool = p.bufferPool
	p.proxy.Director = p.director.Direct
	p.proxy.ServeHTTP(w, r)
}

const bufferPoolSize = 32 * 1024

func newBufferPool() *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferPoolSize)
			},
		},
	}
}

type bufferPool struct {
	pool sync.Pool
}

func (b *bufferPool) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *bufferPool) Put(bytes []byte) {
	b.pool.Put(bytes)
}
