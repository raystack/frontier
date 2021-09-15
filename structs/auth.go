package structs

import "net/http"

type Authorizer interface {
	Do(*http.Request, *Rule) error
}

type Authenticator interface {
	Do(*http.Request, *Rule) error
}
