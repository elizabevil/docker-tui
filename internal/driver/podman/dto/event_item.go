package dto

import "time"

// EventItem is the Podman Libpod REST event item delivered over the
// /events stream. Mirrors the legacy EventItem in aliases_cgo.go.
type EventItem struct {
	Type       string            `json:"Type"`
	Status     string            `json:"Status"`
	Action     string            `json:"Action"`
	ID         string            `json:"ID"`
	Name       string            `json:"Name"`
	Time       time.Time         `json:"Time"`
	TimeNano   int64             `json:"TimeNano"`
	Attributes map[string]string `json:"Attributes"`
}
