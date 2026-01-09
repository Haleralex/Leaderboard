package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// StringPtr возвращает указатель на строку (переиспользуемая утилита)
func StringPtr(s string) *string {
	return &s
}

// IntPtr возвращает указатель на int
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr возвращает указатель на int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr возвращает указатель на bool
func BoolPtr(b bool) *bool {
	return &b
}

// StringValue безопасно извлекает значение из указателя на строку
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// IntValue безопасно извлекает значение из указателя на int
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// Int64Value безопасно извлекает значение из указателя на int64
func Int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// BoolValue безопасно извлекает значение из указателя на bool
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// TruncateString обрезает строку до указанной длины с добавлением многоточия
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// NormalizeEmail нормализует email (lowercase, trim)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeUsername нормализует username (lowercase, trim)
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// GenerateRandomString генерирует случайную строку указанной длины
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// Contains проверяет, содержит ли срез строку
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsInt проверяет, содержит ли срез число
func ContainsInt(slice []int, item int) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates удаляет дубликаты из среза строк
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// Coalesce возвращает первое непустое значение
func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// DefaultString возвращает значение или default если пусто
func DefaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// DefaultInt возвращает значение или default если 0
func DefaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// DefaultInt64 возвращает значение или default если 0
func DefaultInt64(value, defaultValue int64) int64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

// Clamp ограничивает значение в диапазоне
func Clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampInt ограничивает значение int в диапазоне
func ClampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// IsZeroValue проверяет, является ли значение нулевым для разных типов
func IsZeroValue[T comparable](value T) bool {
	var zero T
	return value == zero
}
