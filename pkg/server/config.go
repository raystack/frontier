package server

import (
	"fmt"

	"github.com/raystack/frontier/core/webhook"

	"github.com/raystack/frontier/pkg/server/interceptors"

	"github.com/raystack/frontier/pkg/mailer"

	"github.com/raystack/frontier/internal/bootstrap"

	"github.com/raystack/frontier/core/authenticate"
)

type GRPCConfig struct {
	Port            int    `mapstructure:"port" default:"8081"`
	MaxRecvMsgSize  int    `mapstructure:"max_recv_msg_size" default:"33554432"`
	MaxSendMsgSize  int    `mapstructure:"max_send_msg_size" default:"33554432"`
	TLSCertFile     string `mapstructure:"tls_cert_file" default:""`
	TLSKeyFile      string `mapstructure:"tls_key_file" default:""`
	TLSClientCAFile string `mapstructure:"tls_client_ca_file" default:""`
}

func (cfg Config) grpcAddr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPC.Port) }

type UIConfig struct {
	Port              int      `yaml:"port" mapstructure:"port"`
	Title             string   `yaml:"title" mapstructure:"title"`
	Logo              string   `yaml:"logo" mapstructure:"logo"`
	AppURL            string   `yaml:"app_url" mapstructure:"app_url"`
	TokenProductId    string   `yaml:"token_product_id" mapstructure:"token_product_id"`
	OrganizationTypes []string `yaml:"organization_types" mapstructure:"organization_types"`
}

type ConnectHeader struct {
	ClientIP      string `yaml:"client_ip" mapstructure:"client_ip" default:"x-frontier-ip"`
	ClientCountry string `yaml:"client_country" mapstructure:"client_country" default:"x-frontier-country"`
	ClientCity    string `yaml:"client_city" mapstructure:"client_city" default:"x-frontier-city"`
}

type ConnectConfig struct {
	// port to listen buf connect requests on
	Port int `yaml:"port" mapstructure:"port" default:"8002"`
	// headers to extract from http headers
	Headers ConnectHeader `yaml:"headers" mapstructure:"headers"`
}

type Config struct {
	// Connect server config
	Connect ConnectConfig `yaml:"connect" mapstructure:"connect"`
	// port to listen HTTP requests on
	Port int `yaml:"port" mapstructure:"port" default:"8080"`

	// GRPC Config
	GRPC GRPCConfig `mapstructure:"grpc"`

	// metrics port
	MetricsPort int `yaml:"metrics_port" mapstructure:"metrics_port" default:"9000"`

	// Profiler enables /debug/pprof under metrics port
	Profiler bool `yaml:"profiler" mapstructure:"profiler" default:"false"`

	// the network interface to listen on
	Host string `yaml:"host" mapstructure:"host" default:"127.0.0.1"`

	// TODO might not suitable here because it is also being used by proxy
	// Headers which will have user's email id
	IdentityProxyHeader string `yaml:"identity_proxy_header" mapstructure:"identity_proxy_header" default:""`

	// ResourcesPath is a directory path where resources is defined
	// that this service should implement
	ResourcesConfigPath string `yaml:"resources_config_path" mapstructure:"resources_config_path"`

	// ResourcesPathSecretSecret could be an env name, file path or actual value required
	// to access ResourcesPathSecretPath files
	ResourcesConfigPathSecret string `yaml:"resources_config_path_secret" mapstructure:"resources_config_path_secret"`

	Authentication authenticate.Config `yaml:"authentication" mapstructure:"authentication"`

	// Deprecated: use Cors instead
	CorsOrigin []string `yaml:"cors_origin" mapstructure:"cors_origin"`
	// Cors configuration setup origin value from where we want to allow cors
	// headers and methods are the list of headers and methods we want to allow
	Cors interceptors.CorsConfig `yaml:"cors" mapstructure:"cors"`

	Admin bootstrap.AdminConfig `yaml:"admin" mapstructure:"admin"`

	Mailer mailer.Config `yaml:"mailer" mapstructure:"mailer"`

	Webhook webhook.Config `yaml:"webhook" mapstructure:"webhook"`
}
