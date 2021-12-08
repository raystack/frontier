package structs

type Resource struct {
	Name    string              `json:"name" yaml:"name"`
	Actions map[string][]string `json:"actions" yaml:"actions"`
}
