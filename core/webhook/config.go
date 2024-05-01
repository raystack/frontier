package webhook

type Config struct {
	EncryptionKey string `yaml:"encryption_key" mapstructure:"encryption_key" default:"hash-secret-should-be-32-chars--"`
}
