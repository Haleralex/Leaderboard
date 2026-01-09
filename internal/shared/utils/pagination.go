package utils

import (
	"math"
	"strconv"
)

// PaginationParams переиспользуемые параметры пагинации
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
	Limit    int
}

// PaginationMeta переиспользуемые метаданные пагинации для API ответов
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalCount int64 `json:"total_count"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResponse переиспользуемая обертка для пагинированных ответов
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// NewPaginationParams создает параметры пагинации с валидацией
func NewPaginationParams(page, pageSize int) *PaginationParams {
	// Валидация и установка значений по умолчанию
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // Максимум 100 записей на страницу
	}

	offset := (page - 1) * pageSize

	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
		Limit:    pageSize,
	}
}

// ParsePaginationParams парсит параметры пагинации из строк (query params)
func ParsePaginationParams(pageStr, pageSizeStr string) *PaginationParams {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	return NewPaginationParams(page, pageSize)
}

// NewPaginationMeta создает метаданные пагинации
func NewPaginationMeta(page, pageSize int, totalCount int64) PaginationMeta {
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// NewPaginatedResponse создает пагинированный ответ
func NewPaginatedResponse[T any](data []T, params *PaginationParams, totalCount int64) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data:       data,
		Pagination: NewPaginationMeta(params.Page, params.PageSize, totalCount),
	}
}

// CursorParams параметры для cursor-based пагинации (альтернативный подход)
type CursorParams struct {
	Cursor   string
	PageSize int
	Forward  bool // true для вперед, false для назад
}

// CursorMeta метаданные для cursor-based пагинации
type CursorMeta struct {
	NextCursor     *string `json:"next_cursor,omitempty"`
	PreviousCursor *string `json:"previous_cursor,omitempty"`
	HasMore        bool    `json:"has_more"`
	PageSize       int     `json:"page_size"`
}

// CursorPaginatedResponse ответ с cursor-based пагинацией
type CursorPaginatedResponse[T any] struct {
	Data   []T        `json:"data"`
	Cursor CursorMeta `json:"cursor"`
}

// NewCursorParams создает параметры cursor-пагинации
func NewCursorParams(cursor string, pageSize int) *CursorParams {
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return &CursorParams{
		Cursor:   cursor,
		PageSize: pageSize,
		Forward:  true,
	}
}

// ParseCursorParams парсит параметры cursor-пагинации
func ParseCursorParams(cursor, pageSizeStr string) *CursorParams {
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	return NewCursorParams(cursor, pageSize)
}
