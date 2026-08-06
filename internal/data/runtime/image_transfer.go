package runtime

import (
	"context"
	"fmt"
	"strings"
)

// ImageTransferOperation identifies an image distribution workflow.
type ImageTransferOperation string

// Transfer operation constants. Each maps to a distinct Docker/Podman API
// call; the adapter selects the appropriate SDK method.
const (
	// ImageTransferTag tags a local image with a new reference.
	ImageTransferTag ImageTransferOperation = "tag"
	// ImageTransferPush pushes an image to a remote registry.
	ImageTransferPush ImageTransferOperation = "push"
	// ImageTransferSave saves an image to a tar archive.
	ImageTransferSave ImageTransferOperation = "save"
	// ImageTransferLoad loads an image from a tar archive.
	ImageTransferLoad ImageTransferOperation = "load"
)

// ImageTransferRequest contains runtime-neutral image workflow inputs.
type ImageTransferRequest struct {
	Operation   ImageTransferOperation
	Source      string
	Destination string
	Path        string
}

// Validate reports whether the request has the fields required by its
// operation. The runtime adapter calls Validate before opening a channel
// so that callers never see a half-started workflow for malformed input.
func (r ImageTransferRequest) Validate() error {
	required := func(value, field string) error {
		if strings.TrimSpace(value) == "" {
			return NewError(ErrorInvalid, "image."+string(r.Operation), r.Source,
				fmt.Errorf("%s is required", field))
		}
		return nil
	}
	switch r.Operation {
	case ImageTransferTag, ImageTransferPush:
		if err := required(r.Source, "source"); err != nil {
			return err
		}
		return required(r.Destination, "destination")
	case ImageTransferSave:
		if err := required(r.Source, "source"); err != nil {
			return err
		}
		return required(r.Path, "path")
	case ImageTransferLoad:
		return required(r.Path, "path")
	default:
		return NewError(ErrorInvalid, "image.transfer", r.Source,
			fmt.Errorf("unknown operation %q", r.Operation))
	}
}

// ImageTransferProgress is a point-in-time transfer update. Total is zero
// when the runtime cannot determine the complete transfer size.
type ImageTransferProgress struct {
	Status  string
	Current int64
	Total   int64
}

// ImageTransferResult contains the outcome of a completed image transfer workflow.
type ImageTransferResult struct {
	Operation   ImageTransferOperation
	Source      string
	Destination string
	Path        string
	References  []string
}

// ImageTransferEvent contains either progress or the terminal result. Exactly
// one terminal event is emitted for each successfully started workflow.
type ImageTransferEvent struct {
	Progress *ImageTransferProgress
	Result   *ImageTransferResult
	Error    error
	Done     bool
}

// ImageTransferService drives image distribution workflows (tag, push, save,
// load). The caller receives a channel of progress and terminal events.
type ImageTransferService interface {
	Run(context.Context, ImageTransferRequest) (<-chan ImageTransferEvent, error)
}

// ImageStreamMessage represents a single message from the Docker/Podman image
// pull/push/load/save JSON stream. Both runtimes emit the same wire format.
type ImageStreamMessage struct {
	Stream       string               `json:"stream"`
	Status       string               `json:"status"`
	ErrorMessage string               `json:"error"`
	Error        ImageStreamErr       `json:"errorDetail"`
	Progress     *ImageStreamProgress `json:"progressDetail"`
}

// ImageStreamErr is the nested error payload of ImageStreamMessage.
type ImageStreamErr struct {
	Message string `json:"message"`
}

// ImageStreamProgress is the nested progress payload of ImageStreamMessage.
type ImageStreamProgress struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}
