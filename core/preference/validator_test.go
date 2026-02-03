package preference

import (
	"testing"
)

func TestBooleanValidator_Validate(t *testing.T) {
	v := NewBooleanValidator()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"valid true", "true", true},
		{"valid false", "false", true},
		{"invalid yes", "yes", false},
		{"invalid no", "no", false},
		{"invalid empty", "", false},
		{"invalid number", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Validate(tt.value); got != tt.want {
				t.Errorf("BooleanValidator.Validate(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestTextValidator_Validate(t *testing.T) {
	v := NewTextValidator()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"valid text", "hello", true},
		{"valid text with spaces", "hello world", true},
		{"invalid empty", "", false},
		{"invalid whitespace only", "   ", false},
		{"valid single char", "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Validate(tt.value); got != tt.want {
				t.Errorf("TextValidator.Validate(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestSelectValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		inputHints string
		value      string
		want       bool
	}{
		{"valid option first", "option1,option2,option3", "option1", true},
		{"valid option middle", "option1,option2,option3", "option2", true},
		{"valid option last", "option1,option2,option3", "option3", true},
		{"invalid option", "option1,option2,option3", "option4", false},
		{"invalid empty value", "option1,option2", "", false},
		{"valid with spaces in hints", "option1, option2, option3", "option2", true},
		{"empty hints accepts non-empty", "", "anyvalue", true},
		{"empty hints rejects empty", "", "", false},
		{"case sensitive", "Option1,Option2", "option1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSelectValidator(tt.inputHints)
			if got := v.Validate(tt.value); got != tt.want {
				t.Errorf("SelectValidator.Validate(%q) with hints %q = %v, want %v",
					tt.value, tt.inputHints, got, tt.want)
			}
		})
	}
}

func TestTrait_GetValidator(t *testing.T) {
	tests := []struct {
		name       string
		input      TraitInput
		inputHints string
		testValue  string
		want       bool
	}{
		{"checkbox true", TraitInputCheckbox, "", "true", true},
		{"checkbox invalid", TraitInputCheckbox, "", "yes", false},
		{"text valid", TraitInputText, "", "hello", true},
		{"text empty", TraitInputText, "", "", false},
		{"textarea valid", TraitInputTextarea, "", "hello", true},
		{"select valid", TraitInputSelect, "a,b,c", "b", true},
		{"select invalid", TraitInputSelect, "a,b,c", "d", false},
		{"combobox valid", TraitInputCombobox, "x,y,z", "y", true},
		{"combobox invalid", TraitInputCombobox, "x,y,z", "w", false},
		{"multiselect falls back to text", TraitInputMultiselect, "1,2,3", "1,3", true},
		{"number falls back to text", TraitInputNumber, "", "123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trait := Trait{
				Input:      tt.input,
				InputHints: tt.inputHints,
			}
			v := trait.GetValidator()
			if got := v.Validate(tt.testValue); got != tt.want {
				t.Errorf("Trait.GetValidator().Validate(%q) for input %q = %v, want %v",
					tt.testValue, tt.input, got, tt.want)
			}
		})
	}
}
