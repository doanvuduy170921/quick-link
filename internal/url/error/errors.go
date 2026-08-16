package error

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type URLError struct {
	Type    ErrorType
	Message string
	Err     error // save root error
}

type ErrorType string

const (
	ErrNotFound   ErrorType = "not_found"
	ErrValidation ErrorType = "validation_error"
	ErrInternal   ErrorType = "internal_error"
)

func (e *URLError) Error() string {
	return e.Message
}

func (e *URLError) Unwrap() error {
	return e.Err
}

// NewNotFoundError Constructor functions
func NewNotFoundError(message string, err error) *URLError {
	return &URLError{
		Type:    ErrNotFound,
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message string) *URLError {
	return &URLError{
		Type:    ErrValidation,
		Message: message,
		Err:     nil,
	}
}

func NewInternalError(message string, err error) *URLError {
	return &URLError{
		Type:    ErrInternal,
		Message: message,
		Err:     err,
	}
}

func HandlerError(c *gin.Context, err error) {
	var urlErr *URLError
	if errors.As(err, &urlErr) {
		switch urlErr.Type {
		case ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": urlErr.Message})
			return
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": urlErr.Message})
			return
		case ErrInternal:
			log.Printf("internal error: %s: %v", urlErr.Message, urlErr.Err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		default:
			log.Printf("unknown error type: %s: %v", urlErr.Message, urlErr.Err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	} else {
		log.Printf("unhandled error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
}
