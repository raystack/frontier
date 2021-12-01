package proxy

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/odpf/shield/hook"

	"github.com/odpf/salt/log"

	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	transport *http2.Transport
	log       log.Logger
	hook      hook.Service
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	// we need to apply errors if it failed in Director
	if err, ok := req.Context().Value(CtxRequestErrorKey).(error); ok {
		return nil, err
	}
	t.log.Debug("proxy request", "host", req.URL.Host, "path", req.URL.Path,
		"scheme", req.URL.Scheme, "protocol", req.Proto)

	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return res, err
	}

	return t.hook.ServeHook(res, nil)
}

func NewH2cRoundTripper(log log.Logger, hook hook.Service) http.RoundTripper {
	transport := &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		AllowHTTP: true,
	}
	return &h2cTransportWrapper{
		transport: transport,
		log:       log,
		hook:      hook,
	}
}
