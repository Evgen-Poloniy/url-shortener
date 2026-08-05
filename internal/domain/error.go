package domain

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    ErrCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code ErrCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrMissingAuthHeader = errors.New("missing authentication header")
	ErrInvalidAPIKey     = errors.New("invalid API-Key")
	ErrURLNotFound       = errors.New("URL not found")
	ErrURLConflict       = errors.New("url already exists")
	ErrInternalStorage   = errors.New("internal storage error")
	ErrEmptyURL          = errors.New("url field is empty")
)
