package state

import (
	"github.com/bytedance/sonic"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type DetailSource string

const (
	DetailSourceSection DetailSource = ""
	DetailSourceYAML    DetailSource = "yaml"
	DetailSourceJSON    DetailSource = "json"
)

type DetailLineKind uint8

const (
	DetailLineSection DetailLineKind = iota
	DetailLineSubtitle
	DetailLineValue
	DetailLinePlain
)

type DetailDocumentLine struct {
	Kind  DetailLineKind
	Left  string
	Right string
}

type DetailDocument struct {
	Revision uint64
	Source   DetailSource
	Lines    []DetailDocumentLine
}

type DetailState struct {
	ImageDetailID      string
	ImageDetailContent string
	ImageDetailData    *runtimeapi.ImageDetail
	DetailTitle        string
	DetailHint         string
	DetailOffset       int
	DetailRawJSON      []byte
	DetailSourceType   DetailSource
	DetailResourceType ResourceType
	VolumeDetail       *runtimeapi.VolumeDetail
	NetworkDetail      *runtimeapi.NetworkDetail
	ContainerDetail    *runtimeapi.ContainerDetail
	Revision           uint64
	Documents          map[DetailSource]DetailDocument
}

type (
	ContainerDetailLoaded struct {
		ContainerID string
		Title       string
		Detail      *runtimeapi.ContainerDetail
		Error       error
	}

	VolumeDetailLoaded struct {
		VolumeID string
		Title    string
		Detail   *runtimeapi.VolumeDetail
		Error    error
	}

	NetworkDetailLoaded struct {
		NetworkID string
		Title     string
		Detail    *runtimeapi.NetworkDetail
		Error     error
	}
)

func (s *DetailState) Open(title, content string) {
	revision := s.Revision + 1
	*s = DetailState{DetailTitle: title, ImageDetailContent: content, Revision: revision}
}

func (s *DetailState) OpenImage(id, title string, data *runtimeapi.ImageDetail) {
	revision := s.Revision + 1
	*s = DetailState{ImageDetailID: id, DetailTitle: title, ImageDetailData: data, Revision: revision}
	if data != nil {
		s.DetailResourceType = ResourceImage
		s.DetailRawJSON, _ = sonic.Marshal(data) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	}
}

func (s *DetailState) Close() {
	revision := s.Revision + 1
	*s = DetailState{Revision: revision}
}

func (s *DetailState) Scroll(delta int) { s.DetailOffset = max(0, s.DetailOffset+delta) }

func (s *DetailState) HasRawSource() bool { return len(s.DetailRawJSON) > 0 }

func (s *DetailState) CycleSource() bool {
	if !s.HasRawSource() {
		return false
	}
	s.DetailOffset = 0
	switch s.DetailSourceType {
	case DetailSourceSection:
		s.DetailSourceType = DetailSourceYAML
	case DetailSourceYAML:
		s.DetailSourceType = DetailSourceJSON
	default:
		s.DetailSourceType = DetailSourceSection
	}
	return true
}

func (s *DetailState) SetRaw(resourceType ResourceType, raw []byte) {
	s.DetailRawJSON = raw
	s.DetailResourceType = resourceType
	s.invalidateDocument()
}

func (s *DetailState) SetContainerDetail(detail *runtimeapi.ContainerDetail) {
	s.ContainerDetail = detail
	s.DetailResourceType = ResourceContainer
	if detail != nil {
		s.DetailRawJSON, _ = sonic.Marshal(detail) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	}
	s.invalidateDocument()
}

func (s *DetailState) SetVolumeDetail(detail *runtimeapi.VolumeDetail) {
	s.VolumeDetail = detail
	s.DetailResourceType = ResourceVolume
	if detail != nil {
		s.DetailRawJSON, _ = sonic.Marshal(detail) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	}
	s.invalidateDocument()
}

func (s *DetailState) SetNetworkDetail(detail *runtimeapi.NetworkDetail) {
	s.NetworkDetail = detail
	s.DetailResourceType = ResourceNetwork
	if detail != nil {
		s.DetailRawJSON, _ = sonic.Marshal(detail) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	}
	s.invalidateDocument()
}

func (s *DetailState) ApplyImage(id string, data *runtimeapi.ImageDetail) bool {
	if id != s.ImageDetailID {
		return false
	}
	s.ImageDetailData = data
	s.DetailResourceType = ResourceImage
	if data != nil {
		s.DetailRawJSON, _ = sonic.Marshal(data) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	} else {
		s.DetailRawJSON = nil
	}
	s.DetailOffset = 0
	s.invalidateDocument()
	return true
}

func (s *DetailState) invalidateDocument() {
	s.Revision++
	s.Documents = nil
}

func (s *DetailState) VisibleOffset(total, visible int) int {
	return min(max(0, s.DetailOffset), max(0, total-visible))
}

// ClampVisibleOffset keeps scroll state inside the currently rendered document.
// Without the write-back, repeated scrolling at the bottom accumulates an
// invisible overscroll that must be unwound before upward movement is visible.
func (s *DetailState) ClampVisibleOffset(total, visible int) int {
	s.DetailOffset = s.VisibleOffset(total, visible)
	return s.DetailOffset
}
