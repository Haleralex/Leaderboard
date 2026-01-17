package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_Required(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"Valid value", "test", false},
		{"Empty string", "", true},
		{"Only spaces", "   ", true},
		{"With content", "  test  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Required("field", tt.value)

			if tt.wantError {
				assert.True(t, v.errors.HasErrors(), "Expected validation error")
			} else {
				assert.False(t, v.errors.HasErrors(), "Expected no validation error")
			}
		})
	}
}

func TestValidator_MinLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		min       int
		wantError bool
	}{
		{"Valid length", "test", 3, false},
		{"Exact length", "test", 4, false},
		{"Too short", "te", 3, true},
		{"Empty", "", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.MinLength("field", tt.value, tt.min)

			if tt.wantError {
				assert.True(t, v.errors.HasErrors())
			} else {
				assert.False(t, v.errors.HasErrors())
			}
		})
	}
}

func TestValidator_MaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		max       int
		wantError bool
	}{
		{"Valid length", "test", 5, false},
		{"Exact length", "test", 4, false},
		{"Too long", "testing", 5, true},
		{"Empty", "", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.MaxLength("field", tt.value, tt.max)

			if tt.wantError {
				assert.True(t, v.errors.HasErrors())
			} else {
				assert.False(t, v.errors.HasErrors())
			}
		})
	}
}

func TestValidator_Email(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantError bool
	}{
		{"Valid email", "user@example.com", false},
		{"Valid with subdomain", "user@mail.example.com", false},
		{"Valid with plus", "user+tag@example.com", false},
		{"Invalid no @", "userexample.com", true},
		{"Invalid no domain", "user@", true},
		{"Invalid no TLD", "user@example", true},
		{"Invalid empty", "", true},
		{"Invalid spaces", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Email("email", tt.email)

			if tt.wantError {
				assert.True(t, v.errors.HasErrors())
			} else {
				assert.False(t, v.errors.HasErrors())
			}
		})
	}
}

func TestValidator_Chaining(t *testing.T) {
	v := NewValidator()
	v.Required("name", "John").
		MinLength("name", "John", 3).
		MaxLength("name", "John", 10)

	assert.False(t, v.errors.HasErrors(), "All validations should pass")
}

func TestValidator_MultipleErrors(t *testing.T) {
	v := NewValidator()
	v.Required("name", "").
		Email("email", "invalid").
		MinLength("password", "12", 6)

	assert.True(t, v.errors.HasErrors())
	assert.Equal(t, 3, len(v.errors), "Should have 3 errors")
}

func TestValidator_GetErrors(t *testing.T) {
	v := NewValidator()
	v.Required("name", "").
		Email("email", "invalid")

	errs := v.Errors()
	assert.Equal(t, 2, len(errs))
	assert.Contains(t, errs.Error(), "name")
	assert.Contains(t, errs.Error(), "email")
}

func TestFieldErrors_Error(t *testing.T) {
	errs := FieldErrors{
		{Field: "name", Message: "is required"},
		{Field: "email", Message: "must be valid"},
	}

	errStr := errs.Error()
	assert.Contains(t, errStr, "validation failed")
	assert.Contains(t, errStr, "name: is required")
	assert.Contains(t, errStr, "email: must be valid")
}

func TestFieldErrors_HasErrors(t *testing.T) {
	emptyErrs := FieldErrors{}
	assert.False(t, emptyErrs.HasErrors())

	withErrs := FieldErrors{{Field: "test", Message: "error"}}
	assert.True(t, withErrs.HasErrors())
}
