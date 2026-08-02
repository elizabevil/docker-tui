package keyboard

import (
	"context"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

const fetchTimeout = 5 * time.Second

func openHistoryPage(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		ShowToastWarn(m, i18n.T("history.no_engine"))
		return m, nil
	}
	if m.Navigation.ActivePanel != state.PanelImages {
		return m, nil
	}
	summary := m.Resources.Images.Selected()
	if summary == nil {
		ShowToastWarn(m, i18n.T("history.no_image"))
		return m, nil
	}
	if summary.IsManifest {
		ShowToastNow(m, i18n.T("history.manifest_notice"))
		return m, nil
	}
	ref := fullImageRef(summary)
	m.History.Open(summary.ID, ref)
	m.Navigation.Mode = state.ModeHistory
	return m, historyFetchCmd(m.Connection.Engine, *summary)
}

func historyFetchCmd(engine dockerclient.Engine, summary dockerclient.ImageSummary) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		layers, err := engine.Images().History(ctx, summary)
		return state.HistoryLoadedMsg{
			ImageID: summary.ID,
			Layers:  layers,
			Err:     err,
		}
	}
}
