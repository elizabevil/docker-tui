package state

import (
	"context"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type ImageTransferState struct {
	Generation uint64
	Context    context.Context
	Cancel     context.CancelFunc
	Request    runtimeapi.ImageTransferRequest
	Progress   runtimeapi.ImageTransferProgress
	Audit      audit.Trace
}

func (s *ImageTransferState) Begin(request runtimeapi.ImageTransferRequest, trace audit.Trace) (context.Context, uint64) {
	s.Stop()
	s.Generation++
	s.Context, s.Cancel = context.WithCancel(context.Background())
	s.Request = request
	s.Progress = runtimeapi.ImageTransferProgress{Status: "starting"}
	s.Audit = trace
	return s.Context, s.Generation
}

func (s *ImageTransferState) Stop() {
	if s.Cancel != nil {
		s.Cancel()
	}
	s.Context = nil
	s.Cancel = nil
}

func (s *ImageTransferState) Reset() {
	s.Stop()
	s.Request = runtimeapi.ImageTransferRequest{}
	s.Progress = runtimeapi.ImageTransferProgress{}
	s.Audit = audit.Trace{}
}

func (s *ImageTransferState) Current(generation uint64) bool {
	return generation != 0 && generation == s.Generation && s.Cancel != nil
}

type ImageTransferReceived struct {
	Generation uint64
	Event      runtimeapi.ImageTransferEvent
	Events     <-chan runtimeapi.ImageTransferEvent
}
