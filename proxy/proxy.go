package proxy

import (
	"net/http"
)

type RequestDirector interface {
	// Build prepares a request that will be sent to backend service
	Direct(*http.Request)
}