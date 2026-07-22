package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// DecodeImageProgress decodes a Podman image stream JSON and emits progress
// events to the output channel. Returns nil on EOF.
func DecodeImageProgress(ctx context.Context, reader io.Reader, output chan<- runtimeapi.ImageTransferEvent) error {
	decoder := json.NewDecoder(reader)
	for {
		var message dto.ImageStreamMessage
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
		progress := runtimeapi.ImageTransferProgress{Status: strings.TrimSpace(message.Status)}
		if message.Progress != nil {
			progress.Current, progress.Total = message.Progress.Current, message.Progress.Total
		}
		if progress.Status != "" || progress.Current > 0 {
			EmitImageProgress(ctx, output, progress)
		}
	}
}

// EmitImageProgress sends a progress event to the output channel without
// blocking. Intermediate samples are dropped when the UI is busy.
func EmitImageProgress(ctx context.Context, output chan<- runtimeapi.ImageTransferEvent, progress runtimeapi.ImageTransferProgress) {
	if output == nil {
		return
	}
	select {
	case output <- runtimeapi.ImageTransferEvent{Progress: &progress}:
	case <-ctx.Done():
	default:
	}
}

// ProgressWriter wraps an io.Writer and emits progress events on each write.
type ProgressWriter struct {
	Ctx     context.Context
	Output  chan<- runtimeapi.ImageTransferEvent
	Status  string
	Current int64
	Writer  io.Writer
}

func (w *ProgressWriter) Write(data []byte) (int, error) {
	count, err := w.Writer.Write(data)
	w.Current += int64(count)
	EmitImageProgress(w.Ctx, w.Output, runtimeapi.ImageTransferProgress{Status: w.Status, Current: w.Current})
	return count, err
}

// ProgressReader wraps an io.Reader and emits progress events on each read.
type ProgressReader struct {
	Ctx     context.Context
	Output  chan<- runtimeapi.ImageTransferEvent
	Status  string
	Current int64
	Total   int64
	Reader  io.Reader
}

func (r *ProgressReader) Read(data []byte) (int, error) {
	count, err := r.Reader.Read(data)
	r.Current += int64(count)
	EmitImageProgress(r.Ctx, r.Output, runtimeapi.ImageTransferProgress{Status: r.Status, Current: r.Current, Total: r.Total})
	return count, err
}

// SplitTagTarget splits a container image reference into repository and tag.
func SplitTagTarget(reference string) (repository, tag string) {
	repository, tag = reference, "latest"
	lastSlash, lastColon := strings.LastIndex(reference, "/"), strings.LastIndex(reference, ":")
	if lastColon > lastSlash {
		repository, tag = reference[:lastColon], reference[lastColon+1:]
	}
	return repository, tag
}
