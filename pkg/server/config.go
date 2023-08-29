package server

import (
	"fmt"

	"github.com/raystack/frontier/pkg/mailer"

	"github.com/raystack/frontier/internal/bootstrap"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/pkg/telemetry"
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

type Config struct {
	// port to listen HTTP requests on
	Port int `yaml:"port" mapstructure:"port" default:"8080"`

	// GRPC Config
	GRPC GRPCConfig `mapstructure:"grpc"`

	// metrics port
	MetricsPort int `yaml:"metrics_port" mapstructure:"metrics_port" default:"9000"`

	// the network interface to listen on
	Host string `yaml:"host" mapstructure:"host" default:"127.0.0.1"`

	// TODO might not suitable here because it is also being used by proxy
	// Headers which will have user's email id
	IdentityProxyHeader string `yaml:"identity_proxy_header" mapstructure:"identity_proxy_header" default:""`

	// Header which will have user_id
	UserIDHeader string `yaml:"user_id_header" mapstructure:"user_id_header" default:"X-Frontier-User-Id"`

	// ResourcesPath is a directory path where resources is defined
	// that this service should implement
	ResourcesConfigPath string `yaml:"resources_config_path" mapstructure:"resources_config_path"`

	// ResourcesPathSecretSecret could be an env name, file path or actual value required
	// to access ResourcesPathSecretPath files
	ResourcesConfigPathSecret string `yaml:"resources_config_path_secret" mapstructure:"resources_config_path_secret"`

	TelemetryConfig telemetry.Config `yaml:"telemetry_config" mapstructure:"telemetry_config"`

	Authentication authenticate.Config `yaml:"authentication" mapstructure:"authentication"`

	// DisableOrgsListing if set to true will disallow non-admin APIs to list all organizations
	DisableOrgsListing bool `yaml:"disable_orgs_listing" mapstructure:"disable_orgs_listing" default:"false"`
	// DisableUsersListing if set to true will disallow non-admin APIs to list all users
	DisableUsersListing bool `yaml:"disable_users_listing" mapstructure:"disable_users_listing" default:"false"`
	// InvitationWithRoles if set to true will allow roles to be passed in invitation, when the user accepts the
	// invite, the role will be assigned to the user
	InvitationWithRoles bool `yaml:"invitation_with_roles" mapstructure:"invitation_with_roles" default:"false"`

	// CorsOrigin is origin value from where we want to allow cors
	CorsOrigin []string `yaml:"cors_origin" mapstructure:"cors_origin"`

	Admin bootstrap.AdminConfig `yaml:"admin" mapstructure:"admin"`

	Mailer mailer.Config `yaml:"mailer" mapstructure:"mailer"`
}
