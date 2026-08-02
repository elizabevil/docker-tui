package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func eventKeyModel() *state.AppModel {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	m.Viewport.Height = 40
	m.EventPanel.AppendItems([]runtimeapi.EventItem{
		{Event: runtimeapi.Event{ResourceType: "container", Action: "start"}},
		{Event: runtimeapi.Event{ResourceType: "image", Action: "pull"}},
	})
	return m
}

func TestEventsOpenFilterPauseAndReturn(t *testing.T) {
	m := eventKeyModel()
	m.Navigation.Mode = state.ModeDetail
	m, _ = openEventsPage(m)
	if m.Navigation.Mode != state.ModeEvents || m.EventPanel.PreviousMode != state.ModeDetail {
		t.Fatalf("events open mode=%v previous=%v", m.Navigation.Mode, m.EventPanel.PreviousMode)
	}
	m, _ = handleEventPanelKeys(keys.KeySpace, m)
	if !m.EventPanel.Paused {
		t.Fatal("space did not pause events")
	}
	m, _ = handleEventPanelKeys(keys.KeySlash, m)
	m, _ = handleEventPanelKeys("i", m)
	m, _ = handleEventPanelKeys("m", m)
	m, _ = handleEventPanelKeys("a", m)
	m, _ = handleEventPanelKeys("g", m)
	m, _ = handleEventPanelKeys("e", m)
	if got := len(m.EventPanel.FilteredEvents()); got != 1 {
		t.Fatalf("filtered events=%d, want 1", got)
	}
	m, _ = handleEventPanelKeys(keys.KeyEnter, m)
	m, _ = handleEventPanelKeys(keys.KeyEsc, m)
	if m.Navigation.Mode != state.ModeDetail {
		t.Fatalf("Esc returned to mode=%v, want detail", m.Navigation.Mode)
	}
}

func TestEventsClearUsesConfirmationAndReturns(t *testing.T) {
	m := eventKeyModel()
	m.Navigation.Mode = state.ModeEvents
	m, _ = handleEventPanelKeys(keys.KeyCtrlD, m)
	if m.Navigation.Mode != state.ModeConfirm || m.Confirm.ReturnMode != state.ModeEvents {
		t.Fatalf("clear confirmation mode=%v return=%v", m.Navigation.Mode, m.Confirm.ReturnMode)
	}
	m.Confirm.Focus = 1
	m, _ = handleConfirmKeys(keys.KeyEnter, m)
	if m.Navigation.Mode != state.ModeEvents || len(m.EventPanel.Events) != 0 {
		t.Fatalf("confirmed clear mode=%v events=%d", m.Navigation.Mode, len(m.EventPanel.Events))
	}
}

func TestEventsGlobalBindingOpensFromHistoryAndReturns(t *testing.T) {
	m := eventKeyModel()
	m.Navigation.Mode = state.ModeHistory

	m, _ = HandleKeyPress(tea.KeyPressMsg(tea.Key{Code: tea.KeyF3}), m)
	if m.Navigation.Mode != state.ModeEvents || m.EventPanel.PreviousMode != state.ModeHistory {
		t.Fatalf("F3 mode=%v previous=%v", m.Navigation.Mode, m.EventPanel.PreviousMode)
	}
	m, _ = HandleKeyPress(keyMessage(keys.KeyEsc), m)
	if m.Navigation.Mode != state.ModeHistory {
		t.Fatalf("Esc returned to mode=%v, want history", m.Navigation.Mode)
	}
}
