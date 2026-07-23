package dto

// LayerHistoryEntry mirrors a single entry of GET /libpod/images/{id}/history.
// Field shape and JSON tags match the Libpod response.
type LayerHistoryEntry struct {
	ID        string `json:"Id"`
	Created   int64  `json:"Created"`
	CreatedBy string `json:"CreatedBy"`
	Size      int64  `json:"Size"`
	Comment   string `json:"Comment,omitempty"`
}
