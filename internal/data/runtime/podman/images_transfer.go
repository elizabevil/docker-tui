package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// TagImageREST tags an image via the Podman Libpod REST API.
func TagImageREST(ctx context.Context, client *Client, source, destination string) error {
	if client.REST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "image.tag", source, errPodmanRESTNotReady)
	}
	repository, tag := SplitTagTarget(destination)
	query := url.Values{"repo": {repository}, "tag": {tag}}
	return client.REST.Post(ctx, "image.tag", fmt.Sprintf(PathImageTag, url.PathEscape(source)), query, nil, nil)
}

// PushImageREST pushes an image via the Podman Libpod REST API.
func PushImageREST(ctx context.Context, client *Client, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) error {
	if request.Source != request.Destination {
		if err := TagImageREST(ctx, client, request.Source, request.Destination); err != nil {
			return err
		}
	}
	query := url.Values{"destination": {request.Destination}}
	reader, err := client.REST.StreamPost(ctx, "image.push", fmt.Sprintf(PathImagePush, url.PathEscape(request.Destination)), query, nil, "")
	if err != nil {
		return err
	}
	defer reader.Close()
	return DecodeImageProgress(ctx, reader, output)
}

// SaveImageREST saves an image to a tar archive via the Podman Libpod REST API.
func SaveImageREST(ctx context.Context, client *Client, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) (err error) {
	reader, err := client.REST.StreamGet(ctx, "image.save", PathImageSave, url.Values{"references": {request.Source}})
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
	writer := &ProgressWriter{Ctx: ctx, Output: output, Status: "saving", Writer: file}
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	completed = true
	return nil
}

// LoadImageREST loads an image from a tar archive via the Podman Libpod REST API.
func LoadImageREST(ctx context.Context, client *Client, request runtimeapi.ImageTransferRequest, output chan<- runtimeapi.ImageTransferEvent) ([]string, error) {
	file, err := os.Open(request.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var total int64
	if info, statErr := file.Stat(); statErr == nil {
		total = info.Size()
	}
	progress := &ProgressReader{Ctx: ctx, Output: output, Status: "loading", Total: total, Reader: file}
	reader, err := client.REST.StreamPost(ctx, "image.load", PathImageLoad, nil, progress, "application/x-tar")
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var report dto.ImageLoadReport
	if err := json.NewDecoder(reader).Decode(&report); err != nil && err != io.EOF {
		return nil, err
	}
	return report.Names, nil
}
