package authz

import (
	"net/http"

	"github.com/odpf/salt/log"

	"github.com/odpf/shield/hook"
)

type Authz struct {
	log  log.Logger
	next hook.Service
}

func New(log log.Logger, next hook.Service) Authz {
	return Authz{
		log:  log,
		next: next,
	}
}

func (a Authz) Info() hook.Info {
	return hook.Info{
		Name:        "authz",
		Description: "",
	}
}

func (a Authz) ServeHook(res *http.Response) (*http.Response, error) {
	return a.next.ServeHook(res)
}
