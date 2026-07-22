package tui

import (
	"context"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

const (
	eventCoalesceWindow = 100 * time.Millisecond
	eventFallbackPeriod = 15 * time.Second
)

func startEventStream(m *state.AppModel) tea.Cmd {
	if m.Connection.Docker == nil {
		return nil
	}
	ctx, generation := m.Events.Begin()
	if !m.Connection.Docker.Capabilities().Supports(runtimeapi.CapabilityEvents) {
		m.Events.Degraded = true
		return eventFallbackCmd(generation)
	}
	return subscribeEventsCmd(ctx, generation, m.Connection.Docker.Events())
}

func subscribeEventsCmd(ctx context.Context, generation uint64, service runtimeapi.EventService) tea.Cmd {
	return func() tea.Msg {
		items, err := service.Subscribe(ctx, runtimeapi.EventOptions{})
		if err != nil {
			return state.EventStreamFailed{Generation: generation, Error: err}
		}
		return state.EventStreamReady{Generation: generation, Items: items}
	}
}

func readEventCmd(generation uint64, items <-chan runtimeapi.EventItem) tea.Cmd {
	return func() tea.Msg {
		item, ok := <-items
		if !ok {
			return state.EventStreamClosed{Generation: generation}
		}
		return state.RuntimeEventReceived{Generation: generation, Item: item, Items: items}
	}
}

func handleEventStreamReady(m *state.AppModel, msg state.EventStreamReady) (*state.AppModel, tea.Cmd) {
	if !m.Events.Current(msg.Generation) {
		return m, nil
	}
	m.Events.RecordReady()
	return m, readEventCmd(msg.Generation, msg.Items)
}

func handleRuntimeEvent(m *state.AppModel, msg state.RuntimeEventReceived) (*state.AppModel, tea.Cmd) {
	if !m.Events.Current(msg.Generation) {
		return m, nil
	}
	if msg.Item.Error != nil {
		return scheduleEventReconnect(m, msg.Generation, msg.Item.Error)
	}
	m.Events.RecordReady()
	token, scheduleFlush := m.Events.MarkDirty(msg.Item.Event.ResourceType)
	commands := []tea.Cmd{readEventCmd(msg.Generation, msg.Items)}
	if scheduleFlush {
		commands = append(commands, tea.Tick(eventCoalesceWindow, func(time.Time) tea.Msg {
			return state.EventFlush{Generation: msg.Generation, Token: token}
		}))
	}
	return m, tea.Batch(commands...)
}

func handleEventFlush(m *state.AppModel, msg state.EventFlush) (*state.AppModel, tea.Cmd) {
	if !m.Events.Current(msg.Generation) || m.Connection.Docker == nil {
		return m, nil
	}
	dirty := m.Events.TakeDirty(msg.Token)
	commands := make([]tea.Cmd, 0, len(dirty))
	if dirty["container"] {
		commands = append(commands, keyboard.FetchContainers(m.Connection.Docker, true))
	}
	if dirty["image"] {
		commands = append(commands, keyboard.FetchImages(m.Connection.Docker))
	}
	if dirty["volume"] {
		commands = append(commands, keyboard.FetchVolumes(m.Connection.Docker))
	}
	if dirty["network"] {
		commands = append(commands, keyboard.FetchNetworks(m.Connection.Docker))
	}
	return m, tea.Batch(commands...)
}

func handleEventStreamFailed(m *state.AppModel, generation uint64, err error) (*state.AppModel, tea.Cmd) {
	if !m.Events.Current(generation) {
		return m, nil
	}
	return scheduleEventReconnect(m, generation, err)
}

func scheduleEventReconnect(m *state.AppModel, generation uint64, _ error) (*state.AppModel, tea.Cmd) {
	delay := m.Events.RecordFailure()
	m.Events.RenewContext()
	return m, tea.Batch(
		tea.Tick(delay, func(time.Time) tea.Msg { return state.EventReconnect{Generation: generation} }),
		eventFallbackCmd(generation),
	)
}

func handleEventReconnect(m *state.AppModel, msg state.EventReconnect) (*state.AppModel, tea.Cmd) {
	if !m.Events.Current(msg.Generation) || m.Connection.Docker == nil {
		return m, nil
	}
	return m, subscribeEventsCmd(m.Events.Context, msg.Generation, m.Connection.Docker.Events())
}

func eventFallbackCmd(generation uint64) tea.Cmd {
	return tea.Tick(eventFallbackPeriod, func(time.Time) tea.Msg {
		return state.EventFallbackTick{Generation: generation}
	})
}

func handleEventFallbackTick(m *state.AppModel, msg state.EventFallbackTick) (*state.AppModel, tea.Cmd) {
	if !m.Events.Current(msg.Generation) || !m.Events.Degraded || m.Connection.Docker == nil {
		return m, nil
	}
	commands := keyboard.FetchAll(m.Connection.Docker)
	commands = append(commands, eventFallbackCmd(msg.Generation))
	return m, tea.Batch(commands...)
}
