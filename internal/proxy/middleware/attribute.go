package middleware

const (
	AttributeTypeQuery       AttributeType = "query"
	AttributeTypeHeader      AttributeType = "header"
	AttributeTypeJSONPayload AttributeType = "json_payload"
	AttributeTypeGRPCPayload AttributeType = "grpc_payload"
	AttributeTypePathParam   AttributeType = "path_param"
	AttributeTypeConstant    AttributeType = "constant"
)

type AttributeType string

type Attribute struct {
	Key    string        `yaml:"key" mapstructure:"key"`
	Type   AttributeType `yaml:"type" mapstructure:"type"`
	Index  string        `yaml:"index" mapstructure:"index"` // proto index
	Path   string        `yaml:"path" mapstructure:"path"`
	Params []string      `yaml:"params" mapstructure:"params"`
	Value  string        `yaml:"value" mapstructure:"value"`
}
