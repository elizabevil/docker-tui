package keyboard

import (
	"errors"
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func composeProjectContainers(m *state.AppModel, project string) []runtime.ContainerSummary {
	matched := make([]runtime.ContainerSummary, 0, 8)
	for _, c := range m.Resources.Containers.Items {
		if c.ComposeProject == project {
			matched = append(matched, c)
		}
	}
	return matched
}

func composeProjectVolumes(m *state.AppModel, project string) []string {
	vols := make([]string, 0, 8)
	for _, v := range m.Resources.Volumes.Items {
		if v.Labels == nil {
			continue
		}
		if v.Labels[runtime.ComposeLabelProject] == project {
			vols = append(vols, v.Name)
		}
	}
	return vols
}

func composeProjectNetworks(m *state.AppModel, project string) []string {
	nets := make([]string, 0, 8)
	for _, n := range m.Resources.Networks.Items {
		if n.Labels == nil {
			continue
		}
		if n.Labels[runtime.ComposeLabelProject] == project {
			nets = append(nets, n.ID)
		}
	}
	return nets
}

func selectedComposeService(m *state.AppModel, project string) string {
	if m.Compose.ComposeDetailProject == "" {
		return ""
	}
	services := composeServiceNames(m, project)
	if len(services) == 0 {
		return ""
	}
	if m.Compose.ComposeServiceCursor >= len(services) {
		m.Compose.ComposeServiceCursor = len(services) - 1
	}
	if m.Compose.ComposeServiceCursor < 0 {
		m.Compose.ComposeServiceCursor = 0
	}
	return services[m.Compose.ComposeServiceCursor]
}

func doComposeStart(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	target := audit.ComposeTarget{Name: project, Meta: audit.ComposeMeta{Containers: len(containers)}}
	trace := beginAudit(m, "resource.compose_project.start", target, "Starting compose project "+project)
	if len(containers) == 0 {
		FinishAudit(m, trace, audit.ResultFailed, "Compose start failed: no containers", audit.Details{Error: "no containers"})
		return m, nil
	}
	engine := m.Connection.Engine
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    ComposeScope(composeVerbStart),
			Resource: state.ResourceComposeProject,
			Total:    len(containers),
			Audit:    trace,
		}
		var failures []error
		for _, c := range containers {
			msg := containerStartCmd(engine, c.ID)()
			if v, ok := msg.(state.ContainerActioned); ok {
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, c.ID)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			} else {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, c.ID)
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}

func doComposeStop(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	target := audit.ComposeTarget{Name: project, Meta: audit.ComposeMeta{Containers: len(containers)}}
	trace := beginAudit(m, "resource.compose_project.stop", target, "Stopping compose project "+project)
	if len(containers) == 0 {
		FinishAudit(m, trace, audit.ResultFailed, "Compose stop failed: no containers", audit.Details{Error: "no containers"})
		return m, nil
	}
	engine := m.Connection.Engine
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    ComposeScope(composeVerbStop),
			Resource: state.ResourceComposeProject,
			Total:    len(containers),
			Audit:    trace,
		}
		var failures []error
		for _, c := range containers {
			msg := containerStopCmd(engine, c.ID)()
			if v, ok := msg.(state.ContainerActioned); ok {
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, c.ID)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			} else {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, c.ID)
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}

func doComposeDown(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	volumes := composeProjectVolumes(m, project)
	networks := composeProjectNetworks(m, project)
	target := audit.ComposeTarget{Name: project, Meta: audit.ComposeMeta{Containers: len(containers), Volumes: len(volumes), Networks: len(networks)}}
	trace := beginAudit(m, "resource.compose_project.down", target, "Removing compose project "+project)
	if len(containers) == 0 && len(volumes) == 0 && len(networks) == 0 {
		FinishAudit(m, trace, audit.ResultFailed, "Compose down failed: no resources", audit.Details{Error: "no compose resources"})
		return m, nil
	}
	engine := m.Connection.Engine
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    ComposeScope(composeVerbDown),
			Resource: state.ResourceComposeProject,
			Total:    len(containers) + len(volumes) + len(networks),
			Audit:    trace,
		}
		var failures []error
		for _, c := range containers {
			msg := containerRemoveCmd(engine, c.ID, runtime.LifecycleOptions{Force: true})()
			if v, ok := msg.(state.ContainerActioned); ok {
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, c.ID)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			} else {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, c.ID)
			}
		}
		for _, name := range volumes {
			msg := volumeRemoveCmd(engine, name, runtime.LifecycleOptions{Force: true})()
			if v, ok := msg.(state.GenericActioned); ok {
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, name)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			} else {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, name)
			}
		}
		for _, id := range networks {
			msg := networkRemoveCmd(engine, id)()
			if v, ok := msg.(state.GenericActioned); ok {
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, id)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			} else {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, id)
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}

func doComposeLogs(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	service := selectedComposeService(m, project)
	containers := composeProjectContainers(m, project)
	if len(containers) == 0 {
		ShowToastNow(m, "✕ compose logs failed: no containers")
		return m, nil
	}
	picked := containers[0]
	if service != "" {
		for _, c := range containers {
			if c.ComposeService == service {
				picked = c
				break
			}
		}
	}
	m.Log.Open(picked.ID)
	m.Navigation.Mode = state.ModeLogView
	cfg := m.Dependencies.Config.Logs
	ShowToastNow(m, fmt.Sprintf("✓ compose logs %s/%s", project, picked.Name))
	return m, FetchLogBatch(m.Connection.Engine, picked.ID, cfg.Since, cfg.Tail, cfg.Timestamps)
}
