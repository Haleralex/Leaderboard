package utils

import (
	"fmt"
	"net/http"
)

// AppError переиспользуемая структура для доменных ошибок
type AppError struct {
	Code       string // Код ошибки для клиента
	Message    string // Сообщение для пользователя
	StatusCode int    // HTTP статус код
	Err        error  // Внутренняя ошибка
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap для поддержки errors.Is и errors.As
func (e *AppError) Unwrap() error {
	return e.Err
}

// Предопределенные коды ошибок для переиспользования
const (
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeCacheError         = "CACHE_ERROR"
	ErrCodeNotImplemented     = "NOT_IMPLEMENTED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// NewAppError создает новую доменную ошибку
func NewAppError(code, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Переиспользуемые конструкторы для типовых ошибок

// NotFound создает ошибку "не найдено"
func NotFound(resource string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

// BadRequest создает ошибку "некорректный запрос"
func BadRequest(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// Unauthorized создает ошибку "не авторизован"
func Unauthorized(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// Forbidden создает ошибку "доступ запрещен"
func Forbidden(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Err:        err,
	}
}

// Conflict создает ошибку "конфликт"
func Conflict(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
		Err:        err,
	}
}

// InternalError создает ошибку "внутренняя ошибка сервера"
func InternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// ValidationError создает ошибку валидации
func ValidationError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// DatabaseError создает ошибку базы данных
func DatabaseError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeDatabaseError,
		Message:    fmt.Sprintf("database %s failed", operation),
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// CacheError создает ошибку кэша
func CacheError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeCacheError,
		Message:    fmt.Sprintf("cache %s failed", operation),
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NotImplemented создает ошибку "не реализовано"
func NotImplemented(feature string) *AppError {
	return &AppError{
		Code:       ErrCodeNotImplemented,
		Message:    fmt.Sprintf("%s is not implemented yet", feature),
		StatusCode: http.StatusNotImplemented,
		Err:        nil,
	}
}

// ServiceUnavailable создает ошибку "сервис недоступен"
func ServiceUnavailable(service string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeServiceUnavailable,
		Message:    fmt.Sprintf("%s service is unavailable", service),
		StatusCode: http.StatusServiceUnavailable,
		Err:        err,
	}
}

// ErrorResponse переиспользуемая структура для HTTP ответов с ошибками
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToErrorResponse конвертирует AppError в ErrorResponse для API
func (e *AppError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
	}
}

// WithDetails добавляет детали к ошибке для API ответа
func (e *AppError) WithDetails(details map[string]interface{}) ErrorResponse {
	resp := e.ToErrorResponse()
	resp.Details = details
	return resp
}

// WrapError оборачивает стандартную ошибку в AppError
func WrapError(err error, code, message string, statusCode int) *AppError {
	if err == nil {
		return nil
	}
	return NewAppError(code, message, statusCode, err)
}
