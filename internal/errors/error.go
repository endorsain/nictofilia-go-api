// Package errors provides a common representation for application errors,
// independent from any specific domain (auth, users, etc.) or infrastructure
// (PostgreSQL, JWT, filesystem, ...).
//
// It separates:
//   - the public code/message that is safe to expose to API clients;
//   - the original error, preserved via Go error wrapping for logging and
//     errors.Is / errors.As support.
package errors

import (
	"errors"
	"net/http"
)

// Code is a stable, machine-readable identifier for an application error.
// It is meant to be used programmatically by API clients.
type Code string

// Generic codes used as a fallback when a more specific domain code
// has not been defined yet.
const (
	CodeInternal        Code = "INTERNAL_ERROR"
	CodeInvalidInput    Code = "INVALID_INPUT"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeValidationError Code = "VALIDATION_ERROR"
)

// AppError is the internal representation of an application error.
//
// Code and Message are safe to expose to clients. Err keeps the original
// error (SQL error, JWT error, etc.) for logging/debugging purposes only;
// it must never be rendered directly in an HTTP response.
type AppError struct {
	Code       Code
	Message    string
	HTTPStatus int
	Err        error
}

// New creates an AppError without an underlying wrapped error.
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap creates an AppError that preserves err for logging/errors.Is/errors.As,
// while exposing only code/message/httpStatus publicly.
func Wrap(err error, code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Internal wraps err as an internal error. It is intended for unexpected
// failures (PostgreSQL, filesystem, JWT, etc.) that must never leak details
// to the client.
func Internal(err error) *AppError {
	return Wrap(err, CodeInternal, "internal server error", http.StatusInternalServerError)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap exposes the original error, enabling errors.Is and errors.As
// to traverse the chain past this AppError.
func (e *AppError) Unwrap() error {
	return e.Err
}

// As is a convenience helper to extract an *AppError from an error chain.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
