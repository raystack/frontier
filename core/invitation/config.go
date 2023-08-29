package invitation

type Config struct {
	// WithRoles if set to true will allow roles to be passed in invitation, when the user accepts the
	// invite, the role will be assigned to the user
	WithRoles bool `yaml:"with_roles" mapstructure:"with_roles" default:"false"`

	MailTemplate MailTemplateConfig `yaml:"mail_template" mapstructure:"mail_template"`
}

type MailTemplateConfig struct {
	Subject string `yaml:"subject" mapstructure:"subject" default:"You have been invited to join an organization"`
	Body    string `yaml:"body" mapstructure:"body" default:"<div>Hi {{.UserID}},</div><br><p>You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.</p><br><div>Thanks,<br>Team Frontier</div>"`
}
