package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

/*
ErrorKind classifies an application error and defines
how it maps to HTTP semantics.
*/
type ErrorKind struct {
	name       string
	httpStatus int
}

func (k ErrorKind) Name() string   { return k.name }
func (k ErrorKind) Status() int    { return k.httpStatus }
func (k ErrorKind) String() string { return k.name }

/*
AppError is a structured application error that:
- preserves a human message
- retains the root cause
- exposes HTTP semantics via Kind
*/
type AppError struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

func New(kind ErrorKind, message string) *AppError {
	return &AppError{Kind: kind, Message: message}
}

func Wrap(kind ErrorKind, message string, cause error) *AppError {
	return &AppError{Kind: kind, Message: message, Cause: cause}
}

// Error formats the application error in a structured, log-friendly way.
// Example:
// kind=BusinessLogic msg="insufficient balance" cause="pq: deadlock detected"
func (e *AppError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf(`kind=%s msg="%s"`, e.Kind, e.Message)
	}
	return fmt.Sprintf(
		`kind=%s msg="%s" cause="%v"`,
		e.Kind,
		e.Message,
		e.Cause,
	)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

/*
Helpers for ergonomic checks in handlers and services
*/
func Is(err error, kind ErrorKind) bool {
	var ae *AppError
	return errors.As(err, &ae) && ae.Kind == kind
}

func Status(err error) int {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Kind.Status()
	}
	return http.StatusInternalServerError
}

func Message(err error) string {
	var ae *AppError
	if errors.As(err, &ae) && ae.Message != "" {
		return ae.Message
	}
	return "internal server error"
}

/*
Predefined error kinds
*/
var (
	ErrValidation    = ErrorKind{"Validation", http.StatusBadRequest}
	ErrUnauthorized  = ErrorKind{"Unauthorized", http.StatusUnauthorized}
	ErrForbidden     = ErrorKind{"Forbidden", http.StatusForbidden}
	ErrNotFound      = ErrorKind{"NotFound", http.StatusNotFound}
	ErrConflict      = ErrorKind{"Conflict", http.StatusConflict}
	ErrBusinessLogic = ErrorKind{"BusinessLogic", http.StatusUnprocessableEntity}
	ErrInternal      = ErrorKind{"Internal", http.StatusInternalServerError}
	ErrUpstream      = ErrorKind{"Upstream", http.StatusBadGateway}
	ErrTimeout       = ErrorKind{"Timeout", http.StatusGatewayTimeout}
)
