package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaginationParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := NewPaginationParams(2, 10)
		assert.Equal(t, 2, params.Page)
		assert.Equal(t, 10, params.PageSize)
		assert.Equal(t, 10, params.Offset) // (2-1) * 10
		assert.Equal(t, 10, params.Limit)
	})

	t.Run("page less than 1", func(t *testing.T) {
		params := NewPaginationParams(0, 10)
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 0, params.Offset)
	})

	t.Run("pageSize less than 1", func(t *testing.T) {
		params := NewPaginationParams(1, 0)
		assert.Equal(t, 20, params.PageSize) // default
	})

	t.Run("pageSize exceeds max", func(t *testing.T) {
		params := NewPaginationParams(1, 150)
		assert.Equal(t, 100, params.PageSize) // capped at 100
	})

	t.Run("first page", func(t *testing.T) {
		params := NewPaginationParams(1, 20)
		assert.Equal(t, 0, params.Offset)
	})

	t.Run("third page", func(t *testing.T) {
		params := NewPaginationParams(3, 25)
		assert.Equal(t, 50, params.Offset) // (3-1) * 25
	})
}

func TestParsePaginationParams(t *testing.T) {
	t.Run("valid strings", func(t *testing.T) {
		params := ParsePaginationParams("2", "15")
		assert.Equal(t, 2, params.Page)
		assert.Equal(t, 15, params.PageSize)
	})

	t.Run("invalid page string", func(t *testing.T) {
		params := ParsePaginationParams("invalid", "10")
		assert.Equal(t, 1, params.Page) // default
		assert.Equal(t, 10, params.PageSize)
	})

	t.Run("invalid pageSize string", func(t *testing.T) {
		params := ParsePaginationParams("1", "invalid")
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.PageSize) // default
	})

	t.Run("empty strings", func(t *testing.T) {
		params := ParsePaginationParams("", "")
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.PageSize)
	})

	t.Run("negative values", func(t *testing.T) {
		params := ParsePaginationParams("-1", "-5")
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.PageSize)
	})
}

func TestNewPaginationMeta(t *testing.T) {
	t.Run("basic meta", func(t *testing.T) {
		meta := NewPaginationMeta(1, 20, 100)
		assert.Equal(t, 1, meta.Page)
		assert.Equal(t, 20, meta.PageSize)
		assert.Equal(t, 5, meta.TotalPages) // ceil(100/20)
		assert.Equal(t, int64(100), meta.TotalCount)
		assert.True(t, meta.HasNext)
		assert.False(t, meta.HasPrev)
	})

	t.Run("last page", func(t *testing.T) {
		meta := NewPaginationMeta(5, 20, 100)
		assert.False(t, meta.HasNext)
		assert.True(t, meta.HasPrev)
	})

	t.Run("middle page", func(t *testing.T) {
		meta := NewPaginationMeta(3, 20, 100)
		assert.True(t, meta.HasNext)
		assert.True(t, meta.HasPrev)
	})

	t.Run("single page", func(t *testing.T) {
		meta := NewPaginationMeta(1, 20, 10)
		assert.Equal(t, 1, meta.TotalPages)
		assert.False(t, meta.HasNext)
		assert.False(t, meta.HasPrev)
	})

	t.Run("empty results", func(t *testing.T) {
		meta := NewPaginationMeta(1, 20, 0)
		assert.Equal(t, 0, meta.TotalPages)
		assert.False(t, meta.HasNext)
		assert.False(t, meta.HasPrev)
	})

	t.Run("odd division", func(t *testing.T) {
		meta := NewPaginationMeta(1, 20, 55)
		assert.Equal(t, 3, meta.TotalPages) // ceil(55/20) = 3
	})
}

func TestNewPaginatedResponse(t *testing.T) {
	type TestData struct {
		ID   int
		Name string
	}

	t.Run("with data", func(t *testing.T) {
		data := []TestData{
			{ID: 1, Name: "Item 1"},
			{ID: 2, Name: "Item 2"},
		}
		params := NewPaginationParams(1, 20)
		response := NewPaginatedResponse(data, params, 100)

		assert.Len(t, response.Data, 2)
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 20, response.Pagination.PageSize)
		assert.Equal(t, int64(100), response.Pagination.TotalCount)
		assert.Equal(t, 5, response.Pagination.TotalPages)
	})

	t.Run("empty data", func(t *testing.T) {
		data := []TestData{}
		params := NewPaginationParams(1, 20)
		response := NewPaginatedResponse(data, params, 0)

		assert.Empty(t, response.Data)
		assert.Equal(t, 0, response.Pagination.TotalPages)
	})

	t.Run("last page partial", func(t *testing.T) {
		data := []TestData{
			{ID: 91, Name: "Item 91"},
			{ID: 92, Name: "Item 92"},
		}
		params := NewPaginationParams(5, 20) // Page 5 of 92 items
		response := NewPaginatedResponse(data, params, 92)

		assert.Len(t, response.Data, 2)
		assert.Equal(t, 5, response.Pagination.TotalPages)
		assert.False(t, response.Pagination.HasNext)
	})
}

func TestPaginationEdgeCases(t *testing.T) {
	t.Run("page size of 1", func(t *testing.T) {
		params := NewPaginationParams(1, 1)
		meta := NewPaginationMeta(params.Page, params.PageSize, 100)
		assert.Equal(t, 100, meta.TotalPages)
	})

	t.Run("very large page number", func(t *testing.T) {
		params := NewPaginationParams(1000, 20)
		assert.Equal(t, 1000, params.Page)
		assert.Equal(t, 19980, params.Offset)
	})

	t.Run("total count less than page size", func(t *testing.T) {
		meta := NewPaginationMeta(1, 20, 5)
		assert.Equal(t, 1, meta.TotalPages)
		assert.False(t, meta.HasNext)
	})
}
