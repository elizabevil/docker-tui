package runtime

import (
	"context"
	"errors"
	"fmt"
)

// ErrorKind classifies runtime errors into stable semantic categories.
// Callers use these kinds to decide retry policy, UI messaging, and error
// projection without inspecting transport-specific error strings.
type ErrorKind string

// Standard error kinds returned by all runtime drivers.
const (
	ErrorConnection     ErrorKind = "connection"
	ErrorTimeout        ErrorKind = "timeout"
	ErrorCanceled       ErrorKind = "canceled"
	ErrorAuthentication ErrorKind = "authentication"
	ErrorNotFound       ErrorKind = "not_found"
	ErrorAlreadyExists  ErrorKind = "already_exists"
	ErrorConflict       ErrorKind = "conflict"
	ErrorInvalid        ErrorKind = "invalid"
	ErrorUnsupported    ErrorKind = "unsupported"
	ErrorPermission     ErrorKind = "permission"
	ErrorRateLimited    ErrorKind = "rate_limited"
	ErrorUnavailable    ErrorKind = "unavailable"
	ErrorInternal       ErrorKind = "internal"
)

// Error gives callers stable error semantics across Docker and Podman.
type Error struct {
	Kind         ErrorKind
	Operation    string
	ResourceType string
	Resource     string
	Driver       Type
	Retryable    bool
	StatusCode   int
	Err          error
}

// Error implements the error interface and carries structured metadata for
// runtime-agnostic error handling.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	message := string(e.Kind)
	if e.Operation != "" {
		message = e.Operation + ": " + message
	}
	if e.Resource != "" {
		message += " " + e.Resource
	}
	if e.Err != nil {
		message += ": " + e.Err.Error()
	}
	return message
}

// Unwrap returns the underlying cause for errors.Is/As traversal.
func (e *Error) Unwrap() error { return e.Err }

// NewError constructs a classified runtime error with retry semantics derived
// from the error kind.
func NewError(kind ErrorKind, operation, resource string, err error) error {
	return &Error{Kind: kind, Operation: operation, Resource: resource, Retryable: IsRetryableKind(kind), Err: err}
}

// IsRetryableKind reports whether the error kind indicates a transient failure
// that may succeed on retry (connection, timeout, rate limit, unavailable).
func IsRetryableKind(kind ErrorKind) bool {
	switch kind {
	case ErrorConnection, ErrorTimeout, ErrorRateLimited, ErrorUnavailable:
		return true
	default:
		return false
	}
}

// ClassifyContextError maps cancellation before inspecting transport-specific errors.
func ClassifyContextError(err error) ErrorKind {
	switch {
	case errors.Is(err, context.Canceled):
		return ErrorCanceled
	case errors.Is(err, context.DeadlineExceeded):
		return ErrorTimeout
	default:
		return ErrorInternal
	}
}

// ClassifyHTTPStatus maps an HTTP status code to the closest error kind.
func ClassifyHTTPStatus(status int) ErrorKind {
	switch status {
	case 400, 422:
		return ErrorInvalid
	case 401:
		return ErrorAuthentication
	case 403:
		return ErrorPermission
	case 404:
		return ErrorNotFound
	case 409:
		return ErrorConflict
	case 429:
		return ErrorRateLimited
	case 502, 503, 504:
		return ErrorUnavailable
	default:
		if status >= 500 {
			return ErrorInternal
		}
		return ErrorInvalid
	}
}

// IsErrorKind checks whether err wraps a runtime Error of the given kind.
func IsErrorKind(err error, kind ErrorKind) bool {
	var runtimeErr *Error
	return errors.As(err, &runtimeErr) && runtimeErr.Kind == kind
}

// UnsupportedError returns a pre-classified unsupported-operation error.
func UnsupportedError(operation string) error {
	return NewError(ErrorUnsupported, operation, "", fmt.Errorf("capability is not supported by this runtime"))
}
