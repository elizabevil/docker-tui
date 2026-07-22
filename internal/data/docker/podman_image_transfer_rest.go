package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func (s imageTransferService) executePodman(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (*runtimeapi.ImageTransferResult, error) {
	if s.client.podmanREST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, "image."+string(request.Operation), request.Source, fmt.Errorf("Podman REST transport is not initialized"))
	}
	result := &runtimeapi.ImageTransferResult{Operation: request.Operation, Source: request.Source, Destination: request.Destination, Path: request.Path}
	switch request.Operation {
	case runtimeapi.ImageTransferTag:
		return result, s.client.tagImagePodmanREST(ctx, request.Source, request.Destination)
	case runtimeapi.ImageTransferPush:
		return result, s.pushPodman(ctx, request, output)
	case runtimeapi.ImageTransferSave:
		return result, s.savePodman(ctx, request, output)
	case runtimeapi.ImageTransferLoad:
		refs, err := s.loadPodman(ctx, request, output)
		result.References = refs
		return result, err
	default:
		return nil, runtimeapi.UnsupportedError("image." + string(request.Operation))
	}
}

func (c *Client) tagImagePodmanREST(ctx context.Context, source, destination string) error {
	if c.podmanREST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "image.tag", source, fmt.Errorf("Podman REST transport is not initialized"))
	}
	repository, tag := splitTagTarget(destination)
	query := url.Values{"repo": {repository}, "tag": {tag}}
	return c.podmanREST.Post(ctx, "image.tag", "/images/"+url.PathEscape(source)+"/tag", query, nil, nil)
}

func (s imageTransferService) pushPodman(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) error {
	if request.Source != request.Destination {
		if err := s.client.tagImagePodmanREST(ctx, request.Source, request.Destination); err != nil {
			return err
		}
	}
	query := url.Values{"destination": {request.Destination}}
	reader, err := s.client.podmanREST.StreamPost(ctx, "image.push", "/images/"+url.PathEscape(request.Destination)+"/push", query, nil, "")
	if err != nil {
		return err
	}
	defer reader.Close()
	return decodeImageProgress(ctx, reader, output)
}

func (s imageTransferService) savePodman(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (err error) {
	reader, err := s.client.podmanREST.StreamGet(ctx, "image.save", "/images/export", url.Values{"references": {request.Source}})
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

func (s imageTransferService) loadPodman(ctx context.Context, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) ([]string, error) {
	file, err := os.Open(request.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var total int64
	if info, statErr := file.Stat(); statErr == nil {
		total = info.Size()
	}
	progress := &progressReader{ctx: ctx, output: output, status: "loading", total: total, reader: file}
	reader, err := s.client.podmanREST.StreamPost(ctx, "image.load", "/images/load", nil, progress, "application/x-tar")
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var report struct {
		Names []string `json:"Names"`
	}
	if err := json.NewDecoder(reader).Decode(&report); err != nil && err != io.EOF {
		return nil, err
	}
	return report.Names, nil
}

func splitTagTarget(reference string) (repository, tag string) {
	repository, tag = reference, "latest"
	lastSlash, lastColon := strings.LastIndex(reference, "/"), strings.LastIndex(reference, ":")
	if lastColon > lastSlash {
		repository, tag = reference[:lastColon], reference[lastColon+1:]
	}
	return repository, tag
}

func decodeImageProgress(ctx context.Context, reader io.Reader, output chan<- runtimeapi.ImageTransferEvent) error {
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
