package dto

// ImageStreamMessage represents a single message from the Podman image
// pull/push/load/save JSON stream. All nested shapes are named.
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

// ImagePruneReportItem is the Podman Libpod REST response item for image
// prune operations.
type ImagePruneReportItem struct {
	ID   string `json:"Id"`
	Size uint64 `json:"Size"`
	Err  string `json:"Err"`
}

// VolumePruneReportItem is the Podman Libpod REST response item for volume
// prune operations. Mirrors VolumePruneReport but with int64 Size; kept
// for callers that already use int64 (see runtime/podman/stream_types).
type VolumePruneReportItem struct {
	ID   string `json:"ID"`
	Size int64  `json:"Size"`
	Err  string `json:"Err"`
}

// NetworkPruneReportItem is the Podman Libpod REST response item for network
// prune operations. Mirrors NetworkPruneReport.
type NetworkPruneReportItem struct {
	Name  string `json:"Name"`
	Error string `json:"Error"`
}
