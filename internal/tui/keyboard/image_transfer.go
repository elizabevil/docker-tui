package keyboard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func openImageWorkflow(m *state.AppModel, operation runtimeapi.ImageTransferOperation) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil || m.Navigation.ActivePanel != state.PanelImages {
		return m, nil
	}
	selected := m.Resources.Images.Selected()
	if operation != runtimeapi.ImageTransferLoad && selected == nil {
		return m, nil
	}

	var kind state.DialogKind
	var title, input string
	switch operation {
	case runtimeapi.ImageTransferTag:
		kind, title = state.DialogImageTag, i18n.T("image.tag.title")
	case runtimeapi.ImageTransferPush:
		kind, title = state.DialogImagePush, i18n.T("image.push.title")
		input = fullImageRef(selected)
	case runtimeapi.ImageTransferSave:
		kind, title = state.DialogImageSave, i18n.T("image.save.title")
		input = filepath.Clean(tagName(selected) + ".tar")
	case runtimeapi.ImageTransferLoad:
		kind, title = state.DialogImageLoad, i18n.T("image.load.title")
	default:
		return m, nil
	}
	m.Dialog.Open(state.DialogSpec{Kind: kind, Title: title, Action: string(operation), Input: input})
	m.Navigation.Mode = m.Dialog.Kind.Mode()
	return m, nil
}

func handleImageWorkflowInput(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEsc:
		clearDialogState(m)
		return m, nil
	case keys.KeyEnter:
		value := strings.TrimSpace(m.Dialog.Input.Text)
		if value == "" {
			ShowToastWarn(m, i18n.T("image.transfer.required"))
			return m, nil
		}
		operation := runtimeapi.ImageTransferOperation(m.Dialog.Action)
		request := runtimeapi.ImageTransferRequest{Operation: operation}
		if operation == runtimeapi.ImageTransferLoad {
			request.Path = filepath.Clean(value)
		} else {
			selected := m.Resources.Images.Selected()
			if selected == nil {
				clearDialogState(m)
				return m, nil
			}
			request.Source = fullImageRef(selected)
			switch operation {
			case runtimeapi.ImageTransferTag, runtimeapi.ImageTransferPush:
				request.Destination = value
			case runtimeapi.ImageTransferSave:
				request.Path = filepath.Clean(value)
			}
		}
		return beginImageTransfer(m, request)
	default:
		editQueryInput(key, &m.Dialog.Input)
		return m, nil
	}
}

func beginImageTransfer(m *state.AppModel, request runtimeapi.ImageTransferRequest) (*state.AppModel, tea.Cmd) {
	targetID := request.Source
	if targetID == "" {
		targetID = request.Path
	}
	trace := beginAudit(m, "resource.image."+string(request.Operation), audit.ImageTarget{ID: targetID, Name: targetID}, imageTransferDescription(request))
	ctx, generation := m.ImageTransfer.Begin(request, trace)
	m.Dialog.Open(state.DialogSpec{Kind: state.DialogImageTransfer, Title: i18n.T("image.transfer.title", imageTransferOperationLabel(request.Operation)), Body: imageTransferTarget(request)})
	m.Navigation.Mode = state.ModeImageTransfer
	return m, startImageTransferCmd(ctx, generation, m.Connection.Docker.ImageTransfers(), request)
}

func startImageTransferCmd(ctx context.Context, generation uint64, service runtimeapi.ImageTransferService, request runtimeapi.ImageTransferRequest) tea.Cmd {
	return func() tea.Msg {
		events, err := service.Run(ctx, request)
		if err != nil {
			return state.ImageTransferReceived{Generation: generation, Event: runtimeapi.ImageTransferEvent{Error: err, Done: true}}
		}
		return ReadImageTransferCmd(generation, events)()
	}
}

// ReadImageTransferCmd waits for the next progress or terminal transfer event.
func ReadImageTransferCmd(generation uint64, events <-chan runtimeapi.ImageTransferEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			event = runtimeapi.ImageTransferEvent{Error: fmt.Errorf("image transfer ended without a result"), Done: true}
		}
		return state.ImageTransferReceived{Generation: generation, Event: event, Events: events}
	}
}

func cancelImageTransfer(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	trace := m.ImageTransfer.Audit
	m.ImageTransfer.Reset()
	m.Dialog.Close()
	m.Navigation.Mode = state.ModeNormal
	message := i18n.T("image.transfer.cancelled")
	FinishAudit(m, trace, audit.ResultCancelled, message, audit.Details{})
	ShowToastWarn(m, message)
	return m, nil
}

func imageTransferOperationLabel(operation runtimeapi.ImageTransferOperation) string {
	return i18n.T("key." + string(operation))
}

func imageTransferDescription(request runtimeapi.ImageTransferRequest) string {
	return "Image " + string(request.Operation) + " " + imageTransferTarget(request)
}

func imageTransferTarget(request runtimeapi.ImageTransferRequest) string {
	if request.Path != "" {
		return request.Path
	}
	return request.Destination
}
