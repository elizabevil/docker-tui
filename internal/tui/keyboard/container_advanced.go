package keyboard

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// Advanced container operations (TASK-019 / BR-033): Update / Diff / Export /
// Commit / Wait / Copy. This file owns the keyboard-layer Cmd, Msg and handler
// core. Forms and dialogs for the parameterised actions (Copy / Export /
// Update / Commit) are intentionally left as interface + state design only.
//
// The runtime adapters already expose all six operations through
// runtimeapi.ContainerService; see docker/service_container.go and
// podman/service_container.go.
//
// Delivered now:
//   - Diff and Wait: full Cmd, Msg and handler core (no user form required).
//   - Update / Commit / Export / Copy: Msg types, Cmd factories and the hidden
//     stream-teardown helper so the runtime contract is proven callable. Their
//     doXxx handlers still need dialog/form wiring and are documented below.
//
// Expected wiring (main model): add a case in keyboard/actions.go for each
// ActionContainer* key and dispatch the resulting tea.Msg in update/update.go.

// ---- Diff --------------------------------------------------------------

// ContainerDiffDone carries the filesystem changes between a container and
// its base image, or the failure that prevented the diff.
type ContainerDiffDone struct {
	ContainerID string
	Changes     []runtimeapi.ContainerDiffChange
	Error       error
	Audit       audit.Trace
}

// containerDiffCmd runs the engine diff and reports the result as a
// ContainerDiffDone message. Requires no form.
func containerDiffCmd(engine runtimeapi.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		changes, err := engine.Containers().Diff(context.Background(), id)
		return ContainerDiffDone{ContainerID: id, Changes: changes, Error: err}
	}
}

// doContainerDiff validates a selected container and runs a diff command.
func doContainerDiff(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	trace := beginAudit(m, "resource.container.diff", containerTarget(m, ctr.ID), "Diff container "+ctr.Name)
	return m, withAdvancedAudit(containerDiffCmd(m.Connection.Engine, ctr.ID), trace)
}

// -------------------------------------------------------------------------
// Wait ---------------------------------------------------------------

// ContainerWaitDone carries the exit status returned by a blocking wait, or
// the failure that aborted the wait.
type ContainerWaitDone struct {
	Generation  uint64
	ContainerID string
	Condition   string
	Result      runtimeapi.ContainerWaitResult
	Error       error
	Audit       audit.Trace
}

// containerWaitCmd blocks until the container meets Cond ("" or "not-running"
// by default, "next-exit" or "removed"; see runtimeapi.WaitOptions). The
// caller supplies a cancellable context (see state.ContainerWaitState) so the
// wait never outlives the app or the active runtime connection.
func containerWaitCmd(ctx context.Context, generation uint64, engine runtimeapi.Engine, id, condition string) tea.Cmd {
	return func() tea.Msg {
		result, err := engine.Containers().Wait(ctx, id, condition)
		return ContainerWaitDone{Generation: generation, ContainerID: id, Condition: condition, Result: result, Error: err}
	}
}

// doContainerWait starts a wait for the selected container using the default
// condition ("not-running" → blocks until the container exits). The wait's
// context is registered on m.ContainerWait so the integration layer can cancel
// it on runtime switch or app quit (m.ContainerWait.Stop()).
func doContainerWait(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	ctx, generation := m.ContainerWait.Begin()
	trace := beginAudit(m, "resource.container.wait", containerTarget(m, ctr.ID), "Wait for container "+ctr.Name)
	return m, withAdvancedAudit(containerWaitCmd(ctx, generation, m.Connection.Engine, ctr.ID, ""), trace)
}

// -------------------------------------------------------------------------
// Update ----------------------------------------------------------------

// ContainerUpdateDone carries the engine's response to an update, or the
// failure that rejected the change.
type ContainerUpdateDone struct {
	ContainerID string
	Result      runtimeapi.ContainerUpdateResult
	Error       error
	Audit       audit.Trace
}

// containerUpdateCmd applies resource-limit changes (memory, CPU, restart
// policy) to a running container. The form that gathers UpdateOptions is
// delegate design; the Cmd already matches the runtime contract.
func containerUpdateCmd(engine runtimeapi.Engine, id string, opts runtimeapi.ContainerUpdateOptions) tea.Cmd {
	return func() tea.Msg {
		result, err := engine.Containers().Update(context.Background(), id, opts)
		return ContainerUpdateDone{ContainerID: id, Result: result, Error: err}
	}
}

// -------------------------------------------------------------------------
// Commit
// -------------------------------------------------------------------------

