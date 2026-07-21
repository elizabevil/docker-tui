package runtime

import (
	"errors"
	"fmt"
)

type ErrorKind string

const (
	ErrorConnection  ErrorKind = "connection"
	ErrorNotFound    ErrorKind = "not_found"
	ErrorConflict    ErrorKind = "conflict"
	ErrorInvalid     ErrorKind = "invalid"
	ErrorUnsupported ErrorKind = "unsupported"
	ErrorPermission  ErrorKind = "permission"
	ErrorInternal    ErrorKind = "internal"
)

// Error gives callers stable error semantics across Docker and Podman.
type Error struct {
	Kind      ErrorKind
	Operation string
	Resource  string
	Err       error
}

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

func (e *Error) Unwrap() error { return e.Err }

func NewError(kind ErrorKind, operation, resource string, err error) error {
	return &Error{Kind: kind, Operation: operation, Resource: resource, Err: err}
}

func IsErrorKind(err error, kind ErrorKind) bool {
	var runtimeErr *Error
	return errors.As(err, &runtimeErr) && runtimeErr.Kind == kind
}

func UnsupportedError(operation string) error {
	return NewError(ErrorUnsupported, operation, "", fmt.Errorf("capability is not supported by this runtime"))
}
