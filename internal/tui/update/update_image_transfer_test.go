package update

import (
	"errors"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestHandleImageTransferProgressContinuesReading(t *testing.T) {
	model := &state.AppModel{}
	_, generation := model.ImageTransfer.Begin(runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferLoad}, audit.Trace{})
	events := make(chan runtimeapi.ImageTransferEvent)
	_, command := handleImageTransferReceived(model, state.ImageTransferReceived{
		Generation: generation,
		Event:      runtimeapi.ImageTransferEvent{Progress: &runtimeapi.ImageTransferProgress{Status: "loading", Current: 10, Total: 20}},
		Events:     events,
	})
	if command == nil || model.ImageTransfer.Progress.Current != 10 {
		t.Fatalf("progress state = %#v", model.ImageTransfer.Progress)
	}
}

func TestHandleImageTransferFailureIsVisible(t *testing.T) {
	model := &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeImageTransfer}}
	_, generation := model.ImageTransfer.Begin(runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferPush}, audit.Trace{})
	_, _ = handleImageTransferReceived(model, state.ImageTransferReceived{
		Generation: generation,
		Event:      runtimeapi.ImageTransferEvent{Error: errors.New("denied"), Done: true},
	})
	if model.Navigation.Mode != state.ModeNormal || model.Feedback.ErrorMessage == "" || model.ImageTransfer.Cancel != nil {
		t.Fatalf("terminal transfer state: mode=%v feedback=%#v transfer=%#v", model.Navigation.Mode, model.Feedback, model.ImageTransfer)
	}
}
