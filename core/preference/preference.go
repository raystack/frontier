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

type Trait struct {
	ResourceType    string     `json:"resource_type"`
	Name            string     `json:"name"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	LongDescription string     `json:"long_description"`
	Heading         string     `json:"heading"`
	SubHeading      string     `json:"sub_heading"`
	Breadcrumb      string     `json:"breadcrumb"`
	Input           TraitInput `json:"input"`
	InputHints      string     `json:"input_hints"`
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
	// User Traits
	{
		ResourceType: schema.UserPrincipal,
		Name:         "profile_picture",
		Title:        "Profile picture",
		Description:  "Profile picture of the user",
		Heading:      "Profile",
		Input:        TraitInputText,
	},
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
		Name:         "logo",
		Title:        "Logo",
		Description:  "Select a logo for your organization.",
		Heading:      "General",
		Input:        TraitInputText,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         "social_login",
		Title:        "Social Login",
		Description:  "Allow login through Google/Github/Facebook/etc single sign-on functionality.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         "mail_otp",
		Title:        "Email code",
		Description:  "Allow password less login via code delivered over email.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
	{
		ResourceType: schema.OrganizationNamespace,
		Name:         "mail_link",
		Title:        "Email magic link",
		Description:  "Allow password less login via a link delivered over email.",
		Heading:      "Security",
		SubHeading:   "Manage organization security and how it's members authenticate.",
		Input:        TraitInputCheckbox,
	},
}
