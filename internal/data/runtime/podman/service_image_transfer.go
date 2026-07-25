package podman

import (
	"context"
	"io"
	"os"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
)

// PodmanImageTransferService implements ImageTransferService for the Podman adapter.
type PodmanImageTransferService struct {
	Client *podman.Client
}

func (s PodmanImageTransferService) Run(ctx context.Context, request runtimeapi.ImageTransferRequest) (<-chan runtimeapi.ImageTransferEvent, error) {
	output := make(chan runtimeapi.ImageTransferEvent, 32)
	go func() {
		defer close(output)
		result, err := s.execute(ctx, request, output)
		if err != nil {
			err = runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, string(request.Operation)), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: request.Source}, runtimeapi.Podman)
		}
		output <- runtimeapi.ImageTransferEvent{Result: result, Error: err, Done: true}
	}()
	return output, nil
}

func (s PodmanImageTransferService) execute(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (*runtimeapi.ImageTransferResult, error) {
	result := &runtimeapi.ImageTransferResult{Operation: request.Operation, Source: request.Source, Destination: request.Destination, Path: request.Path}
	switch request.Operation {
	case runtimeapi.ImageTransferTag:
		return result, s.Client.REST.TagImage(ctx, request.Source, request.Destination)
	case runtimeapi.ImageTransferPush:
		if request.Source != request.Destination {
			if err := s.Client.REST.TagImage(ctx, request.Source, request.Destination); err != nil {
				return result, err
			}
		}
		reader, err := s.Client.REST.PushImageStream(ctx, request.Destination)
		if err != nil {
			return result, err
		}
		defer func() { _ = reader.Close() }() //nolint:errcheck // push stream drained by DecodeImageProgress.
		return result, runtimeapi.DecodeImageProgress(ctx, reader, output)
	case runtimeapi.ImageTransferSave:
		reader, err := s.Client.REST.SaveImageStream(ctx, request.Source)
		if err != nil {
			return result, err
		}
		defer func() { _ = reader.Close() }() //nolint:errcheck // save stream drained via io.Copy.
		file, err := os.OpenFile(request.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return result, err
		}
		completed := false
		defer func() {
			closeErr := file.Close()
			if err == nil {
				err = closeErr
			}
			if !completed {
				_ = os.Remove(request.Path) //nolint:errcheck // partial save cleanup; best-effort.
			}
		}()
		writer := &runtimeapi.ProgressWriter{Ctx: ctx, Output: output, Status: "saving", Writer: file}
		if _, err = io.Copy(writer, reader); err != nil {
			return result, err
		}
		completed = true
		return result, nil
	case runtimeapi.ImageTransferLoad:
		file, err := os.Open(request.Path)
		if err != nil {
			return result, err
		}
		defer func() { _ = file.Close() }() //nolint:errcheck // load source consumed by StreamLoad.
		var total int64
		if info, statErr := file.Stat(); statErr == nil {
			total = info.Size()
		}
		progress := &runtimeapi.ProgressReader{Ctx: ctx, Output: output, Status: "loading", Total: total, Reader: file}
		reader, err := s.Client.REST.LoadImageStream(ctx, progress)
		if err != nil {
			return result, err
		}
		defer func() { _ = reader.Close() }() //nolint:errcheck // load response body fully decoded.
		refs, err := podman.DecodeLoadReport(reader)
		result.References = refs
		return result, err
	default:
		return nil, runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceImage, string(request.Operation)))
	}
}
