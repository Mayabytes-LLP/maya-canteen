package errors

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Common error types
var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal error")
	ErrAlreadyExists = errors.New("resource already exists")
)

// AppError represents an application error with context
type AppError struct {
	Err     error
	Message string
	Code    string
}

// Error returns the error message
func (e *AppError) Error() string {
	log.Error(e.Err)
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	log.Error(e.Err)
	return e.Err
}

// New creates a new AppError
func New(err error, message string, code string) *AppError {
	log.Error(err)
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// Newf creates a new AppError with formatted message
func Newf(err error, code string, format string, args ...interface{}) *AppError {
	log.Error(err)
	return &AppError{
		Err:     err,
		Message: fmt.Sprintf(format, args...),
		Code:    code,
	}
}

// NotFound creates a new not found error
func NotFound(resource string, id interface{}) *AppError {
	log.Error(ErrNotFound)
	return &AppError{
		Err:     ErrNotFound,
		Message: fmt.Sprintf("%s with ID %v not found", resource, id),
		Code:    "NOT_FOUND",
	}
}

// InvalidInput creates a new invalid input error
func InvalidInput(message string) *AppError {
	log.Error(ErrInvalidInput)
	return &AppError{
		Err:     ErrInvalidInput,
		Message: message,
		Code:    "INVALID_INPUT",
	}
}

// Internal creates a new internal error
func Internal(err error) *AppError {
	log.Error(err)
	if err != nil {
		return &AppError{
			Err:     err,
			Message: fmt.Sprintf("An internal error occurred: %v", err),
			Code:    "INTERNAL_ERROR",
		}
	}
	return &AppError{
		Err:     ErrInternal,
		Message: "An internal error occurred",
		Code:    "INTERNAL_ERROR",
	}
}

// Is checks if the error is of the given type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As sets the target to the error value if it is of the same type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
