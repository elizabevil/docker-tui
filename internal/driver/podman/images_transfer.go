package podman

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

// TagImage tags an image via the Podman Libpod REST API.
func (c *RESTClient) TagImage(ctx context.Context, source, destination string) error {
	repo, tag := splitTagTarget(destination)
	query := url.Values{"repo": {repo}, "tag": {tag}}
	return c.Post(ctx, ImageTagPath(source), query, nil, nil)
}

// PushImageStream opens a push stream to the Podman Libpod endpoint and
// returns the response body. The caller owns the returned stream and is
// responsible for decoding progress events.
func (c *RESTClient) PushImageStream(ctx context.Context, destination string) (io.ReadCloser, error) {
	query := url.Values{"destination": {destination}}
	return c.StreamPost(ctx, ImagePushPath(destination), query, nil, "")
}

// SaveImageStream opens a save stream to the Podman Libpod endpoint and
// returns the response body. The caller owns the returned stream and is
// responsible for writing to a file.
func (c *RESTClient) SaveImageStream(ctx context.Context, references string) (io.ReadCloser, error) {
	return c.StreamGet(ctx, PathImageSave, url.Values{"references": {references}})
}

// LoadImageStream sends a tar archive body to the Podman Libpod load
// endpoint and returns the response body, which contains the loaded image
// names. The caller owns the returned stream.
func (c *RESTClient) LoadImageStream(ctx context.Context, body io.Reader) (io.ReadCloser, error) {
	return c.StreamPost(ctx, PathImageLoad, nil, body, "application/x-tar")
}

// ImageLoadReport is the JSON response from the Podman load endpoint.
type ImageLoadReport struct {
	Names []string `json:"names"`
}

// DecodeLoadReport decodes the image load response from the reader.
func DecodeLoadReport(reader io.Reader) ([]string, error) {
	var report ImageLoadReport
	if err := json.NewDecoder(reader).Decode(&report); err != nil {
		return nil, err
	}
	return report.Names, nil
}

// splitTagTarget splits a container image reference into repository and tag.
func splitTagTarget(reference string) (repository, tag string) {
	repository, tag = reference, "latest"
	lastSlash, lastColon := strings.LastIndex(reference, "/"), strings.LastIndex(reference, ":")
	if lastColon > lastSlash {
		repository, tag = reference[:lastColon], reference[lastColon+1:]
	}
	return repository, tag
}
