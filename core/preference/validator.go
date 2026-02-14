package preference

import (
	"strings"

	"golang.org/x/exp/slices"
)

type PreferenceValidator interface {
	Validate(value string) bool
}

type BooleanValidator struct {
	allowedValues []string
}

func NewBooleanValidator() *BooleanValidator {
	return &BooleanValidator{
		allowedValues: []string{"true", "false"},
	}
}

func (v *BooleanValidator) Validate(value string) bool {
	return slices.Contains(v.allowedValues, value)
}

type TextValidator struct{}

func NewTextValidator() *TextValidator {
	return &TextValidator{}
}

func (v *TextValidator) Validate(value string) bool {
	if strings.TrimSpace(value) != "" {
		return true // accepts any non-empty string
	}
	return false
}

// SelectValidator validates that a value is one of the allowed options
// specified in the InputHints field (comma-separated values)
type SelectValidator struct {
	allowedValues []string
}

func NewSelectValidator(inputHints string) *SelectValidator {
	var allowed []string
	for _, v := range strings.Split(inputHints, ",") {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			allowed = append(allowed, trimmed)
		}
	}
	return &SelectValidator{
		allowedValues: allowed,
	}
}

// NewSelectValidatorFromOptions creates a SelectValidator from InputHintOptions
// It extracts the Name field from each option as the allowed value
func NewSelectValidatorFromOptions(options []InputHintOption) *SelectValidator {
	var allowed []string
	for _, opt := range options {
		if opt.Name != "" {
			allowed = append(allowed, opt.Name)
		}
	}
	return &SelectValidator{
		allowedValues: allowed,
	}
}

func (v *SelectValidator) Validate(value string) bool {
	// If no allowed values are configured, accept any non-empty value
	if len(v.allowedValues) == 0 {
		return strings.TrimSpace(value) != ""
	}
	return slices.Contains(v.allowedValues, value)
}
