package common

import "net/http"

// AppError — lỗi nghiệp vụ trả về theo format {error:{code,message,field?,...extra}}.
type AppError struct {
	StatusCode int            `json:"-"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Field      string         `json:"field,omitempty"`
	Extra      map[string]any `json:"-"`
}

func (e *AppError) Error() string { return e.Code + ": " + e.Message }

func NewAppError(status int, code, message string) *AppError {
	return &AppError{StatusCode: status, Code: code, Message: message}
}

func (e *AppError) WithField(field string) *AppError {
	e.Field = field
	return e
}

func (e *AppError) WithExtra(extra map[string]any) *AppError {
	e.Extra = extra
	return e
}

func NewUnprocessable(code, message, field string) *AppError {
	return &AppError{StatusCode: http.StatusUnprocessableEntity, Code: code, Message: message, Field: field}
}

func NewUnauthorized(code, message string) *AppError {
	return NewAppError(http.StatusUnauthorized, code, message)
}

func NewNotFound(message string) *AppError {
	return NewAppError(http.StatusNotFound, ErrCodeRecordGone, message)
}

func NewConflict(message string) *AppError {
	return NewAppError(http.StatusConflict, ErrCodeConcurrencyConflict, message)
}

func NewBadRequest(message string) *AppError {
	return NewAppError(http.StatusBadRequest, ErrCodeInvalidRequest, message)
}

func NewInternal(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, ErrCodeInternal, "đã có lỗi xảy ra, vui lòng thử lại")
}
