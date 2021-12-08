package structs

import (
	"net/http"
	"regexp"
)

type Ruleset struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Frontend    Frontend        `yaml:"frontend"`
	Backend     Backend         `yaml:"backend"`
	Middlewares MiddlewareSpecs `yaml:"middlewares"`
}

type MiddlewareSpec struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}

type MiddlewareSpecs []MiddlewareSpec

func (m MiddlewareSpecs) Get(name string) (MiddlewareSpec, bool) {
	for _, n := range m {
		if n.Name == name {
			return n, true
		}
	}
	return MiddlewareSpec{}, false
}

type Frontend struct {
	URL   string         `yaml:"url"`
	URLRx *regexp.Regexp `yaml:"-"`

	Method string `yaml:"method"`
}

type Backend struct {
	URL string `yaml:"url"`
}

type RuleMatcher interface {
	Match(req *http.Request) (*Rule, error)
}
