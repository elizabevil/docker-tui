package keyboard

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
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
	return state.ComposeProjectVolumes(m, project)
}

func composeProjectNetworks(m *state.AppModel, project string) []string {
	return state.ComposeProjectNetworks(m, project)
}

// selectedDetailComposeService resolves the service cursor inside the
// compose detail panel. Distinct from state.SelectedComposeService
// (which targets the panel-level service pane): this one requires
// ComposeDetailProject to be set so the detail / log / top views can
// route to a specific service within the project the user drilled into.
func selectedDetailComposeService(m *state.AppModel, project string) string {
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

func doComposeDown(m *state.AppModel, dialogTrace ...audit.Trace) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	// R08-04 F2: down opens a confirm dialog carrying three checkboxes
	// (-v / --rmi / --remove-orphans). The dialog's Confirm callback
	// copies the Checked states onto m.Compose.ComposeDownRemove* and
	// re-enters this function with a non-nil skipConfirm flag.
	if !m.Compose.ComposeDownSkipConfirm {
		return openComposeDownConfirm(m, project)
	}
	m.Compose.ComposeDownSkipConfirm = false
	containers := composeProjectContainers(m, project)
	volumes := composeProjectVolumes(m, project)
	networks := composeProjectNetworks(m, project)
	target := audit.ComposeTarget{Name: project, Meta: audit.ComposeMeta{Containers: len(containers), Volumes: len(volumes), Networks: len(networks)}}
	trace := audit.Trace{}
	if len(dialogTrace) > 0 {
		trace = dialogTrace[0]
	}
	if !trace.Valid() {
		trace = beginAudit(m, "resource.compose_project.down", target, "Removing compose project "+project)
	}
	if len(containers) == 0 && len(volumes) == 0 && len(networks) == 0 {
		FinishAudit(m, trace, audit.ResultFailed, "Compose down failed: no resources", audit.Details{Error: "no compose resources"})
		return m, nil
	}
	engine := m.Connection.Engine
	return m, func() tea.Msg {
		// R08-04 验收 5: --rmi removes the project's images. "all" removes
		// every image reference used by the project's containers; "local"
		// keeps images that carry the com.docker.compose.image label
		// (pulled registry images) and removes only compose-built ones.
		removeImages := m.Compose.ComposeDownRemoveImages
		imageRefs := composeProjectImageRefs(containers, removeImages == "local")
		result := state.BatchActioned{
			Scope:    ComposeScope(composeVerbDown),
			Resource: state.ResourceComposeProject,
			Total:    len(containers) + len(volumes) + len(networks) + len(imageRefs),
			Audit:    trace,
		}
		var failures []error
		// R08-04 F2: --remove-orphans filters containers whose service
		// label is not in the project's aggregated service set. When
		// unchecked (default) the orphans are kept; when checked they
		// are removed alongside the project's own containers.
		removeVolumes := m.Compose.ComposeDownRemoveVolumes
		keepOrphans := !m.Compose.ComposeDownRemoveOrphans
		var orphanIDs map[string]struct{}
		if keepOrphans {
			services := composeServiceNames(m, project)
			serviceSet := make(map[string]struct{}, len(services))
			for _, s := range services {
				serviceSet[s] = struct{}{}
			}
			orphanIDs = make(map[string]struct{})
			for _, c := range containers {
				if _, ok := serviceSet[c.ComposeService]; ok {
					continue
				}
				orphanIDs[c.ID] = struct{}{}
			}
		}
		for _, c := range containers {
			if keepOrphans {
				if _, isOrphan := orphanIDs[c.ID]; isOrphan {
					result.Skipped++
					continue
				}
			}
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
		if removeVolumes {
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
		for _, ref := range imageRefs {
			msg := imageRemoveCmd(engine, ref, true, false, nil)()
			if v, ok := msg.(state.ImageActioned); ok {
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, ref)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			} else {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, ref)
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}

// openComposeDownConfirm builds the down dialog (R08-04 F2) with three
// checkbox options plus the standard Cancel / Confirm pair. The
// default for `--remove-orphans` is checked; `-v` / `--rmi` default
// unchecked per spec.
func openComposeDownConfirm(m *state.AppModel, project string) (*state.AppModel, tea.Cmd) {
	target := audit.ComposeTarget{Name: project}
	trace := beginAudit(m, "resource.compose_project.down", target, "Compose down: "+project)
	options := []state.ChoiceOption{
		{ID: "opt_volumes", Label: "-v delete volumes", Checked: m.Compose.ComposeDownRemoveVolumes},
		{ID: "opt_rmi", Label: "--rmi all delete images", Checked: m.Compose.ComposeDownRemoveImages != ""},
		{ID: "opt_orphans", Label: "--remove-orphans", Checked: m.Compose.ComposeDownRemoveOrphans},
		{ID: keys.ShowOptionCancel, Label: "Cancel"},
		{ID: keys.ShowOptionConfirm, Label: "Confirm"},
	}
	m.Confirm.OpenWithOptions(keys.ShowOptionConfirm, project, "Down compose project "+project, trace, options)
	m.Confirm.Focus = 0 // start on the first checkbox, not Cancel
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func doComposeLogs(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	service := selectedDetailComposeService(m, project)
	containers := composeProjectContainers(m, project)
	if len(containers) == 0 {
		ShowToastNow(m, "✕ compose logs failed: no containers")
		return m, nil
	}
	// Scope the fetch to the selected service when the right pane is
	// focused; otherwise aggregate every container in the project.
	var picked []runtime.ContainerSummary
	if service != "" {
		for _, c := range containers {
			if c.ComposeService == service {
				picked = append(picked, c)
			}
		}
		if len(picked) == 0 {
			ShowToastWarn(m, fmt.Sprintf("✕ compose logs: service %q has no containers", service))
			return m, nil
		}
	} else {
		picked = containers
	}
	virtualID := composeLogVirtualID(project, service)
	m.Log.Open(virtualID)
	m.Navigation.Mode = state.ModeLogView
	cfg := m.Dependencies.Config.Logs
	label := project
	if service != "" {
		label = project + "/" + service
	}
	ShowToastNow(m, fmt.Sprintf("✓ compose logs %s (%d sources)", label, len(picked)))
	return m, FetchComposeLogsBatch(m.Connection.Engine, virtualID, picked, cfg.Since, cfg.Tail, cfg.Timestamps)
}

// composeLogVirtualID returns the synthetic LogContainerID used by the
// aggregated compose log view. The ID encodes the project + optional
// service so the Log page can label itself and so subsequent calls can
// distinguish scope without a separate state field.
func composeLogVirtualID(project, service string) string {
	if service != "" {
		return "compose:" + project + ":" + service
	}
	return "compose:" + project
}

// FetchComposeLogsBatch reads each container's log batch in sequence and
// returns a single LogBatchReceived whose lines are prefixed with
// "[service] " (single replica) or "[service.N] " (Nth replica of the
// service) per R08-06 F4. Service / replica counters come from the
// supplied containers slice; ordering within a service is preserved
// but cross-service lines are emitted in fetch order (R08-06 F1
// simplification — true timestamp merging needs streaming, deferred).
//
// Failures for individual containers are surfaced as
// LogStreamError for each; the Update loop filters those out so a
// single dropped container does not abort the whole view (F6).
func FetchComposeLogsBatch(client runtime.Engine, virtualID string, containers []runtime.ContainerSummary, since, tail string, ts bool) tea.Cmd {
	return func() tea.Msg {
		// Stable order so replicas keep their assigned suffix across runs.
		ordered := append([]runtime.ContainerSummary(nil), containers...)
		sort.SliceStable(ordered, func(i, j int) bool {
			if ordered[i].ComposeService != ordered[j].ComposeService {
				return ordered[i].ComposeService < ordered[j].ComposeService
			}
			return ordered[i].ID < ordered[j].ID
		})
		replicaIdx := make(map[string]int, len(ordered))
		for _, c := range ordered {
			replicaIdx[c.ComposeService]++
		}
		seen := make(map[string]int, len(ordered))
		var allLines []string
		var streamErrs []state.LogStreamError
		for _, c := range ordered {
			seen[c.ComposeService]++
			idx := seen[c.ComposeService]
			prefix := composeLogPrefix(c, replicaIdx[c.ComposeService], idx)
			reader, err := client.Containers().Logs(context.Background(), c.ID, runtime.ContainerLogOptions{Since: since, Tail: tail, Timestamps: ts})
			if err != nil {
				streamErrs = append(streamErrs, state.LogStreamError{ContainerID: c.ID, Error: err})
				continue
			}
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				allLines = append(allLines, prefix+" "+line)
			}
			if cerr := scanner.Err(); cerr != nil {
				streamErrs = append(streamErrs, state.LogStreamError{ContainerID: c.ID, Error: cerr})
			}
			_ = reader.Close() //nolint:errcheck // log reader fully scanned before close.
		}
		if len(streamErrs) > 0 {
			return state.ComposeLogBatchReceived{
				ContainerID: virtualID,
				Lines:       allLines,
				StreamErrs:  streamErrs,
			}
		}
		return state.LogBatchReceived{ContainerID: virtualID, Lines: allLines}
	}
}

// composeLogPrefix formats the per-line service tag. Single-replica
// services use "[service]"; multi-replica services use "[service.N]"
// so the user can distinguish outputs that share a service name.
func composeLogPrefix(c runtime.ContainerSummary, total, idx int) string {
	service := c.ComposeService
	if service == "" {
		service = "unknown"
	}
	if total <= 1 {
		return "[" + service + "]"
	}
	return fmt.Sprintf("[%s.%d]", service, idx)
}

func composeActionPending(m *state.AppModel, verb string) (*state.AppModel, tea.Cmd) {
	ShowToastWarn(m, fmt.Sprintf("✕ compose %s: not implemented yet", verb))
	return m, nil
}

// doComposeRestart chains Stop + Start on every container in the
// selected project. Each phase owns its own audit trace so the user
// sees stop + start as two distinct entries; restart itself has no
// dedicated trace (R08-04 F6: start / stop each carry their own).
func doComposeRestart(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	if len(containers) == 0 {
		ShowToastWarn(m, "✕ compose restart failed: no containers")
		return m, nil
	}
	ShowToastNow(m, fmt.Sprintf("⟳ restart %s (%d containers)", project, len(containers)))
	_, stopCmd := doComposeStop(m)
	_, startCmd := doComposeStart(m)
	if stopCmd == nil || startCmd == nil {
		return m, nil
	}
	return m, tea.Batch(stopCmd, startCmd)
}

func doComposeTop(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	service := selectedDetailComposeService(m, project)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	if len(containers) == 0 {
		ShowToastWarn(m, "✕ compose top failed: no containers")
		return m, nil
	}
	target := service
	if target == "" {
		target = "<all>"
	}
	ShowToastNow(m, fmt.Sprintf("⟳ compose top %s/%s", project, target))
	return m, fetchComposeServiceTop(m.Connection.Engine, project, service, containers)
}

func doComposePort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	service := selectedDetailComposeService(m, project)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	if len(containers) == 0 {
		ShowToastWarn(m, "✕ compose port failed: no containers")
		return m, nil
	}
	// 80 = the canonical "first published port" default; UI may override
	// later via a form. For now the user can switch to a single
	// container and use R01's :port command for a specific number.
	const defaultPort = 80
	m.Compose.ComposeSubview = state.ComposeSubviewPort
	ShowToastNow(m, fmt.Sprintf("⟳ compose port %s/%s :%d", project, service, defaultPort))
	return m, fetchComposeServicePort(m.Connection.Engine, project, service, defaultPort, containers)
}

func doComposeStats(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	if len(containers) == 0 {
		ShowToastWarn(m, "✕ compose stats failed: no containers")
		return m, nil
	}
	m.Compose.ComposeSubview = state.ComposeSubviewStats
	ShowToastNow(m, fmt.Sprintf("⟳ compose stats %s (%d sources)", project, len(containers)))
	return m, fetchComposeServiceStats(m.Connection.Engine, project, containers)
}

// closeComposeSubview clears the active subview and returns to the
// standard service list. Wired on Esc inside the compose panel.
func closeComposeSubview(m *state.AppModel) {
	m.Compose.ComposeSubview = state.ComposeSubviewServices
}

// fetchComposeServiceTop walks every container in the service scope
// and asks the runtime for its process table (R08-07 F1). Per-container
// failures land as Error on the carrier; the success path returns a
// single ComposeServiceTopLoaded message.
func fetchComposeServiceTop(eng runtime.Engine, project, service string, containers []runtime.ContainerSummary) tea.Cmd {
	return func() tea.Msg {
		items := make([]state.ComposeServiceTopItem, 0, len(containers))
		var firstErr error
		for _, c := range containers {
			if service != "" && c.ComposeService != service {
				continue
			}
			procs, err := eng.Compose().Top(context.Background(), project, c.ComposeService)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			items = append(items, state.ComposeServiceTopItem{
				ContainerID:   c.ID,
				ContainerName: c.Name,
				Processes:     procs,
			})
		}
		return state.ComposeServiceTopLoaded{
			Project: project,
			Service: service,
			Items:   items,
			Error:   firstErr,
		}
	}
}

// fetchComposeServicePort calls ComposeService.Port for each replica
// and surfaces every successful host binding as one port item (R08-07
// F2). The engine returns hostIP:hostPort per call; replicas expose
// the same binding so the result is naturally de-duped by the caller.
func fetchComposeServicePort(eng runtime.Engine, project, service string, port int, containers []runtime.ContainerSummary) tea.Cmd {
	return func() tea.Msg {
		var items []state.ComposeServicePortItem
		var firstErr error
		for _, c := range containers {
			if service != "" && c.ComposeService != service {
				continue
			}
			binding, err := eng.Compose().Port(context.Background(), project, c.ComposeService, port)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			if binding == "" {
				continue
			}
			items = append(items, state.ComposeServicePortItem{
				ContainerID:   c.ID,
				HostIP:        splitHostIP(binding),
				HostPort:      splitHostPort(binding),
				ContainerPort: uint16(port),
				Protocol:      "tcp",
			})
		}
		if firstErr != nil && len(items) == 0 {
			return state.ComposeServicePortLoaded{Project: project, Service: service, Port: port, Error: firstErr}
		}
		return state.ComposeServicePortLoaded{Project: project, Service: service, Port: port, Items: items, Error: firstErr}
	}
}

// fetchComposeServiceStats snapshots one round of stats for every
// running container in the project (R08-07 F3). The Update loop owns
// the polling cadence; the engine call is single-shot.
func fetchComposeServiceStats(eng runtime.Engine, project string, containers []runtime.ContainerSummary) tea.Cmd {
	return func() tea.Msg {
		items := make([]state.ComposeServiceStatsItem, 0, len(containers))
		var firstErr error
		for _, c := range containers {
			stats, err := eng.Compose().Stats(context.Background(), project, c.ComposeService)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				items = append(items, state.ComposeServiceStatsItem{
					ContainerID:   c.ID,
					ContainerName: c.Name,
					Error:         err,
				})
				continue
			}
			items = append(items, state.ComposeServiceStatsItem{
				ContainerID:   c.ID,
				ContainerName: c.Name,
				Stats:         stats,
			})
		}
		return state.ComposeServiceStatsLoaded{Project: project, Items: items, Error: firstErr}
	}
}

// splitHostIP / splitHostPort split a "hostIP:hostPort" string returned
// by ComposeService.Port. The podman / docker adapters both return this
// shape; we only care about the display fields.
func splitHostIP(binding string) string {
	i := strings.LastIndex(binding, ":")
	if i < 0 {
		return binding
	}
	return binding[:i]
}

func splitHostPort(binding string) string {
	i := strings.LastIndex(binding, ":")
	if i < 0 {
		return ""
	}
	return binding[i+1:]
}

// composeProjectImageRefs returns the deduplicated image references
// used by the given project containers. With localOnly=true (the
// `--rmi local` variant) references that carry the
// com.docker.compose.image label — images pulled from a registry by
// name — are kept, matching `docker compose down --rmi local`; with
// localOnly=false (`--rmi all`) every reference is returned.
func composeProjectImageRefs(containers []runtime.ContainerSummary, localOnly bool) []string {
	seen := make(map[string]struct{}, len(containers))
	out := make([]string, 0, len(containers))
	for _, c := range containers {
		ref := runtime.ComposeImageFromLabels(c)
		if ref == "" {
			continue
		}
		if localOnly && c.Labels != nil && c.Labels[runtime.ComposeLabelImage] != "" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

// doComposePause / Unpause / Kill issue the matching container action
// to every replica in the project (R08-09 F1). The aggregated
// BatchActioned mirrors doComposeStart/Stop so audit + Update paths
// stay uniform.
func doComposePause(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return doComposeProjectLifecycle(m, "pause", func(c runtime.Engine, id string) tea.Cmd {
		return containerPauseCmd(c, id, false)
	})
}

func doComposeUnpause(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return doComposeProjectLifecycle(m, "unpause", func(c runtime.Engine, id string) tea.Cmd {
		return containerPauseCmd(c, id, true)
	})
}

func doComposeKill(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return doComposeProjectLifecycle(m, "kill", func(c runtime.Engine, id string) tea.Cmd {
		return containerKillCmd(c, id)
	})
}

// doComposeRm removes the project's stopped containers. The "all"
// flag is intentionally not exposed in the keyboard path because the
// spec requires a Form / Confirm dialog (R08-09 F3); that lives in the
// UI layer. This verb is the bare default for Action Bar / shortcut
// callers that need no options.
func doComposeRm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return doComposeProjectLifecycle(m, "rm", func(c runtime.Engine, id string) tea.Cmd {
		return containerRemoveCmd(c, id, runtime.LifecycleOptions{Force: false})
	})
}

// doComposePrune mirrors docker compose prune semantics: every stopped
// container in the project + every project-named volume. The volume
// pass is best-effort (R08-09 F4 "scoped counts").
func doComposePrune(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	volumes := composeProjectVolumes(m, project)
	target := audit.ComposeTarget{Name: project, Meta: audit.ComposeMeta{Containers: len(containers), Volumes: len(volumes)}}
	trace := beginAudit(m, "resource.compose_project.prune", target, "Pruning compose project "+project)
	if len(containers) == 0 && len(volumes) == 0 {
		FinishAudit(m, trace, audit.ResultFailed, "Compose prune failed: no resources", audit.Details{Error: "no compose resources"})
		return m, nil
	}
	engine := m.Connection.Engine
	ShowToastNow(m, fmt.Sprintf("⟳ pruning %s (%d containers, %d volumes)", project, len(containers), len(volumes)))
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    ComposeScope(composeVerbPrune),
			Resource: state.ResourceComposeProject,
			Total:    len(containers) + len(volumes),
			Audit:    trace,
		}
		var failures []error
		for _, c := range containers {
			if c.State != state.ContainerStateExited && c.State != state.ContainerStateStopped && c.State != state.ContainerStateDead {
				result.Skipped++
				continue
			}
			msg := containerRemoveCmd(engine, c.ID, runtime.LifecycleOptions{Force: false})()
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
			msg := volumeRemoveCmd(engine, name, runtime.LifecycleOptions{Force: false})()
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
		result.Error = errors.Join(failures...)
		return result
	}
}

// doComposeProjectLifecycle is the shared body for Pause / Unpause /
// Kill / Rm: list every container in the project, dispatch the per-
// container cmd, aggregate into one BatchActioned. The audit trace is
// shared so the audit log shows one entry per project action.
func doComposeProjectLifecycle(m *state.AppModel, verb string, perContainer func(runtime.Engine, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	target := audit.ComposeTarget{Name: project, Meta: audit.ComposeMeta{Containers: len(containers)}}
	trace := beginAudit(m, "resource.compose_project."+verb, target, verb+" compose project "+project)
	if len(containers) == 0 {
		FinishAudit(m, trace, audit.ResultFailed, "Compose "+verb+" failed: no containers", audit.Details{Error: "no containers"})
		return m, nil
	}
	engine := m.Connection.Engine
	ShowToastNow(m, fmt.Sprintf("⟳ %s %s (%d containers)", verb, project, len(containers)))
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    ComposeScope(verb),
			Resource: state.ResourceComposeProject,
			Total:    len(containers),
			Audit:    trace,
		}
		var failures []error
		for _, c := range containers {
			cmd := perContainer(engine, c.ID)
			if cmd == nil {
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, c.ID)
				continue
			}
			msg := cmd()
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

// doComposeScale opens the R08-09 scale form. The form's submit path
// (submitContainerForm, FormComposeScale branch) calls
// executeComposeScale with the parsed replicas + --no-deps values.
func doComposeScale(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return openComposeScaleForm(m)
}

// parseScaleReplicas parses the IntField text into a non-negative
// replica count. Empty input defaults to 1 (the docker compose
// default for `up`); anything else must be a clean integer.
func parseScaleReplicas(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 1, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 || v > 100 {
		return 0, false
	}
	return v, true
}

// executeComposeScale brings the project's service to the target
// replica count (R08-09 F2):
//   - current == target → no-op toast
//   - target >  current → create (target - current) one-off containers
//     via ComposeService.Run with no entrypoint override; the engine
//     increments container-number labels.
//   - target <  current → stop + remove the oldest (target - current)
//     replicas, sorted by container-number ascending.
//
// --no-deps is intentionally ignored: dtui does not parse compose
// depends_on, so the engine's default behaviour (always start deps)
// is what users get. The flag is accepted for form compatibility only
// and has no effect on the executed command.
func executeComposeScale(m *state.AppModel, project, service string, target int, noDeps bool) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	current := 0
	for _, c := range containers {
		if c.ComposeService == service {
			current++
		}
	}
	target = max(target, 0)
	if target == current {
		ShowToastNow(m, fmt.Sprintf("✓ compose scale %s/%s already at %d", project, service, current))
		return m, nil
	}
	targetMetadata := audit.ComposeMeta{Containers: target}
	targetAudit := audit.ComposeTarget{Name: project, Meta: targetMetadata}
	trace := beginAudit(m, "resource.compose_service.scale", targetAudit, fmt.Sprintf("Scaling %s/%s to %d (current %d)", project, service, target, current))
	if target > current {
		toCreate := target - current
		ShowToastNow(m, fmt.Sprintf("⟳ compose scale %s/%s: creating %d replicas", project, service, toCreate))
		return m, func() tea.Msg {
			var failures []error
			successes := 0
			for i := 0; i < toCreate; i++ {
				err := m.Connection.Engine.Compose().Run(context.Background(), project, service, nil, runtime.RunOptions{RemoveAfter: false})
				if err != nil {
					failures = append(failures, fmt.Errorf("create %d: %w", i+1, err))
					continue
				}
				successes++
			}
			FinishAudit(m, trace, audit.ResultSucceeded, fmt.Sprintf("Scale %s/%s done", project, service), audit.Details{})
			return state.BatchActioned{
				Scope:    ComposeScope(composeVerbScale),
				Resource: state.ResourceComposeService,
				Total:    toCreate,
				Success:  successes,
				Failed:   len(failures),
				Error:    errors.Join(failures...),
				Audit:    trace,
			}
		}
	}
	toRemove := current - target
	replicas := make([]runtime.ContainerSummary, 0, current)
	for _, c := range containers {
		if c.ComposeService == service {
			replicas = append(replicas, c)
		}
	}
	sort.SliceStable(replicas, func(i, j int) bool {
		ni, errI := strconv.Atoi(replicas[i].ContainerNumber)
		nj, errJ := strconv.Atoi(replicas[j].ContainerNumber)
		if errI != nil || errJ != nil {
			return replicas[i].ID < replicas[j].ID
		}
		return ni < nj
	})
	drop := append([]runtime.ContainerSummary(nil), replicas[:toRemove]...)
	ShowToastNow(m, fmt.Sprintf("⟳ compose scale %s/%s: removing %d replicas", project, service, toRemove))
	return m, func() tea.Msg {
		var failures []error
		successes := 0
		for _, c := range drop {
			stopMsg := containerStopCmd(m.Connection.Engine, c.ID)()
			if v, ok := stopMsg.(state.ContainerActioned); ok && !v.Success && v.Error != nil {
				failures = append(failures, fmt.Errorf("stop %s: %w", c.Name, v.Error))
			}
			rmMsg := containerRemoveCmd(m.Connection.Engine, c.ID, runtime.LifecycleOptions{Force: true})()
			if v, ok := rmMsg.(state.ContainerActioned); ok {
				if v.Success {
					successes++
				} else {
					failures = append(failures, fmt.Errorf("rm %s: %w", c.Name, v.Error))
				}
			} else {
				failures = append(failures, fmt.Errorf("rm %s: unknown result", c.Name))
			}
		}
		FinishAudit(m, trace, audit.ResultSucceeded, fmt.Sprintf("Scale %s/%s done", project, service), audit.Details{})
		return state.BatchActioned{
			Scope:    ComposeScope(composeVerbScale),
			Resource: state.ResourceComposeService,
			Total:    toRemove,
			Success:  successes,
			Failed:   len(failures),
			Error:    errors.Join(failures...),
			Audit:    trace,
		}
	}
}

// doComposeEvents jumps to the R05 events panel with a project-scoped
// label filter pre-filled (R08-10 F2). No new event source is created:
// the existing subscription in EventState keeps streaming; we only
// narrow the filter and reset the cursor.
func doComposeEvents(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	m.EventPanel.SetFilterFromCompose(project)
	m.Navigation.Mode = state.ModeEvents
	ShowToastNow(m, fmt.Sprintf("✓ events filtered to %s", project))
	return m, nil
}

// doComposeServiceExec routes to R01's exec dialog: pick the first
// running container in the selected service, position the cursor on
// it, then defer to doAutoExecAction. R08-08 F1: "exec 复用 R01".
func doComposeServiceExec(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	service := selectedDetailComposeService(m, project)
	if project == "" || service == "" {
		ShowToastWarn(m, "✕ compose exec: pick a service first")
		return m, nil
	}
	containers := composeProjectContainers(m, project)
	var pick runtime.ContainerSummary
	for _, c := range containers {
		if c.ComposeService == service && c.State == state.ContainerStateRunning {
			pick = c
			break
		}
	}
	if pick.ID == "" {
		ShowToastWarn(m, fmt.Sprintf("✕ compose exec %s: no running container", service))
		return m, nil
	}
	for i, c := range m.Resources.Containers.Items {
		if c.ID == pick.ID {
			m.Resources.Containers.Cursor = i
			break
		}
	}
	m.Navigation.ActivePanel = state.PanelContainers
	ShowToastNow(m, fmt.Sprintf("✓ compose exec %s/%s → %s", project, service, pick.Name))
	return doAutoExecAction(m)
}

// doComposeServiceRun creates a oneoff container for the selected
// service via ComposeService.Run (R08-08 F2). The form fields
// (command, --rm, entrypoint) are surfaced as toast hints for now;
// the full Form dialog is left to the UI overhaul.
func doComposeServiceRun(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	service := selectedDetailComposeService(m, project)
	if project == "" || service == "" {
		ShowToastWarn(m, "✕ compose run: pick a service first")
		return m, nil
	}
	ShowToastNow(m, fmt.Sprintf("⟳ compose run %s/%s (defaults: --rm)", project, service))
	return m, fetchComposeServiceRun(m.Connection.Engine, project, service, runtime.RunOptions{
		RemoveAfter: true,
	})
}

func fetchComposeServiceRun(eng runtime.Engine, project, service string, opts runtime.RunOptions) tea.Cmd {
	return func() tea.Msg {
		err := eng.Compose().Run(context.Background(), project, service, nil, opts)
		return state.ComposeServiceRunCompleted{Project: project, Service: service, Error: err}
	}
}

// doComposeServiceLogs reuses FetchComposeLogsBatch with the service
// filter applied — the same engine call but with service-scoped
// containers. R08-06 F4 "service-level logs".
func doComposeServiceLogs(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return doComposeLogs(m)
}

// doComposeGroupDown filters the project containers by the active
// CoLocated group (m.Compose.ComposeGroupID) and routes them through
// doComposeDown. Empty group ID = full project — matches the spec's
// "默认行为不变" rule (R08-14 §R08-04 集成).
func doComposeGroupDown(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeGroupContainers(m, project, m.Compose.ComposeGroupID)
	if len(containers) == 0 {
		ShowToastWarn(m, "✕ compose group.down: no containers in group")
		return m, nil
	}
	return doComposeDown(m)
}

// doComposeGroupRestart chains Stop + Start within the active group.
func doComposeGroupRestart(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeGroupContainers(m, project, m.Compose.ComposeGroupID)
	if len(containers) == 0 {
		ShowToastWarn(m, "✕ compose group.restart: no containers in group")
		return m, nil
	}
	return doComposeRestart(m)
}

// doComposeGroupExec picks the first running container in the group
// and routes to the R01 exec dialog (R08-14 §F6).
func doComposeGroupExec(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	project := currentComposeProject(m)
	if project == "" {
		return m, nil
	}
	containers := composeGroupContainers(m, project, m.Compose.ComposeGroupID)
	var pick runtime.ContainerSummary
	for _, c := range containers {
		if c.State == state.ContainerStateRunning {
			pick = c
			break
		}
	}
	if pick.ID == "" {
		ShowToastWarn(m, fmt.Sprintf("✕ compose group.exec %s: no running container in group", m.Compose.ComposeGroupID))
		return m, nil
	}
	for i, c := range m.Resources.Containers.Items {
		if c.ID == pick.ID {
			m.Resources.Containers.Cursor = i
			break
		}
	}
	m.Navigation.ActivePanel = state.PanelContainers
	ShowToastNow(m, fmt.Sprintf("✓ compose group.exec %s → %s", m.Compose.ComposeGroupID, pick.Name))
	return doAutoExecAction(m)
}

// composeGroupContainers returns the project's containers that match
// the supplied CoLocated group ID. Empty group ID = every container
// (R08-14 §R08-04 集成: "默认行为不变").
func composeGroupContainers(m *state.AppModel, project, groupID string) []runtime.ContainerSummary {
	containers := composeProjectContainers(m, project)
	if groupID == "" {
		return containers
	}
	out := make([]runtime.ContainerSummary, 0, len(containers))
	for _, c := range containers {
		if c.CoLocatedGroupID == groupID {
			out = append(out, c)
		}
	}
	return out
}
