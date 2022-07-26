package spicedb

type Config struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port" default:"50051"`
	PreSharedKey string `yaml:"pre_shared_key" mapstructure:"pre_shared_key"`
}
