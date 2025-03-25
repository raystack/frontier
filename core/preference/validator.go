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
