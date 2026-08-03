package utils

// Bool returns a pointer to the bool value passed in.
//
//go:fix inline
func Bool(v bool) *bool {
	return new(v)
}

// BoolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}
