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
	PlatformDisableOrgsOnCreate = "disable_orgs_on_create"
	PlatformDisableOrgsListing  = "disable_orgs_listing"
	PlatformDisableUsersListing = "disable_users_listing"
	PlatformInviteWithRoles     = "invite_with_roles"
	PlatformInviteMailSubject   = "invite_mail_template_subject"
	PlatformInviteMailBody      = "invite_mail_template_body"

	// organization default traits
	OrganizationMailLink    = "mail_link"
	OrganizationMailOTP     = "mail_otp"
	OrganizationSocialLogin = "social_login"

	// user default traits
	UserFirstName = "first_name"
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
		Name:         PlatformDisableOrgsOnCreate,
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
		Name:         PlatformDisableOrgsListing,
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
		Name:         PlatformDisableUsersListing,
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
		Name:         PlatformInviteWithRoles,
		Title:        "Invite With Roles",
		Description:  "Allow inviting new members with set of role ids. When the invitation is accepted, the user will be added to the org with the roles specified. This can be a security risk if the user who is inviting is not careful about the roles he is adding and cause permission escalation. Note: this is dangerous and should be used with caution. Default is false.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputCheckbox,
		InputHints:   "true,false",
		Default:      "true",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PlatformInviteMailSubject,
		Title:        "Invite Mail Subject",
		Description:  "The subject of the invite mail sent to new members.",
		Heading:      "Platform Settings",
		SubHeading:   "Manage platform settings and how it's members interact with the platform.",
		Input:        TraitInputText,
		Default:      "You have been invited to join an organization",
	},
	{
		ResourceType: schema.PlatformNamespace,
		Name:         PlatformInviteMailBody,
		Title:        "Invite Mail Body",
		Description:  "The body of the invite mail sent to new members. The following variables can be used in the template: {{.UserID}}, {{.Organization}} to personalize the mail for the user.",
		Default:      "<div>Hi {{.UserID}},</div><br><p>You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.</p><br><div>Thanks,<br>Team Frontier</div>",
	},
	// User Traits
	{
		ResourceType: schema.UserPrincipal,
		Name:         UserFirstName,
		Title:        "Full name",
		Description:  "Full name of the user",
		Heading:      "Profile",
		Input:        TraitInputText,
	},
	// Organization Traits
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         OrganizationSocialLogin,
		Title:        "Social Login",
		Description:  "Allow login through Google/Github/Facebook/etc single sign-on functionality.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         OrganizationMailOTP,
		Title:        "Email code",
		Description:  "Allow password less login via code delivered over email.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         OrganizationMailLink,
		Title:        "Email magic link",
		Description:  "Allow password less login via a link delivered over email.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
}
