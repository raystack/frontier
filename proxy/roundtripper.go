package proxy

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"net"
	"net/http"
)

type h2cTransportWrapper struct {
	transport *http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	// we need to apply errors if it failed in Director
	reqErr := req.Context().Value(RequestErrorKey)
	if reqErr != nil {
		return nil, reqErr.(error)
	}
	return t.transport.RoundTrip(req)
}

func NewH2cRoundTripper() http.RoundTripper {
	transport := &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		AllowHTTP: true,
	}
	return &h2cTransportWrapper{
		transport: transport,
	}
}