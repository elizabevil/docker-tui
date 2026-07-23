package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/errdefs"
)

type statusCoder interface{ Code() int }

func MapRuntimeError(err error, operation string, ref ResourceRef, driver Type) error {
	if err == nil {
		return nil
	}
	if existing, ok := errors.AsType[*Error](err); ok {
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
	kind := ClassifyContextError(err)
	if kind == ErrorInternal {
		switch {
		case errors.Is(err, os.ErrNotExist):
			kind = ErrorNotFound
		case errors.Is(err, os.ErrExist):
			kind = ErrorConflict
		case errors.Is(err, os.ErrPermission):
			kind = ErrorPermission
		case errdefs.IsNotFound(err):
			kind = ErrorNotFound
		case errdefs.IsConflict(err):
			kind = ErrorConflict
		case errdefs.IsInvalidParameter(err):
			kind = ErrorInvalid
		case errdefs.IsUnauthorized(err):
			kind = ErrorAuthentication
		case errdefs.IsForbidden(err):
			kind = ErrorPermission
		case errdefs.IsNotImplemented(err):
			kind = ErrorUnsupported
		case errdefs.IsUnavailable(err):
			kind = ErrorUnavailable
		case errdefs.IsDeadline(err):
			kind = ErrorTimeout
		case errdefs.IsCancelled(err):
			kind = ErrorCanceled
		default:
			var coded statusCoder
			if errors.As(err, &coded) {
				kind = ClassifyHTTPStatus(coded.Code())
			}
		}
	}
	return &Error{Kind: kind, Operation: operation, ResourceType: string(ref.Type), Resource: ref.ID, Driver: driver, Retryable: IsRetryableKind(kind), Err: err}
}

func DecodeImageProgress(ctx context.Context, reader io.Reader, output chan<- ImageTransferEvent) error {
	decoder := json.NewDecoder(reader)
	for {
		var message ImageStreamMessage
		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if message.Error.Message != "" {
			return fmt.Errorf("%s", message.Error.Message)
		}
		if message.ErrorMessage != "" {
			return fmt.Errorf("%s", message.ErrorMessage)
		}
		progress := ImageTransferProgress{Status: strings.TrimSpace(message.Status)}
		if message.Progress != nil {
			progress.Current, progress.Total = message.Progress.Current, message.Progress.Total
		}
		if progress.Status != "" || progress.Current > 0 {
			EmitImageProgress(ctx, output, progress)
		}
	}
}

func EmitImageProgress(ctx context.Context, output chan<- ImageTransferEvent, progress ImageTransferProgress) {
	if output == nil {
		return
	}
	select {
	case output <- ImageTransferEvent{Progress: &progress}:
	case <-ctx.Done():
	default:
	}
}

type ProgressWriter struct {
	Ctx     context.Context
	Output  chan<- ImageTransferEvent
	Status  string
	Current int64
	Writer  io.Writer
}

func (w *ProgressWriter) Write(data []byte) (int, error) {
	count, err := w.Writer.Write(data)
	w.Current += int64(count)
	EmitImageProgress(w.Ctx, w.Output, ImageTransferProgress{Status: w.Status, Current: w.Current})
	return count, err
}

type ProgressReader struct {
	Ctx     context.Context
	Output  chan<- ImageTransferEvent
	Status  string
	Current int64
	Total   int64
	Reader  io.Reader
}

func (r *ProgressReader) Read(data []byte) (int, error) {
	count, err := r.Reader.Read(data)
	r.Current += int64(count)
	EmitImageProgress(r.Ctx, r.Output, ImageTransferProgress{Status: r.Status, Current: r.Current, Total: r.Total})
	return count, err
}

func SplitTagTarget(reference string) (repository, tag string) {
	repository, tag = reference, "latest"
	lastSlash, lastColon := strings.LastIndex(reference, "/"), strings.LastIndex(reference, ":")
	if lastColon > lastSlash {
		repository, tag = reference[:lastColon], reference[lastColon+1:]
	}
	return repository, tag
}
