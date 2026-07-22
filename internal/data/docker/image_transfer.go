package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/image"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type imageTransferService struct{ client *Client }

func (c *Client) ImageTransfers() runtimeapi.ImageTransferService {
	return imageTransferService{client: c}
}

func (s imageTransferService) Run(ctx context.Context, request runtimeapi.ImageTransferRequest) (<-chan runtimeapi.ImageTransferEvent, error) {
	if err := validateImageTransferRequest(request); err != nil {
		return nil, err
	}
	output := make(chan runtimeapi.ImageTransferEvent, 32)
	go func() {
		defer close(output)
		result, err := s.execute(ctx, request, output)
		if err != nil {
			err = mapRuntimeError(err, "image."+string(request.Operation), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: request.Source}, s.client.RuntimeType)
		}
		output <- runtimeapi.ImageTransferEvent{Result: result, Error: err, Done: true}
	}()
	return output, nil
}

func validateImageTransferRequest(request runtimeapi.ImageTransferRequest) error {
	required := func(value, field string) error {
		if strings.TrimSpace(value) == "" {
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, "image."+string(request.Operation), request.Source, fmt.Errorf("%s is required", field))
		}
		return nil
	}
	switch request.Operation {
	case runtimeapi.ImageTransferTag, runtimeapi.ImageTransferPush:
		if err := required(request.Source, "source"); err != nil {
			return err
		}
		return required(request.Destination, "destination")
	case runtimeapi.ImageTransferSave:
		if err := required(request.Source, "source"); err != nil {
			return err
		}
		return required(request.Path, "path")
	case runtimeapi.ImageTransferLoad:
		return required(request.Path, "path")
	default:
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, "image.transfer", request.Source, fmt.Errorf("unknown operation %q", request.Operation))
	}
}

func (s imageTransferService) execute(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (*runtimeapi.ImageTransferResult, error) {
	result := &runtimeapi.ImageTransferResult{Operation: request.Operation, Source: request.Source, Destination: request.Destination, Path: request.Path}
	switch request.Operation {
	case runtimeapi.ImageTransferTag:
		return result, s.client.cli.ImageTag(ctx, request.Source, request.Destination)
	case runtimeapi.ImageTransferPush:
		return result, s.push(ctx, request, output)
	case runtimeapi.ImageTransferSave:
		return result, s.save(ctx, request, output)
	case runtimeapi.ImageTransferLoad:
		refs, err := s.load(ctx, request, output)
		result.References = refs
		return result, err
	default:
		return nil, runtimeapi.UnsupportedError("image." + string(request.Operation))
	}
}

func (s imageTransferService) push(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) error {
	if request.Source != request.Destination {
		if err := s.client.cli.ImageTag(ctx, request.Source, request.Destination); err != nil {
			return err
		}
	}
	reader, err := s.client.cli.ImagePush(ctx, request.Destination, image.PushOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		var message imageStreamMessage
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
			emitImageProgress(ctx, output, progress)
		}
	}
}

func (s imageTransferService) save(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (err error) {
	reader, err := s.client.cli.ImageSave(ctx, []string{request.Source})
	if err != nil {
		return err
	}
	defer reader.Close()
	file, err := os.OpenFile(request.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	completed := false
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
		if !completed {
			_ = os.Remove(request.Path)
		}
	}()
	writer := &progressWriter{ctx: ctx, output: output, status: "saving", writer: file}
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	completed = true
	return nil
}

func (s imageTransferService) load(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) ([]string, error) {
	file, err := os.Open(request.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var total int64
	if info, statErr := file.Stat(); statErr == nil {
		total = info.Size()
	}
	reader := &progressReader{ctx: ctx, output: output, status: "loading", total: total, reader: bufio.NewReader(file)}
	response, err := s.client.cli.ImageLoad(ctx, reader)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var references []string
	decoder := json.NewDecoder(response.Body)
	for {
		var message imageStreamMessage
		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if message.Error.Message != "" {
			return nil, fmt.Errorf("%s", message.Error.Message)
		}
		if message.ErrorMessage != "" {
			return nil, fmt.Errorf("%s", message.ErrorMessage)
		}
		if stream := strings.TrimSpace(message.Stream); stream != "" {
			references = append(references, stream)
		}
	}
	return references, nil
}

type imageStreamMessage struct {
	Stream       string `json:"stream"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error"`
	Error        struct {
		Message string `json:"message"`
	} `json:"errorDetail"`
	Progress *struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
}

type progressWriter struct {
	ctx     context.Context
	output  chan<- runtimeapi.ImageTransferEvent
	status  string
	current int64
	writer  io.Writer
}

func (w *progressWriter) Write(data []byte) (int, error) {
	count, err := w.writer.Write(data)
	w.current += int64(count)
	emitImageProgress(w.ctx, w.output, runtimeapi.ImageTransferProgress{Status: w.status, Current: w.current})
	return count, err
}

type progressReader struct {
	ctx     context.Context
	output  chan<- runtimeapi.ImageTransferEvent
	status  string
	current int64
	total   int64
	reader  io.Reader
}

func (r *progressReader) Read(data []byte) (int, error) {
	count, err := r.reader.Read(data)
	r.current += int64(count)
	emitImageProgress(r.ctx, r.output, runtimeapi.ImageTransferProgress{Status: r.status, Current: r.current, Total: r.total})
	return count, err
}

func emitImageProgress(ctx context.Context, output chan<- runtimeapi.ImageTransferEvent, progress runtimeapi.ImageTransferProgress) {
	select {
	case output <- runtimeapi.ImageTransferEvent{Progress: &progress}:
	case <-ctx.Done():
	default:
		// Progress is advisory; drop intermediate samples instead of stalling
		// the runtime stream when the UI is busy.
	}
}
