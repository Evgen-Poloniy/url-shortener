package domain

type ErrCode string

const (
	CodeNotFound           ErrCode = "NOT_FOUND"
	CodeMethodNotAllowed   ErrCode = "METHOD_NOT_ALLOWED"
	CodeMissingAuthHeaders ErrCode = "MISSING_AUTH_HEADER"
	CodeInvalidAPIKey      ErrCode = "INVALID_API_KEY"
	CodeValidationError    ErrCode = "VALIDATION_ERROR"
	CodeURLNotFound        ErrCode = "URL_NOT_FOUND"
	CodeURLConflict        ErrCode = "URL_ALREADY_EXIST"
	CodeInternal           ErrCode = "INTERNAL_ERROR"
)
