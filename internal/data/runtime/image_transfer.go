package runtime

import "context"

// ImageTransferOperation identifies an image distribution workflow.
type ImageTransferOperation string

const (
	ImageTransferTag  ImageTransferOperation = "tag"
	ImageTransferPush ImageTransferOperation = "push"
	ImageTransferSave ImageTransferOperation = "save"
	ImageTransferLoad ImageTransferOperation = "load"
)

// ImageTransferRequest contains runtime-neutral image workflow inputs.
type ImageTransferRequest struct {
	Operation   ImageTransferOperation
	Source      string
	Destination string
	Path        string
}

// ImageTransferProgress is a point-in-time transfer update. Total is zero
// when the runtime cannot determine the complete transfer size.
type ImageTransferProgress struct {
	Status  string
	Current int64
	Total   int64
}

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

type ImageTransferService interface {
	Run(context.Context, ImageTransferRequest) (<-chan ImageTransferEvent, error)
}
