package podman

import (
	"context"
	"errors"
	"fmt"
)

// ErrorKind classifies transport errors into stable semantic categories.
// This is a local subset of the runtime.ErrorKind values, sufficient for
// REST client error construction. Callers in the runtime package convert
// these into runtime.Error via MapRuntimeError.
type ErrorKind int

const (
	KindInvalid    ErrorKind = iota // 400-range client error
	KindNotFound                    // 404
	KindConnection                  // network / context failure
	KindInternal                    // 500-range server error
)

// Error implements the error interface for Podman transport errors.
type Error struct {
	Kind       ErrorKind
	Operation  string
	StatusCode int
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("podman: %s: %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("podman: %s", e.Operation)
}

func (e *Error) Unwrap() error { return e.Err }

// newPodmanError creates a classified podman transport error.
func newPodmanError(kind ErrorKind, op string, err error) *Error {
	return &Error{Kind: kind, Operation: op, Err: err}
}

// classifyHTTPStatus maps an HTTP status code to the closest error kind.
func classifyHTTPStatus(status int) ErrorKind {
	switch status {
	case 400, 422:
		return KindInvalid
	case 401:
		return KindInvalid
	case 403:
		return KindInvalid
	case 404:
		return KindNotFound
	case 409:
		return KindInvalid
	case 429:
		return KindInvalid
	case 502, 503, 504:
		return KindInternal
	default:
		if status >= 500 {
			return KindInternal
		}
		return KindInvalid
	}
}

// classifyContextError maps context errors before inspecting transport errors.
func classifyContextError(err error) ErrorKind {
	switch {
	case errors.Is(err, context.Canceled):
		return KindConnection
	case errors.Is(err, context.DeadlineExceeded):
		return KindConnection
	default:
		return KindInternal
	}
}

// isRetryableKind reports whether the error kind indicates a transient failure.
func isRetryableKind(kind ErrorKind) bool {
	return kind == KindConnection || kind == KindInternal
}

// IsPodmanError checks whether err wraps a Podman transport Error of the
// given kind.
func IsPodmanError(err error, kind ErrorKind) bool {
	var pe *Error
	return errors.As(err, &pe) && pe.Kind == kind
}

// newPodmanErrorf is a convenience wrapper that creates a podman.Error with
// a formatted message.
func newPodmanErrorf(kind ErrorKind, format string, args ...any) *Error {
	return newPodmanError(kind, "", fmt.Errorf(format, args...))
}
