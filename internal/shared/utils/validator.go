package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// FieldError представляет ошибку валидации поля
type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// FieldErrors коллекция ошибок валидации
type FieldErrors []FieldError

func (errs FieldErrors) Error() string {
	if len(errs) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, err := range errs {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// HasErrors проверяет наличие ошибок
func (errs FieldErrors) HasErrors() bool {
	return len(errs) > 0
}

// Validator переиспользуемый валидатор с цепочкой проверок
type Validator struct {
	errors FieldErrors
}

// NewValidator создает новый валидатор
func NewValidator() *Validator {
	return &Validator{
		errors: make(FieldErrors, 0),
	}
}

// Required проверяет, что строка не пустая
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "is required",
		})
	}
	return v
}

// MinLength проверяет минимальную длину строки
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", min),
		})
	}
	return v
}

// MaxLength проверяет максимальную длину строки
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if len(value) > max {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: fmt.Sprintf("must be at most %d characters", max),
		})
	}
	return v
}

// Email проверяет формат email
func (v *Validator) Email(field, value string) *Validator {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "must be a valid email address",
		})
	}
	return v
}

// Range проверяет, что число в диапазоне
func (v *Validator) Range(field string, value int64, min, max int64) *Validator {
	if value < min || value > max {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: fmt.Sprintf("must be between %d and %d", min, max),
		})
	}
	return v
}

// Min проверяет минимальное значение
func (v *Validator) Min(field string, value int64, min int64) *Validator {
	if value < min {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d", min),
		})
	}
	return v
}

// Max проверяет максимальное значение
func (v *Validator) Max(field string, value int64, max int64) *Validator {
	if value > max {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: fmt.Sprintf("must be at most %d", max),
		})
	}
	return v
}

// Pattern проверяет соответствие регулярному выражению
func (v *Validator) Pattern(field, value, pattern, message string) *Validator {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: message,
		})
	}
	return v
}

// Username проверяет формат username (буквы, цифры, _, -)
func (v *Validator) Username(field, value string) *Validator {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	if !usernameRegex.MatchString(value) {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "must contain only letters, numbers, underscores and hyphens",
		})
	}
	return v
}

// Password проверяет сложность пароля
func (v *Validator) Password(field, value string, minLength int) *Validator {
	if len(value) < minLength {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters long", minLength),
		})
		return v
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "must contain at least one uppercase letter",
		})
	}
	if !hasLower {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "must contain at least one lowercase letter",
		})
	}
	if !hasDigit {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "must contain at least one digit",
		})
	}

	return v
}

// Season проверяет формат сезона (YYYY-season-name)
func (v *Validator) Season(field, value string) *Validator {
	seasonRegex := regexp.MustCompile(`^\d{4}-(spring|summer|fall|winter|q[1-4]|[a-z0-9_\-]+)$`)
	if !seasonRegex.MatchString(value) {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: "must be in format YYYY-season-name (e.g., 2024-spring, 2024-q1)",
		})
	}
	return v
}

// OneOf проверяет, что значение есть в списке допустимых
func (v *Validator) OneOf(field, value string, allowed []string) *Validator {
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.errors = append(v.errors, FieldError{
		Field:   field,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
	})
	return v
}

// Custom добавляет кастомное условие валидации
func (v *Validator) Custom(field string, condition bool, message string) *Validator {
	if !condition {
		v.errors = append(v.errors, FieldError{
			Field:   field,
			Message: message,
		})
	}
	return v
}

// IsValid проверяет, прошла ли валидация
func (v *Validator) IsValid() bool {
	return !v.errors.HasErrors()
}

// Errors возвращает все ошибки валидации
func (v *Validator) Errors() FieldErrors {
	return v.errors
}

// Error возвращает первую ошибку или nil
func (v *Validator) Error() error {
	if !v.errors.HasErrors() {
		return nil
	}
	return v.errors
}
