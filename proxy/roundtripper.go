package proxy

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/odpf/salt/log"

	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	transport *http2.Transport
	log       log.Logger
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	// we need to apply errors if it failed in Director
	if err, ok := req.Context().Value(CtxRequestErrorKey).(error); ok {
		return nil, err
	}
	t.log.Debug("proxy request", "host", req.URL.Host, "path", req.URL.Path,
		"scheme", req.URL.Scheme, "protocol", req.Proto)
	return t.transport.RoundTrip(req)
}

func NewH2cRoundTripper(log log.Logger) http.RoundTripper {
	transport := &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		AllowHTTP: true,
	}
	return &h2cTransportWrapper{
		transport: transport,
		log:       log,
	}
}
