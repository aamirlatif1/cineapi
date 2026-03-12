// Package validator provides a simple field-level error accumulator for input validation.
package validator

import (
	"regexp"
	"slices"
)

// EmailRX is the regular expression used to validate email addresses.
var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Validator accumulates field-level validation errors.
type Validator struct {
	errors map[string]string
}

// New returns a fresh Validator with no errors.
func New() *Validator {
	return &Validator{errors: make(map[string]string)}
}

// Valid reports whether no validation errors have been accumulated.
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Check adds an error message for key if ok is false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// AddError adds an error message for the given key, provided no error already exists for that key.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.errors[key]; !exists {
		v.errors[key] = message
	}
}

// Errors returns the map of accumulated field errors.
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// Matches reports whether a string value matches a compiled regular expression pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique reports whether all values in a slice are distinct.
func Unique[T comparable](values []T) bool {
	return len(slices.Compact(append([]T(nil), values...))) == len(values)
}

// In reports whether a value is contained in the provided list.
func In[T comparable](value T, list ...T) bool {
	return slices.Contains(list, value)
}
