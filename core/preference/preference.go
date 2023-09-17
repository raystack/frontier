package preference

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

var (
	ErrInvalidID     = fmt.Errorf("invalid preference id")
	ErrNotFound      = fmt.Errorf("preference not found")
	ErrInvalidFilter = fmt.Errorf("invalid preference filter set")
	ErrTraitNotFound = fmt.Errorf("preference trait not found, preferences can only be created with valid trait")
)

type TraitInput string

const (
	TraitInputText        TraitInput = "text"
	TraitInputTextarea    TraitInput = "textarea"
	TraitInputSelect      TraitInput = "select"
	TraitInputCombobox    TraitInput = "combobox"
	TraitInputCheckbox    TraitInput = "checkbox"
	TraitInputMultiselect TraitInput = "multiselect"
	TraitInputNumber      TraitInput = "number"
)

const (
	// platform default traits
	PreferenceDisableOrgsOnCreate = "disable_orgs_on_create"
	PreferenceDisableOrgsListing  = "disable_orgs_listing"
	PreferenceDisableUsersListing = "disable_users_listing"
	PreferenceInviteWithRoles     = "invite_with_roles"
	PreferenceInviteMailSubject   = "invite_mail_template_subject"
	PreferenceInviteMailBody      = "invite_mail_template_body"
	PreferenceMailOTPBody         = "oidc_mail_otp_body"
	PreferenceMailOTPSubject      = "oidc_mail_otp_subject"
	PreferenceMailOTPValidity     = "oidc_mail_otp_validity"
	PreferenceSocialLogin         = "social_login"
	PreferenceMailLinkSubject     = "mail_link_subject"
	PreferenceMailLinkBody        = "mail_link_body"
	PreferenceMailLinkValidity    = "mail_link_validity"

	// organization default traits
	PreferenceMailLink = "mail_link"
	PreferenceMailOTP  = "mail_otp"
)

type Trait struct {
	// Level at which the trait is applicable (say "platform", "organization", "user")
	ResourceType string `json:"resource_type"`
	// Name of the trait (say "disable_orgs_on_create", "disable_orgs_listing", "disable_users_listing", "invite_with_roles", "invite_mail_template_subject", "invite_mail_template_body")
	Name string `json:"name"`
	// Readable name of the trait (say "Disable Orgs On Create", "Disable Orgs Listing")
	Title           string `json:"title"`
	Description     string `json:"description"`
	LongDescription string `json:"long_description"`
	Heading         string `json:"heading"`
	SubHeading      string `json:"sub_heading"`
	// Breadcrumb to be used to group the trait with other traits (say "Platform.Settings.Authentication", "Platform.Settings.Invitation")
	Breadcrumb string `json:"breadcrumb"`
	// Type of input to be used to collect the value for the trait (say "text", "select", "checkbox", etc.)
	Input TraitInput `json:"input"`
	// Acceptable values to be provided in the input (say "true,false") for a TraitInput of type Checkbox
	InputHints string `json:"input_hints"`
	// Default value to be used for the trait if the preference is not set (say "true" for a TraitInput of type Checkbox)
	Default string `json:"default"`
}

type Preference struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Value        string    `json:"value"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var DefaultTraits = []Trait{
	// Platform Traits
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceDisableOrgsOnCreate,
		Title:        "Disable Orgs On Create",
		Description:  "If selected the new orgs created by members will be disabled by default.This can be used to prevent users from accessing the org until they contact the admin and get it enabled. Default is false.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputCheckbox,
		InputHints:   "true,false",
		Default:      "false",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceDisableOrgsListing,
		Title:        "Disable Orgs Listing",
		Description:  "If selected will disallow non-admin APIs to list all organizations on the platform. Default is false.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputCheckbox,
		InputHints:   "true,false",
		Default:      "false",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceDisableUsersListing,
		Title:        "Disable Users Listing",
		Description:  "If selected will will disallow non-admin APIs to list all users on the platform. Default is false.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputCheckbox,
		InputHints:   "true,false",
		Default:      "false",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceInviteWithRoles,
		Title:        "Invite With Roles",
		Description:  "Allow inviting new members with set of role ids. When the invitation is accepted, the user will be added to the org with the roles specifiedThis can be a security risk if the user who is inviting is not careful about the roles he is adding and cause permission escalation. Note: this is dangerous and should be used with caution. Default is false.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputCheckbox,
		InputHints:   "true,false",
		Default:      "true",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceInviteMailSubject,
		Title:        "Invite Mail Subject",
		Description:  "The subject of the invite mail sent to new members.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		Default:      "You have been invited to join an organization",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceInviteMailBody,
		Title:        "Invite Mail Body",
		Description:  "The body of the invite mail sent to new members.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		Default:      "<div>Hi {{.UserID}},</div><br><p>You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.</p><br><div>Thanks,<br>Team Frontier</div>",
	},
	{
		ResourceType:    schema.PlatformNamespace,
		Name:            PreferenceMailOTPSubject,
		Title:           "Mail OTP Subject",
		Description:     "Allow password less login via code delivered over email. This field is used to set the subject of the email.",
		LongDescription: "The user must retrieve the OTP from their email and enter it into the application within a specified time frame to gain access.",
		Heading:         "Platform Settings",
		SubHeading:      "Manage platform settings and how it's members interact with the platform.",
		Input:           TraitInputText,
		Default:         "Frontier - Login Link",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceMailOTPBody,
		Title:        "Mail OTP Body",
		Description:  "The body of the OTP mail to be sent to new members for OIDC authentication.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		Default:      "Please copy/paste the OneTimePassword in login form.<h2>{{.Otp}}</h2>This code will expire in 15 minutes.",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceMailOTPValidity,
		Title:        "Mail OTP Validity",
		Description:  "The expiry time until which the mail OTP is valid. Default is 15 minutes.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		InputHints:   "15m,1h,1d",
		Default:      "15m",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceMailLinkSubject,
		Title:        "Mail Link Subject",
		Description:  "Allow password less login via a link delivered over email. This field is used to set the subject of the email.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		Default:      "Frontier Login - One time link",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceMailLinkBody,
		Title:        "Mail Link Body",
		Description:  "The body of the mail with clickable link for authentication.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		Default:      "Click on the following link or copy/paste the url in browser to login.<br><h2><a href='{{.Link}}' target='_blank'>Login</a></h2><br>Address: {{.Link}} <br>This link will expire in 15 minutes.",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PreferenceMailLinkValidity,
		Title:        "Mail Link Validity",
		Description:  "The expiry time until which the mail link is valid. Default is 15 minutes.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		InputHints:   "15m,1h,1d",
		Default:      "15m",
	},
	// User Traits
	{
		ResourceType: schema.UserPrincipal,
		Name:         "first_name",
		Title:        "Full name",
		Description:  "Full name of the user",
		Heading:      "Profile",
		Input:        TraitInputText,
	},
	// Organization Traits
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         PreferenceSocialLogin,
		Title:        "Social Login",
		Description:  "Allow login through Google/Github/Facebook/etc single sign-on functionality.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         PreferenceMailOTP,
		Title:        "Email code",
		Description:  "Allow password less login via code delivered over email.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         PreferenceMailLink,
		Title:        "Email magic link",
		Description:  "Allow password less login via a link delivered over email.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
}
