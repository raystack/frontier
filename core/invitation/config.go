package invitation

type Config struct {
	// InvitationWithRoles if set to true will allow roles to be passed in invitation, when the user accepts the
	// invite, the role will be assigned to the user
	InvitationWithRoles bool `yaml:"invitation_with_roles" mapstructure:"invitation_with_roles" default:"false"`

	MailInvite MailInviteConfig `yaml:"mail_invite" mapstructure:"mail_invite"`
}

type MailInviteConfig struct {
	Subject string `yaml:"subject" mapstructure:"subject" default:"You have been invited to join an organization"`
	Body    string `yaml:"body" mapstructure:"body" default:" <div>Hi Hi {{.UserID}},</div><br><p>You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.</p><br><div>Thanks,<br>Team Frontier</div>"`
}
