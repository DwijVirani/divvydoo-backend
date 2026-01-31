// Custom error types
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(statusCode int, message string) *CustomError {
	return &CustomError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func RespondWithError(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{"error": message})
}

func GetStatusCode(err error) int {
	switch err := err.(type) {
	case *CustomError:
		return err.StatusCode
	default:
		return http.StatusInternalServerError
	}
}
