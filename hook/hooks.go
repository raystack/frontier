package hook

import "net/http"

type Service interface {
	Info() Info
	ServeHook(res *http.Response) (*http.Response, error)
}

type Info struct {
	Name        string
	Description string
}

type Hook struct{}

func New() Hook {
	return Hook{}
}

func (h Hook) Info() Info {
	return Info{}
}

func (h Hook) ServeHook(res *http.Response) (*http.Response, error) {
	return res, nil
}
