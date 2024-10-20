package db

import "time"

type Config struct {
	Driver          string        `yaml:"driver" mapstructure:"driver" default:"postgres"`
	URL             string        `yaml:"url" mapstructure:"url"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns" default:"10"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns" default:"10"`
	ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time" mapstructure:"conn_max_life_time" default:"15m"`
	MaxQueryTimeout time.Duration `yaml:"max_query_timeout" mapstructure:"max_query_timeout" default:"5s"`
}
