package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func historyKeyModel() *state.AppModel {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.Mode = state.ModeHistory
	m.Viewport.Height = 40
	m.History.Open("sha256:test", "example:test")
	m.History.Apply([]runtimeapi.ImageHistoryLayer{
		{CreatedBy: "FROM alpine"},
		{CreatedBy: "RUN apk add curl", Comment: "packages"},
		{CreatedBy: "CMD server"},
	}, runtimeapi.ImageHistoryLayerAPI, nil)
	return m
}

func TestHistoryModeDispatchesNavigationAndClose(t *testing.T) {
	m := historyKeyModel()
	m, _, handled := dispatchByMode(keys.KeyDown, keys.KeyDown, m)
	if !handled || m.History.Cursor != 1 {
		t.Fatalf("down: handled=%v cursor=%d", handled, m.History.Cursor)
	}
	m, _, handled = dispatchByMode(keys.KeyGUpper, keys.KeyGUpper, m)
	if !handled || m.History.Cursor != 2 {
		t.Fatalf("G: handled=%v cursor=%d", handled, m.History.Cursor)
	}
	m, _, handled = dispatchByMode(keys.KeyEsc, keys.KeyEsc, m)
	if !handled || m.Navigation.Mode != state.ModeNormal || m.History.ImageID != "" {
		t.Fatalf("Esc did not close history: mode=%v history=%+v", m.Navigation.Mode, m.History)
	}
}

func TestHistoryInlineFilter(t *testing.T) {
	m := historyKeyModel()
	m, _ = handleHistoryKeys(keys.KeySlash, m)
	if !m.History.Filtering {
		t.Fatal("slash did not enter history filter")
	}
	for _, key := range []string{"c", "u", "r", "l"} {
		m, _ = handleHistoryKeys(key, m)
	}
	if m.History.Filter != "curl" || len(historyVisibleItems(m)) != 1 {
		t.Fatalf("filter=%q visible=%d", m.History.Filter, len(historyVisibleItems(m)))
	}
	m, _ = handleHistoryKeys(keys.KeyBackspace, m)
	if m.History.Filter != "cur" {
		t.Fatalf("backspace filter=%q", m.History.Filter)
	}
	m, _ = handleHistoryKeys(keys.KeyEnter, m)
	if m.History.Filtering || m.Navigation.Mode != state.ModeHistory {
		t.Fatalf("enter should apply inline filter: filtering=%v mode=%v", m.History.Filtering, m.Navigation.Mode)
	}
}
