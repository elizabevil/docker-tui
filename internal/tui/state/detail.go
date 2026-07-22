package state

import (
	"github.com/bytedance/sonic"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
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
	ImageDetailData    *dockerclient.ImageDetailData
	DetailTitle        string
	DetailHint         string
	DetailOffset       int
	DetailRawJSON      []byte
	DetailSourceType   DetailSource
	DetailResourceType ResourceType
	VolumeDetail       *runtimeapi.VolumeDetail
	NetworkDetail      *runtimeapi.NetworkDetail
}

func (s *DetailState) Open(title, content string) {
	*s = DetailState{DetailTitle: title, ImageDetailContent: content}
}

func (s *DetailState) OpenImage(id, title string, data *dockerclient.ImageDetailData) {
	*s = DetailState{ImageDetailID: id, DetailTitle: title, ImageDetailData: data}
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

func (s *DetailState) SetVolumeDetail(detail *runtimeapi.VolumeDetail) {
	s.VolumeDetail = detail
	s.DetailResourceType = ResourceVolume
	s.DetailRawJSON, _ = sonic.Marshal(detail)
}

func (s *DetailState) SetNetworkDetail(detail *runtimeapi.NetworkDetail) {
	s.NetworkDetail = detail
	s.DetailResourceType = ResourceNetwork
	s.DetailRawJSON, _ = sonic.Marshal(detail)
}

func (s *DetailState) ApplyImage(id string, data *dockerclient.ImageDetailData) bool {
	if id != s.ImageDetailID {
		return false
	}
	s.ImageDetailData = data
	s.DetailOffset = 0
	return true
}

func (s *DetailState) VisibleOffset(total, visible int) int {
	return min(max(0, s.DetailOffset), max(0, total-visible))
}
