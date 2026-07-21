package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func composeProjectContainers(m *state.AppModel, project string) []docker.ContainerSummary {
	matched := make([]docker.ContainerSummary, 0, 8)
	for _, c := range m.Containers.Items {
		if c.ComposeProject == project {
			matched = append(matched, c)
		}
	}
	return matched
}

func composeProjectVolumes(m *state.AppModel, project string) []string {
	vols := make([]string, 0, 8)
	for _, v := range m.Volumes.Items {
		if v.Labels == nil {
			continue
		}
		if v.Labels["com.docker.compose.project"] == project {
			vols = append(vols, v.Name)
		}
	}
	return vols
}

func composeProjectNetworks(m *state.AppModel, project string) []string {
	nets := make([]string, 0, 8)
	for _, n := range m.Networks.Items {
		if n.Labels == nil {
			continue
		}
		if n.Labels["com.docker.compose.project"] == project {
			nets = append(nets, n.ID)
		}
	}
	return nets
}

func selectedComposeService(m *state.AppModel, project string) string {
	if m.ComposeDetailProject == "" {
		return ""
	}
	services := composeServiceNames(m, project)
	if len(services) == 0 {
		return ""
	}
	if m.ComposeServiceCursor >= len(services) {
		m.ComposeServiceCursor = len(services) - 1
	}
	if m.ComposeServiceCursor < 0 {
		m.ComposeServiceCursor = 0
	}
	return services[m.ComposeServiceCursor]
}

func doComposeStart(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
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
	cmds := make([]tea.Cmd, 0, len(containers))
	for _, c := range containers {
		cmds = append(cmds, withContainerAudit(containerStartCmd(m.Docker, c.ID), trace))
	}
	return m, tea.Batch(cmds...)
}

func doComposeStop(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
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
	cmds := make([]tea.Cmd, 0, len(containers))
	for _, c := range containers {
		cmds = append(cmds, withContainerAudit(containerStopCmd(m.Docker, c.ID), trace))
	}
	return m, tea.Batch(cmds...)
}

func doComposeDown(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
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
	cmds := make([]tea.Cmd, 0, len(containers)+len(volumes)+len(networks))
	for _, c := range containers {
		cmds = append(cmds, withContainerAudit(containerRemoveCmd(m.Docker, c.ID, true), trace))
	}
	for _, name := range volumes {
		cmds = append(cmds, withGenericAudit(volumeRemoveCmd(m.Docker, name, true), trace))
	}
	for _, id := range networks {
		cmds = append(cmds, withGenericAudit(networkRemoveCmd(m.Docker, id), trace))
	}
	return m, tea.Batch(cmds...)
}

func doComposeLogs(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
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
	m.LogContainerID = picked.ID
	m.LogContent = m.LogContent[:0]
	m.LogViewOffset = 0
	m.Mode = state.ModeLogView
	cfg := m.Config.Logs
	ShowToastNow(m, fmt.Sprintf("✓ compose logs %s/%s", project, picked.Name))
	return m, FetchLogBatch(m.Docker, picked.ID, cfg.Since, cfg.Tail, cfg.Timestamps)
}
