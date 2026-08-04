package domain

import "errors"

type AppError struct {
	Code    ErrCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
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
)