// ContainerCommitDone carries the new image ID from a commit, or the failure.
type ContainerCommitDone struct {
	ContainerID string
	ImageID     string
	ExportPath  string
	Error       error
	Audit       audit.Trace
}

// containerCommitCmd snapshots the container as a new image. The commit form
// (repository/tag/comment/author/pause) is deferred to the dialog layer.
func containerCommitCmd(engine runtimeapi.Engine, id string, opts runtimeapi.ContainerCommitOptions, exportPath string) tea.Cmd {
	return func() tea.Msg {
		result, err := engine.Containers().Commit(context.Background(), id, opts)
		return ContainerCommitDone{ContainerID: id, ImageID: result.ID, ExportPath: exportPath, Error: err}
	}
}

// -------------------------------------------------------------------------
// Export & Copy (stream teardown core)
// -------------------------------------------------------------------------

// ContainerExportDone reports whether the container filesystem tar was written
// to Destination, or the failure.
type ContainerExportDone struct {
	ContainerID string
	Destination string
	Bytes       int64
	Error       error
	Audit       audit.Trace
}

// ContainerCopyDone reports whether a container file/dir was copied out to
// Destination, or the failure.
type ContainerCopyDone struct {
	ContainerID string
	SourcePath  string
	Destination string
	Bytes       int64
	Error       error
	Audit       audit.Trace
}

// copyStreamToFile drains rc into dst, always closing rc — even when the
// destination cannot be opened or the copy is interrupted, so the underlying
// engine FD is never leaked. The tar is first written to a temp file in the
// destination directory and atomically renamed into place only on success, so
// a failed copy never truncates an existing destination or leaves partial tar
// data behind. This is the shared stream-teardown core for Export and Copy.
func copyStreamToFile(rc io.ReadCloser, dst string) (int64, error) {
	if rc == nil {
		return 0, errors.New("runtime returned a nil stream")
	}
	defer func() { _ = rc.Close() }() //nolint:errcheck // always release the engine stream.

	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".dtui-stream-*")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		_ = tmp.Close() //nolint:errcheck // may already be closed on success.
		if !committed {
			_ = os.Remove(tmpName) //nolint:errcheck // best-effort temp cleanup.
		}
	}()

	n, err := io.Copy(tmp, rc)
	if err != nil {
		return n, err
	}
	if err := tmp.Close(); err != nil {
		return n, err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return n, err
	}
	committed = true
	return n, nil
}

// containerExportCmd streams the container filesystem to destination. The
// Export form (target tar path) is deferred; the Cmd already owns reader
// closure.
func containerExportCmd(engine runtimeapi.Engine, id, destination string) tea.Cmd {
	return func() tea.Msg {
		rc, err := engine.Containers().Export(context.Background(), id)
		if err != nil {
			return ContainerExportDone{ContainerID: id, Destination: destination, Error: err}
		}
		n, err := copyStreamToFile(rc, destination)
		return ContainerExportDone{ContainerID: id, Destination: destination, Bytes: n, Error: err}
	}
}

// containerCopyCmd extracts SourcePath (a container-side file or directory)
// into a tar written to Destination. Reader closure is owned here.
func containerCopyCmd(engine runtimeapi.Engine, id, srcPath, destination string) tea.Cmd {
	return func() tea.Msg {
		rc, err := engine.Containers().CopyFromContainer(context.Background(), id, srcPath)
		if err != nil {
			return ContainerCopyDone{ContainerID: id, SourcePath: srcPath, Destination: destination, Error: err}
		}
		n, err := copyStreamToFile(rc, destination)
		return ContainerCopyDone{ContainerID: id, SourcePath: srcPath, Destination: destination, Bytes: n, Error: err}
	}
}

// -------------------------------------------------------------------------
// Audit plumbing (mirrors withContainerAudit / withGenericAudit)
// -------------------------------------------------------------------------

// withAdvancedAudit attaches an audit trace to any terminal container-action
// message emitted by this package, analogous to withContainerAudit.
func withAdvancedAudit(cmd tea.Cmd, trace audit.Trace) tea.Cmd {
	if !trace.Valid() {
		return cmd
	}
	return func() tea.Msg {
		msg := cmd()
		switch v := msg.(type) {
		case ContainerDiffDone:
			v.Audit = trace
			return v
		case ContainerWaitDone:
			v.Audit = trace
			return v
		case ContainerUpdateDone:
			v.Audit = trace
			return v
		case ContainerCommitDone:
			v.Audit = trace
			return v
		case ContainerExportDone:
			v.Audit = trace
			return v
		case ContainerCopyDone:
			v.Audit = trace
			return v
		}
		return msg
	}
}
