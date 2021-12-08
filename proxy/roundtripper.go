package proxy

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/odpf/salt/log"

	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	// Defining two different RoundTripper
	// - httptransport: for http, h2, h2c
	// - http2transport: h2c, grpc
	// this is because &http2.Transport is not supporting
	// proxy for http & h2
	// Reference: https://sourcegraph.com/github.com/tsenart/vegeta/-/blob/lib/attack.go?L206:6#tab=references

	httpTransport *http.Transport
	grpcTransport *http2.Transport

	log log.Logger
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	// we need to apply errors if it failed in Director
	if err, ok := req.Context().Value(CtxRequestErrorKey).(error); ok {
		return nil, err
	}
	t.log.Debug("proxy request", "host", req.URL.Host, "path", req.URL.Path,
		"scheme", req.URL.Scheme, "protocol", req.Proto)

	var transport http.RoundTripper = t.httpTransport
	if req.Header.Get("Content-Type") == "application/grpc" {
		transport = t.grpcTransport
	}

	return transport.RoundTrip(req)
}

func NewH2cRoundTripper(log log.Logger) http.RoundTripper {
	return &h2cTransportWrapper{
		httpTransport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 1 * time.Minute,
			}).DialContext,
		},
		grpcTransport: &http2.Transport{
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
			AllowHTTP: true,
		},
		log: log,
	}
}
