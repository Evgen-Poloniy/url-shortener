package middleware

import (
	"errors"
	"net/http"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"

	"github.com/gin-gonic/gin"
)

// ResponseError represents DTO with error.
type ResponseError struct {
	Error Error `json:"error"`
}

// Error represents details about error with string code and message for client.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var httpStatusMap = map[domain.ErrCode]int{
	domain.CodeNotFound:           http.StatusNotFound,
	domain.CodeMethodNotAllowed:   http.StatusMethodNotAllowed,
	domain.CodeMissingAuthHeaders: http.StatusUnauthorized,
	domain.CodeInvalidAPIKey:      http.StatusUnauthorized,
	domain.CodeValidationError:    http.StatusBadRequest,
	domain.CodeURLNotFound:        http.StatusNotFound,
	domain.CodeURLConflict:        http.StatusConflict,
	domain.CodeInternal:           http.StatusInternalServerError,
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		var statusCode int
		var code string
		var message string
		err := c.Errors.Last().Err

		// Error mapping
		if appError, ok := errors.AsType[*domain.AppError](err); ok {
			code = string(appError.Code)

			var ok bool
			if statusCode, ok = httpStatusMap[appError.Code]; !ok {
				statusCode = http.StatusInternalServerError
			}

			message = appError.Message
		} else {
			statusCode = http.StatusInternalServerError
			code = "UNKNOWN_ERROR"
			message = "unknown error"
		}

		c.Status(statusCode)
		c.Set("code", code)

		c.JSON(statusCode, ResponseError{
			Error: Error{
				Code:    code,
				Message: message,
			},
		})
	}
}
