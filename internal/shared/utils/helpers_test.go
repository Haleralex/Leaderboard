package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPtr(t *testing.T) {
	str := "test"
	ptr := StringPtr(str)
	assert.NotNil(t, ptr)
	assert.Equal(t, "test", *ptr)
}

func TestIntPtr(t *testing.T) {
	num := 42
	ptr := IntPtr(num)
	assert.NotNil(t, ptr)
	assert.Equal(t, 42, *ptr)
}

func TestInt64Ptr(t *testing.T) {
	num := int64(12345)
	ptr := Int64Ptr(num)
	assert.NotNil(t, ptr)
	assert.Equal(t, int64(12345), *ptr)
}

func TestBoolPtr(t *testing.T) {
	b := true
	ptr := BoolPtr(b)
	assert.NotNil(t, ptr)
	assert.Equal(t, true, *ptr)
}

func TestStringValue(t *testing.T) {
	t.Run("valid pointer", func(t *testing.T) {
		str := "test"
		assert.Equal(t, "test", StringValue(&str))
	})

	t.Run("nil pointer", func(t *testing.T) {
		assert.Equal(t, "", StringValue(nil))
	})
}

func TestIntValue(t *testing.T) {
	t.Run("valid pointer", func(t *testing.T) {
		num := 42
		assert.Equal(t, 42, IntValue(&num))
	})

	t.Run("nil pointer", func(t *testing.T) {
		assert.Equal(t, 0, IntValue(nil))
	})
}

func TestInt64Value(t *testing.T) {
	t.Run("valid pointer", func(t *testing.T) {
		num := int64(12345)
		assert.Equal(t, int64(12345), Int64Value(&num))
	})

	t.Run("nil pointer", func(t *testing.T) {
		assert.Equal(t, int64(0), Int64Value(nil))
	})
}

func TestBoolValue(t *testing.T) {
	t.Run("valid pointer", func(t *testing.T) {
		b := true
		assert.Equal(t, true, BoolValue(&b))
	})

	t.Run("nil pointer", func(t *testing.T) {
		assert.Equal(t, false, BoolValue(nil))
	})
}

func TestTruncateString(t *testing.T) {
	t.Run("string shorter than max", func(t *testing.T) {
		assert.Equal(t, "short", TruncateString("short", 10))
	})

	t.Run("string longer than max", func(t *testing.T) {
		assert.Equal(t, "this is...", TruncateString("this is a long string", 10))
	})

	t.Run("maxLen less than 3", func(t *testing.T) {
		assert.Equal(t, "te", TruncateString("test", 2))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", TruncateString("", 10))
	})
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test@Example.COM", "test@example.com"},
		{" user@domain.com ", "user@domain.com"},
		{"CAPS@DOMAIN.COM", "caps@domain.com"},
		{"normal@test.com", "normal@test.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeEmail(tt.input))
		})
	}
}

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestUser", "testuser"},
		{" username ", "username"},
		{"CAPS", "caps"},
		{"normal", "normal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeUsername(tt.input))
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	t.Run("generates string of correct length", func(t *testing.T) {
		str, err := GenerateRandomString(10)
		assert.NoError(t, err)
		assert.Len(t, str, 10)
	})

	t.Run("generates different strings", func(t *testing.T) {
		str1, err1 := GenerateRandomString(20)
		str2, err2 := GenerateRandomString(20)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, str1, str2)
	})

	t.Run("zero length", func(t *testing.T) {
		str, err := GenerateRandomString(0)
		assert.NoError(t, err)
		assert.Empty(t, str)
	})
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	t.Run("item exists", func(t *testing.T) {
		assert.True(t, Contains(slice, "banana"))
	})

	t.Run("item does not exist", func(t *testing.T) {
		assert.False(t, Contains(slice, "grape"))
	})

	t.Run("empty slice", func(t *testing.T) {
		assert.False(t, Contains([]string{}, "test"))
	})
}
