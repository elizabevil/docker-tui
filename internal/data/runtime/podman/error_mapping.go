package podman

import (
	"errors"
	"fmt"
	"os"

	"github.com/docker/docker/errdefs"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type statusCoder interface{ Code() int }

func mapRuntimeError(err error, operation string, ref runtimeapi.ResourceRef, driver runtimeapi.Type) error {
	if err == nil {
		return nil
	}
	var existing *runtimeapi.Error
	if errors.As(err, &existing) {
		mapped := *existing
		if mapped.ResourceType == "" {
			mapped.ResourceType = string(ref.Type)
		}
		if mapped.Resource == "" {
			mapped.Resource = ref.ID
		}
		if mapped.Driver == "" {
			mapped.Driver = driver
		}
		return &mapped
	}
	kind := runtimeapi.ClassifyContextError(err)
	if kind == runtimeapi.ErrorInternal {
		switch {
		case errors.Is(err, os.ErrNotExist):
			kind = runtimeapi.ErrorNotFound
		case errors.Is(err, os.ErrExist):
			kind = runtimeapi.ErrorConflict
		case errors.Is(err, os.ErrPermission):
			kind = runtimeapi.ErrorPermission
		case errdefs.IsNotFound(err):
			kind = runtimeapi.ErrorNotFound
		case errdefs.IsConflict(err):
			kind = runtimeapi.ErrorConflict
		case errdefs.IsInvalidParameter(err):
			kind = runtimeapi.ErrorInvalid
		case errdefs.IsUnauthorized(err):
			kind = runtimeapi.ErrorAuthentication
		case errdefs.IsForbidden(err):
			kind = runtimeapi.ErrorPermission
		case errdefs.IsNotImplemented(err):
			kind = runtimeapi.ErrorUnsupported
		case errdefs.IsUnavailable(err):
			kind = runtimeapi.ErrorUnavailable
		case errdefs.IsDeadline(err):
			kind = runtimeapi.ErrorTimeout
		case errdefs.IsCancelled(err):
			kind = runtimeapi.ErrorCanceled
		default:
			var coded statusCoder
			if errors.As(err, &coded) {
				kind = runtimeapi.ClassifyHTTPStatus(coded.Code())
			}
		}
	}
	return &runtimeapi.Error{Kind: kind, Operation: operation, ResourceType: string(ref.Type), Resource: ref.ID, Driver: driver, Retryable: runtimeapi.IsRetryableKind(kind), Err: err}
}

func unsupportedError(operation string) error {
	return fmt.Errorf("unsupported operation %q", operation)
}
