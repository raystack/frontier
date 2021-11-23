package proxy

import (
	"net/http"

	"github.com/odpf/shield/hook"

	"github.com/odpf/salt/log"
)

type h2cTransportWrapper struct {
	transport *http.Transport
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

	return t.hook.ServeHook(res)
}

func NewH2cRoundTripper(log log.Logger, hook hook.Service) http.RoundTripper {
	transport := &http.Transport{}

	return &h2cTransportWrapper{
		transport: transport,
		log:       log,
		hook:      hook,
	}
}
