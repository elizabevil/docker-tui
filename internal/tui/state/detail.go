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
	*s = DetailState{DetailTitle: title, ImageDetailContent: content}
}

func (s *DetailState) OpenImage(id, title string, data *runtimeapi.ImageDetail) {
	*s = DetailState{ImageDetailID: id, DetailTitle: title, ImageDetailData: data}
	if data != nil {
		s.DetailResourceType = ResourceImage
		s.DetailRawJSON, _ = sonic.Marshal(data) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	}
}

func (s *DetailState) Close() { *s = DetailState{} }

func (s *DetailState) Scroll(delta int) { s.DetailOffset = max(0, s.DetailOffset+delta) }

func (s *DetailState) CycleSource() {
	s.DetailOffset = 0
	switch s.DetailSourceType {
	case DetailSourceSection:
		s.DetailSourceType = DetailSourceYAML
	case DetailSourceYAML:
		s.DetailSourceType = DetailSourceJSON
	default:
		s.DetailSourceType = DetailSourceSection
	}
}

func (s *DetailState) SetRaw(resourceType ResourceType, raw []byte) {
	s.DetailRawJSON = raw
	s.DetailResourceType = resourceType
}

func (s *DetailState) SetContainerDetail(detail *runtimeapi.ContainerDetail) {
	s.ContainerDetail = detail
	s.DetailResourceType = ResourceContainer
	s.DetailRawJSON, _ = sonic.Marshal(detail) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
}

func (s *DetailState) SetVolumeDetail(detail *runtimeapi.VolumeDetail) {
	s.VolumeDetail = detail
	s.DetailResourceType = ResourceVolume
	s.DetailRawJSON, _ = sonic.Marshal(detail) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
}

func (s *DetailState) SetNetworkDetail(detail *runtimeapi.NetworkDetail) {
	s.NetworkDetail = detail
	s.DetailResourceType = ResourceNetwork
	s.DetailRawJSON, _ = sonic.Marshal(detail) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
}

func (s *DetailState) ApplyImage(id string, data *runtimeapi.ImageDetail) bool {
	if id != s.ImageDetailID {
		return false
	}
	s.ImageDetailData = data
	s.DetailResourceType = ResourceImage
	s.DetailRawJSON, _ = sonic.Marshal(data) //nolint:errcheck // marshalling typed structs; cannot fail in practice.
	s.DetailOffset = 0
	return true
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
