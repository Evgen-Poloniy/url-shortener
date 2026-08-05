package domain

type ErrCode string

const (
	CodeNotFound           ErrCode = "NOT_FOUND"
	CodeMethodNotAllowed   ErrCode = "METHOD_NOT_ALLOWED"
	CodeMissingAuthHeaders ErrCode = "MISSING_AUTH_HEADER"
	CodeInvalidAPIKey      ErrCode = "INVALID_API_KEY" //nolint:gosec
	CodeValidationError    ErrCode = "VALIDATION_ERROR"
	CodeURLNotFound        ErrCode = "URL_NOT_FOUND"
	CodeURLConflict        ErrCode = "URL_ALREADY_EXIST"
	CodeInternalStorage    ErrCode = "INTERNAL_STORAGE_ERROR"
	CodeInternal           ErrCode = "INTERNAL_ERROR"
	CodeInvalidInput       ErrCode = "INVALID_INPUT"
)
