package keyboard

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// Form field keys shared between the keyboard handlers and the renderer.
const (
	fieldSourcePath      = "source"
	fieldDestinationPath = "destination"
	fieldMemory          = "memory"
	fieldCPUs            = "cpus"
	fieldRestartPolicy   = "restart"
	fieldMaxRetries      = "maxRetries"
	fieldRepository      = "repository"
	fieldTag             = "tag"
	fieldAuthor          = "author"
	fieldComment         = "comment"
	fieldPause           = "pause"
	fieldExportTar       = "exportTar"
	fieldArchivePath     = "archivePath"
	fieldImagePath       = "imagePath"
)

const restartPolicyUnchanged = "unchanged"

// restartPolicyChoices are the Docker/Podman restart-policy names. The leading
// "unchanged" option keeps the current effective policy when no choice is made.
var restartPolicyChoices = []string{restartPolicyUnchanged, "no", "always", "unless-stopped", "on-failure"}

// openContainerCopyForm opens the Copy form for the selected container. The
// source is a container-side file or directory; the destination is a local tar
// path on the dtui machine (BR-041 §3.1).
func openContainerCopyForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	cwd := workingDir()
	shortID := shortContainerID(ctr.ID)
	m.Form.Open(state.FormSpec{
		Kind:       state.FormContainerCopy,
		Title:      i18n.T("container.copy.form.title"),
		TargetID:   ctr.ID,
		TargetName: ctr.Name,
		CWD:        cwd,
		Fields: []state.FormField{
			{Key: fieldSourcePath, Label: i18n.T("container.copy.form.source"), Kind: state.FormPath, PathSource: state.PathContainer, PathMode: state.PathAny, Required: true},
			{Key: fieldDestinationPath, Label: i18n.T("container.copy.form.destination"), Kind: state.FormPath, PathSource: state.PathLocal, PathMode: state.PathSaveFile, Required: true},
		},
	})
	// 打开时生成默认本地目标名（BR-041 §4.1）。目标为 PathSaveFile，
	// 用户未手工编辑前，源路径变化会触发重新生成；已编辑则不覆盖。
	prefillDefaultDestination(m, cwd, shortID)
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

func openImageSaveForm(m *state.AppModel, image *runtimeapi.ImageSummary) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelImages || image == nil {
		return m, nil
	}
	cwd := workingDir()
	ref := fullImageRef(image)
	shortID := strings.TrimPrefix(image.ID, "sha256:")
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	path := state.DefaultImageSaveName(cwd, ref, shortID, time.Now())
	field := state.FormField{Key: fieldImagePath, Label: i18n.T("image.save.form.path"), Kind: state.FormPath, PathSource: state.PathLocal, PathMode: state.PathSaveFile, Required: true}
	field.Input.Set(path)
	m.Form.Open(state.FormSpec{
		Kind: state.FormImageSave, Title: i18n.T("image.save.title"), TargetID: ref,
		TargetName: tagName(image), CWD: cwd, Fields: []state.FormField{field},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

func openImageLoadForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelImages {
		return m, nil
	}
	m.Form.Open(state.FormSpec{
		Kind: state.FormImageLoad, Title: i18n.T("image.load.title"), CWD: workingDir(),
		Fields: []state.FormField{{Key: fieldImagePath, Label: i18n.T("image.load.form.path"), Kind: state.FormPath, PathSource: state.PathLocal, PathMode: state.PathFile, Required: true}},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

// openContainerExportForm opens the Export form: the local destination tar path.
func openContainerExportForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	cwd := workingDir()
	m.Form.Open(state.FormSpec{
		Kind:       state.FormContainerExport,
		Title:      i18n.T("container.export.form.title"),
		TargetID:   ctr.ID,
		TargetName: ctr.Name,
		CWD:        cwd,
		Fields: []state.FormField{
			{Key: fieldDestinationPath, Label: i18n.T("container.export.form.destination"), Kind: state.FormPath, PathSource: state.PathLocal, PathMode: state.PathSaveFile, Required: true},
		},
	})
	// Export 打开时直接生成默认本地目标名（BR-041 §4.2）。
	if dst := m.Form.Get(fieldDestinationPath); dst != nil && dst.Input.Text == "" {
		dst.Input.Set(state.DefaultExportName(cwd, ctr.Name, shortContainerID(ctr.ID), time.Now()))
	}
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

// openContainerUpdateForm opens the Update / resource-limit form. Text and
// int fields left empty leave the corresponding resource untouched.
func openContainerUpdateForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	restart := state.FormField{Key: fieldRestartPolicy, Label: i18n.T("container.update.form.restart"), Kind: state.FormSelect, Options: append([]string(nil), restartPolicyChoices...)}
	m.Form.Open(state.FormSpec{
		Kind:       state.FormContainerUpdate,
		Title:      i18n.T("container.update.form.title"),
		TargetID:   ctr.ID,
		TargetName: ctr.Name,
		Fields: []state.FormField{
			{Key: fieldMemory, Label: i18n.T("container.update.form.memory"), Kind: state.FormInt},
			{Key: fieldCPUs, Label: i18n.T("container.update.form.cpus"), Kind: state.FormInt},
			restart,
			{Key: fieldMaxRetries, Label: i18n.T("container.update.form.max_retries"), Kind: state.FormInt},
		},
	})
	m.Form.Loading = true
	m.Navigation.Mode = state.ModeContainerForm
	return m, containerUpdateConfigCmd(m.Connection.Engine, ctr.ID)
}

func containerUpdateConfigCmd(engine runtimeapi.Engine, containerID string) tea.Cmd {
	return func() tea.Msg {
		detail, err := engine.Containers().Inspect(context.Background(), containerID)
		return state.ContainerUpdateConfigLoaded{ContainerID: containerID, Detail: detail, Error: err}
	}
}

// HandleContainerUpdateConfigLoaded pre-fills untouched fields with the
// container's effective resource configuration.
func HandleContainerUpdateConfigLoaded(m *state.AppModel, msg state.ContainerUpdateConfigLoaded) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode != state.ModeContainerForm || m.Form.Kind != state.FormContainerUpdate || m.Form.TargetID != msg.ContainerID {
		return m, nil
	}
	m.Form.Loading = false
	if msg.Error != nil || msg.Detail == nil {
		err := msg.Error
		if err == nil {
			err = fmt.Errorf("container inspect returned no data")
		}
		ShowToastWarn(m, i18n.T("container.update.form.load_failed", err.Error()))
		return m, nil
	}
	resources := msg.Detail.Resources
	if resources.Memory > 0 {
		setUntouchedFormText(m.Form.Get(fieldMemory), formatResourceValue(float64(resources.Memory)/(1024*1024)))
	}
	if resources.NanoCPUs > 0 {
		setUntouchedFormText(m.Form.Get(fieldCPUs), formatResourceValue(float64(resources.NanoCPUs)/1e9))
	}
	if resources.RestartPolicy == "on-failure" && resources.MaximumRetryCount > 0 {
		setUntouchedFormText(m.Form.Get(fieldMaxRetries), strconv.Itoa(resources.MaximumRetryCount))
	}
	if restart := m.Form.Get(fieldRestartPolicy); restart != nil && !restart.Touched {
		policy := resources.RestartPolicy
		if policy == "" {
			policy = "no"
		}
		for i, option := range restart.Options {
			if option == policy {
				restart.Index = i
				break
			}
		}
	}
	return m, nil
}

func setUntouchedFormText(field *state.FormField, value string) {
	if field != nil && !field.Touched {
		field.Input.Set(value)
	}
}

func formatResourceValue(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// openContainerCommitForm opens the Commit form. Repository is required; tag,
// author, comment are optional; pause defaults to true.
func openContainerCommitForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	cwd := workingDir()
	repository, tag := defaultCommitImage(ctr.Image, ctr.Name)
	author := strings.TrimSpace(os.Getenv("USER"))
	if author == "" {
		author = strings.TrimSpace(os.Getenv("USERNAME"))
	}
	if author == "" {
		author = "dtui"
	}
	fields := []state.FormField{
		{Key: fieldRepository, Label: i18n.T("container.commit.form.repository"), Kind: state.FormText},
		{Key: fieldTag, Label: i18n.T("container.commit.form.tag"), Kind: state.FormText},
		{Key: fieldAuthor, Label: i18n.T("container.commit.form.author"), Kind: state.FormText},
		{Key: fieldComment, Label: i18n.T("container.commit.form.comment"), Kind: state.FormText},
		{Key: fieldPause, Label: i18n.T("container.commit.form.pause"), Kind: state.FormBool, Toggle: true},
		{Key: fieldExportTar, Label: i18n.T("container.commit.form.export_tar"), Kind: state.FormBool},
		{Key: fieldArchivePath, Label: i18n.T("container.commit.form.archive"), Kind: state.FormPath, PathSource: state.PathLocal, PathMode: state.PathSaveFile, DependsOn: fieldExportTar, DependsEq: true},
	}
	fields[0].Input.Set(repository)
	fields[1].Input.Set(tag)
	fields[2].Input.Set(author)
	fields[3].Input.Set(i18n.T("container.commit.form.comment_default", ctr.Name))
	fields[6].Input.Set(state.DefaultImageSaveName(cwd, repository+":"+tag, shortContainerID(ctr.ID), time.Now()))
	m.Form.Open(state.FormSpec{
		Kind:       state.FormContainerCommit,
		Title:      i18n.T("container.commit.form.title"),
		TargetID:   ctr.ID,
		TargetName: ctr.Name,
		CWD:        cwd,
		Fields:     fields,
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

func defaultCommitImage(image, containerName string) (repository, tag string) {
	image = strings.TrimSpace(image)
	if image == "" || strings.HasPrefix(image, "sha256:") {
		return state.SanitizeName(containerName, "container") + "-snapshot", "latest"
	}
	if at := strings.IndexByte(image, '@'); at >= 0 {
		image = image[:at]
	}
	repository, tag = image, "latest"
	if colon := strings.LastIndexByte(image, ':'); colon > strings.LastIndexByte(image, '/') {
		repository, tag = image[:colon], image[colon+1:]
	}
	if repository == "" {
		repository = state.SanitizeName(containerName, "container") + "-snapshot"
	}
	if tag == "" {
		tag = "latest"
	}
	return repository, tag
}

// handleContainerFormKey drives the container-action form overlay (BR-041 §3).
// Tab is reserved for completion; field and button navigation runs entirely on
// Up/Down/Left/Right. When a popup is open it overrides ordinary field keys.
func handleContainerFormKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Form.Popup.Open {
		return handleFormPopupKey(key, m)
	}

	// Button row: focus is on Cancel (default) or Confirm (BR-043 §3.3).
	if m.Form.FocusedButton() != "" {
		switch key {
		case keys.KeyTab:
			if m.Form.Kind == state.FormContainerUpdate {
				m.Form.MoveField(1)
			}
		case keys.KeyShiftTab:
			if m.Form.Kind == state.FormContainerUpdate {
				m.Form.MoveField(-1)
			}
		case keys.KeyUp, keys.KeyLeft:
			m.Form.MoveField(-1)
		case keys.KeyDown, keys.KeyRight:
			m.Form.MoveField(1)
		case keys.KeyEnter:
			if m.Form.FocusedButton() == "confirm" {
				return submitContainerForm(m)
			}
			clearContainerForm(m)
		case keys.KeyEsc:
			clearContainerForm(m)
		}
		return m, nil
	}

	f := m.Form.Field()
	prevFocus := m.Form.FieldFocus // captured for blur-time path absolutization (BR-043 §3.1)

	if handled := handleFormFieldEditKey(key, m, f); handled {
		return m, nil
	}

	switch key {
	case keys.KeyUp:
		m.Form.MoveField(-1)
	case keys.KeyDown:
		m.Form.MoveField(1)
	case keys.KeyTab:
		if m.Form.Kind == state.FormContainerUpdate {
			m.Form.MoveField(1)
		} else if f != nil && f.Kind == state.FormPath {
			return m, pathTabCycle(m, f)
		}
	case keys.KeyShiftTab:
		if m.Form.Kind == state.FormContainerUpdate {
			m.Form.MoveField(-1)
		}
	case keys.KeyCtrlSpace:
		if f != nil && f.Kind == state.FormPath {
			if f.PathSource == state.PathContainer {
				return m, requestContainerPathCompletion(m, f, true)
			} else {
				openPathPopup(m)
			}
		}
	case keys.KeyEsc:
		clearContainerForm(m)
	case keys.KeyEnter:
		if f == nil {
			return submitContainerForm(m)
		}
		switch f.Kind {
		case state.FormSelect, state.FormMultiSelect:
			m.Form.OpenPopup()
		case state.FormPath:
			if m.Form.Popup.Open {
				return confirmFormPopup(m)
			}
			if len(f.Suggestions) == 1 {
				applyPathCandidate(m)
				return m, nil
			}
			m.Form.MoveField(1)
		default:
			m.Form.MoveField(1)
		}
	default:
		// Single-rune input on Text/Int/Path fields.
		if f != nil && (f.Kind == state.FormText || f.Kind == state.FormInt || f.Kind == state.FormPath) {
			before := f.Input.Text
			editQueryInput(key, &f.Input)
			if f.Input.Text != before {
				handleFormFieldChanged(m, f)
			}
		}
	}

	// BR-043 §3.1: on field-exit, rewrite a FormPath field's text to its
	// absolute form so the display always shows the path that will be used
	// on submit. Cursor is clamped to the new rune length.
	if prevFocus != m.Form.FieldFocus && prevFocus >= 0 && prevFocus < len(m.Form.Fields) {
		absolutizePathFieldOnBlur(&m.Form.Fields[prevFocus], m.Form.CWD)
	}
	return m, nil
}

// handleFormFieldEditKey is the single field-type interaction policy shared
// by every Form. Navigation and popup activation remain in the form-level
// state machine; primitive editing and Bool activation are defined here.
func handleFormFieldEditKey(key string, m *state.AppModel, f *state.FormField) bool {
	if f == nil {
		return false
	}
	switch f.Kind {
	case state.FormBool:
		if keys.IsSpace(key) {
			f.ToggleBool()
			m.Form.RecomputeVisibility()
			return true
		}
		// Bool never treats Left/Right or printable input as an edit.
		return key == keys.KeyLeft || key == keys.KeyRight || len([]rune(key)) == 1
	case state.FormText, state.FormInt, state.FormPath:
		switch key {
		case keys.KeyLeft, keys.KeyRight, keys.KeyHome, keys.KeyEnd,
			keys.KeyBackspace, keys.KeyDelete, keys.KeySpace:
			_, changed := editQueryInput(key, &f.Input)
			if changed {
				handleFormFieldChanged(m, f)
			}
			return true
		}
	}
	return false
}

func handleFormFieldChanged(m *state.AppModel, f *state.FormField) {
	f.Touched = true
	if f.Kind == state.FormPath {
		f.PathLoading = false
		f.PathTabInput = ""
		if f.PathSource == state.PathContainer {
			f.Suggestions = nil
		} else {
			completeForField(m, f)
		}
	}
	if m.Form.Kind == state.FormContainerCopy && f.Key == fieldSourcePath {
		prefillDefaultDestination(m, m.Form.CWD, shortContainerID(m.Form.TargetID))
	}
}

// absolutizePathFieldOnBlur rewrites a FormPath field's text to its absolute
// form when focus moves away. Expands `~` and `$VAR`, then joins with cwd if
// still relative. Cursor is clamped to the new rune length. No-op if the
// field is empty, not a FormPath, or already absolute.
func absolutizePathFieldOnBlur(f *state.FormField, cwd string) {
	if f == nil || f.Kind != state.FormPath || f.Input.Text == "" {
		return
	}
	home, _ := os.UserHomeDir()
	expanded, err := state.ExpandPath(f.Input.Text, home)
	if err != nil {
		return
	}
	abs := state.Absolute(expanded, cwd)
	if abs == f.Input.Text {
		return
	}
	cursor := f.Input.Cursor
	f.Input.Set(abs)
	runes := []rune(abs)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	f.Input.Cursor = cursor
}

// pathTabCycle implements the path-field Tab state machine (BR-041 §3.3):
//   - popup already open: forward cycle the popup cursor.
//   - 1 candidate: apply directly.
//   - 0 candidates: keep input and focus, show a non-blocking toast.
//   - many candidates: extend the input to the common prefix, re-complete,
//     and open the popup if candidates remain.
func pathTabCycle(m *state.AppModel, f *state.FormField) tea.Cmd {
	if m.Form.Popup.Open {
		m.Form.PopupCursor(1)
		return nil
	}
	if f.PathSource == state.PathContainer {
		return requestContainerPathCompletion(m, f, false)
	}
	completeForField(m, f)
	applyPathSuggestions(m, f, false)
	return nil
}

func applyPathSuggestions(m *state.AppModel, f *state.FormField, openPopup bool) {
	switch len(f.Suggestions) {
	case 1:
		applyPathEntry(f, f.Suggestions[0])
		f.PathTabInput = ""
	case 0:
		f.PathTabInput = ""
		ShowToastWarn(m, i18n.T("form.path.no_matches"))
	default:
		prefix := commonPathPrefix(f.Suggestions)
		if !openPopup && len(prefix) > len(f.Input.Text) {
			f.Input.Set(prefix)
			f.Touched = true
			// The next Tab with the newly extended prefix lists candidates,
			// matching interactive shell completion.
			f.PathTabInput = prefix
			if f.PathSource == state.PathLocal {
				completeForField(m, f)
			}
			return
		}
		if openPopup || f.PathTabInput == f.Input.Text {
			m.Form.OpenPopup()
			f.PathTabInput = ""
		} else {
			f.PathTabInput = f.Input.Text
		}
	}
}

func requestContainerPathCompletion(m *state.AppModel, f *state.FormField, openPopup bool) tea.Cmd {
	if f.PathLoading {
		return nil
	}
	if m.Connection.Engine == nil || m.Connection.Engine.Exec() == nil || m.Form.TargetID == "" {
		ShowToastWarn(m, i18n.T("form.path.container_unavailable"))
		return nil
	}
	f.PathLoading = true
	f.Suggestions = nil
	return containerPathCompletionCmd(m.Connection.Engine.Exec(), m.Form.TargetID, f.Key, f.Input.Text, f.PathMode, openPopup)
}

func containerPathCompletionCmd(service runtimeapi.ExecService, containerID, fieldKey, input string, mode state.PathMode, openPopup bool) tea.Cmd {
	return func() tea.Msg {
		entries, err := listContainerPath(context.Background(), service, containerID, input, mode)
		return state.ContainerPathCompleted{ContainerID: containerID, FieldKey: fieldKey, Input: input, Entries: entries, OpenPopup: openPopup, Error: err}
	}
}

const containerPathListScript = `dir=$1
for entry in "$dir"/* "$dir"/.[!.]* "$dir"/..?*; do
  [ -e "$entry" ] || [ -L "$entry" ] || continue
  name=${entry##*/}
  if [ -d "$entry" ]; then kind=d; else kind=f; fi
  printf '%s\t%s\n' "$kind" "$name"
done`

func listContainerPath(ctx context.Context, service runtimeapi.ExecService, containerID, input string, mode state.PathMode) ([]state.PathEntry, error) {
	dir, prefix := splitContainerPath(input)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	session, err := service.Open(ctx, containerID, runtimeapi.ExecOptions{
		Command: []string{"/bin/sh", "-c", containerPathListScript, "dtui-path", dir},
		TTY:     true, AttachStdout: true,
	})
	if err != nil {
		return nil, err
	}
	defer session.Close()
	output, err := io.ReadAll(io.LimitReader(session, 1<<20))
	if err != nil {
		return nil, err
	}
	entries := parseContainerPathEntries(string(output), dir, prefix, mode)
	return entries, nil
}

func splitContainerPath(input string) (dir, prefix string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "/", ""
	}
	if strings.HasSuffix(input, "/") {
		return path.Clean(input), ""
	}
	return path.Dir(input), path.Base(input)
}

func parseContainerPathEntries(output, dir, prefix string, mode state.PathMode) []state.PathEntry {
	entries := make([]state.PathEntry, 0)
	for _, line := range strings.Split(strings.ReplaceAll(output, "\r", ""), "\n") {
		kind, name, ok := strings.Cut(line, "\t")
		if !ok || name == "" || !state.PrefixMatch(name, prefix) {
			continue
		}
		isDir := kind == "d"
		if mode == state.PathDirectory && !isDir {
			continue
		}
		entries = append(entries, state.PathEntry{Name: name, Path: path.Join(dir, name), IsDir: isDir})
		if len(entries) >= 4096 {
			break
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// HandleContainerPathCompleted applies a non-stale async completion result.
func HandleContainerPathCompleted(m *state.AppModel, msg state.ContainerPathCompleted) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode != state.ModeContainerForm || m.Form.TargetID != msg.ContainerID {
		return m, nil
	}
	f := m.Form.Get(msg.FieldKey)
	if f == nil || f.PathSource != state.PathContainer || f.Input.Text != msg.Input || !f.PathLoading {
		return m, nil
	}
	f.PathLoading = false
	if msg.Error != nil {
		f.Suggestions = nil
		ShowToastWarn(m, i18n.T("form.path.container_failed", msg.Error.Error()))
		return m, nil
	}
	f.Suggestions = msg.Entries
	applyPathSuggestions(m, f, msg.OpenPopup)
	return m, nil
}

// handleFormPopupKey routes keys while a selection or path popup is open:
// Esc closes, Up/Down/Home/End/PgUp/PgDn navigate, Enter commits, Space
// toggles multi / commits select, Tab and Shift+Tab cycle the cursor
// (BR-041 §3.4, §3.5, §3.3).
func handleFormPopupKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEsc:
		if m.Form.Popup.Kind == state.PopupPath {
			if f := m.Form.PopupField(); f != nil {
				f.PathLoading = false
			}
		}
		m.Form.ClosePopup()
	case keys.KeyUp:
		m.Form.PopupCursor(-1)
	case keys.KeyDown:
		m.Form.PopupCursor(1)
	case keys.KeyLeft:
		if m.Form.Popup.Kind == state.PopupPath {
			return navigatePathPopupParent(m)
		}
	case keys.KeyRight:
		if m.Form.Popup.Kind == state.PopupPath {
			return navigatePathPopupChild(m)
		}
	case keys.KeyHome:
		m.Form.PopupCursorHome()
	case keys.KeyEnd:
		m.Form.PopupCursorEnd()
	case keys.KeyPgUp:
		m.Form.PopupCursorPage(-1, popupVisibleRows)
	case keys.KeyPgDn:
		m.Form.PopupCursorPage(1, popupVisibleRows)
	case keys.KeyTab:
		if m.Form.Popup.Kind == state.PopupPath {
			m.Form.PopupCursor(1)
		}
	case keys.KeyShiftTab:
		if m.Form.Popup.Kind == state.PopupPath {
			m.Form.PopupCursor(-1)
		}
	case keys.KeyEnter:
		return confirmFormPopup(m)
	case keys.KeySpace:
		if f := m.Form.PopupField(); f != nil {
			if idx := m.Form.Popup.Cursor; idx >= 0 && idx < len(f.Options) {
				switch f.Kind {
				case state.FormMultiSelect:
					m.Form.TogglePopupMulti(f.Options[idx])
				case state.FormSelect:
					f.Index = idx
					f.Touched = true
					m.Form.ClosePopup()
				}
			}
		}
	}
	return m, nil
}

// popupVisibleRows matches the visible-row cap used by renderFormPopup so
// PgUp/PgDn advance by exactly one window.
const popupVisibleRows = state.FormPopupVisibleRows

// confirmFormPopup commits the popup selection onto its field and closes the
// popup. For FormPath the candidate is applied; for Select/MultiSelect the
// selection/focus updates before closing.
func confirmFormPopup(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	f := m.Form.PopupField()
	if f == nil {
		m.Form.ClosePopup()
		return m, nil
	}
	switch f.Kind {
	case state.FormSelect:
		if idx := m.Form.Popup.Cursor; idx >= 0 && idx < len(f.Options) {
			f.Index = idx
			f.Touched = true
		}
		m.Form.ClosePopup()
	case state.FormMultiSelect:
		m.Form.CommitPopupMulti()
	case state.FormPath:
		if idx := m.Form.Popup.Cursor; idx >= 0 && idx < len(f.Suggestions) {
			entry := f.Suggestions[idx]
			applyPathEntry(f, entry)
		}
		m.Form.ClosePopup()
	}
	return m, nil
}

func navigatePathPopupChild(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	f := m.Form.PopupField()
	if f == nil || m.Form.Popup.Cursor < 0 || m.Form.Popup.Cursor >= len(f.Suggestions) {
		return m, nil
	}
	entry := f.Suggestions[m.Form.Popup.Cursor]
	if !entry.IsDir {
		return m, nil
	}
	applyPathEntry(f, entry)
	return reloadPathPopup(m, f)
}

func navigatePathPopupParent(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	f := m.Form.PopupField()
	if f == nil || f.PathLoading {
		return m, nil
	}
	separator := string(os.PathSeparator)
	current := f.Input.Text
	if f.PathSource == state.PathContainer {
		separator = "/"
		current = strings.ReplaceAll(current, "\\", "/")
	}
	hadTrailingSeparator := strings.HasSuffix(current, separator)
	current = strings.TrimSuffix(current, separator)
	var parent string
	if f.PathSource == state.PathContainer {
		parent = path.Dir(current)
		if !hadTrailingSeparator {
			parent = path.Dir(parent)
		}
		if parent == "." {
			parent = "/"
		}
	} else {
		parent = filepath.Dir(current)
		if !hadTrailingSeparator {
			parent = filepath.Dir(parent)
		}
	}
	if parent != separator && !strings.HasSuffix(parent, separator) {
		parent += separator
	}
	f.Input.Set(parent)
	f.Touched = true
	return reloadPathPopup(m, f)
}

func reloadPathPopup(m *state.AppModel, f *state.FormField) (*state.AppModel, tea.Cmd) {
	f.PathTabInput = ""
	if f.PathSource == state.PathContainer {
		// Preserve the mounted popup while the next directory is loading so its
		// width and height remain stable throughout navigation.
		return m, requestContainerPathCompletion(m, f, true)
	}
	m.Form.ClosePopup()
	completeForField(m, f)
	if len(f.Suggestions) > 0 {
		m.Form.OpenPopup()
	}
	return m, nil
}

// submitContainerForm validates the active form and, on success, issues the
// matching runtime Cmd wrapped in an audit trace. Validation failures show a
// toast and leave the form open for correction.
func submitContainerForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	if m.Form.FocusedButton() != "confirm" {
		clearContainerForm(m)
		return m, nil
	}
	id, name := m.Form.TargetID, m.Form.TargetName
	switch m.Form.Kind {
	case state.FormContainerCopy:
		if m.Navigation.ActivePanel != state.PanelContainers || id == "" {
			return m, nil
		}
		src, dst := m.Form.Get(fieldSourcePath), m.Form.Get(fieldDestinationPath)
		if src == nil || dst == nil {
			clearContainerForm(m)
			return m, nil
		}
		if src.Text() == "" || dst.Text() == "" {
			ShowToastWarn(m, i18n.T("container.copy.form.required"))
			return m, nil
		}
		destination, err := normalizeSaveDestination(dst.Text(), m.Form.CWD)
		if err != nil {
			ShowToastWarn(m, i18n.T("form.path.invalid", err.Error()))
			return m, nil
		}
		dst.Input.Set(destination)
		if localDestExists(m, destination) {
			return openOverwriteConfirm(m, "container-copy", "resource.container.copy", "Copy file from container "+name)
		}
		trace := beginAudit(m, "resource.container.copy", containerTarget(m, id), "Copy file from container "+name)
		clearContainerForm(m)
		return m, withAdvancedAudit(containerCopyCmd(m.Connection.Engine, id, src.Text(), dst.Text()), trace)

	case state.FormContainerExport:
		if m.Navigation.ActivePanel != state.PanelContainers || id == "" {
			return m, nil
		}
		dst := m.Form.Get(fieldDestinationPath)
		if dst == nil {
			clearContainerForm(m)
			return m, nil
		}
		if dst.Text() == "" {
			ShowToastWarn(m, i18n.T("container.export.form.required"))
			return m, nil
		}
		destination, err := normalizeSaveDestination(dst.Text(), m.Form.CWD)
		if err != nil {
			ShowToastWarn(m, i18n.T("form.path.invalid", err.Error()))
			return m, nil
		}
		dst.Input.Set(destination)
		if localDestExists(m, destination) {
			return openOverwriteConfirm(m, "container-export", "resource.container.export", "Export container "+name)
		}
		trace := beginAudit(m, "resource.container.export", containerTarget(m, id), "Export container "+name)
		clearContainerForm(m)
		return m, withAdvancedAudit(containerExportCmd(m.Connection.Engine, id, dst.Text()), trace)

	case state.FormContainerUpdate:
		if m.Navigation.ActivePanel != state.PanelContainers || id == "" {
			return m, nil
		}
		opts := runtimeapi.ContainerUpdateOptions{}
		changed := false
		if mem := m.Form.Get(fieldMemory); mem != nil && mem.Touched && mem.Text() != "" {
			mb, err := strconv.ParseFloat(mem.Text(), 64)
			if err != nil || mb < 0 || math.IsNaN(mb) || math.IsInf(mb, 0) || mb > float64(math.MaxInt64)/(1024*1024) {
				ShowToastWarn(m, i18n.T("container.update.form.invalid"))
				return m, nil
			}
			bytes := int64(mb * 1024 * 1024)
			opts.Memory = &bytes
			changed = true
		}
		if cpu := m.Form.Get(fieldCPUs); cpu != nil && cpu.Touched && cpu.Text() != "" {
			cores, err := strconv.ParseFloat(cpu.Text(), 64)
			if err != nil || cores < 0 || math.IsNaN(cores) || math.IsInf(cores, 0) || cores > float64(math.MaxInt64)/1e9 {
				ShowToastWarn(m, i18n.T("container.update.form.invalid"))
				return m, nil
			}
			nano := int64(cores * 1e9)
			opts.NanoCPUs = &nano
			changed = true
		}
		if restart := m.Form.Get(fieldRestartPolicy); restart != nil && restart.Touched && restart.Option() != restartPolicyUnchanged {
			policy := restart.Option()
			opts.RestartPolicy = &policy
			changed = true
			if policy == "on-failure" {
				if retries := m.Form.Get(fieldMaxRetries); retries != nil && retries.Touched && retries.Text() != "" {
					val, err := strconv.Atoi(retries.Text())
					if err != nil || val < 0 {
						ShowToastWarn(m, i18n.T("container.update.form.invalid"))
						return m, nil
					}
					opts.RestartMaxRetries = &val
				}
			}
		}
		if !changed {
			ShowToastWarn(m, i18n.T("container.update.form.no_changes"))
			return m, nil
		}
		trace := beginAudit(m, "resource.container.update", containerTarget(m, id), "Update container "+name)
		clearContainerForm(m)
		return m, withAdvancedAudit(containerUpdateCmd(m.Connection.Engine, id, opts), trace)

	case state.FormContainerCommit:
		if m.Navigation.ActivePanel != state.PanelContainers || id == "" {
			return m, nil
		}
		opts, archivePath, err := containerCommitFormRequest(m)
		if err != nil {
			ShowToastWarn(m, i18n.T("container.commit.form.required"))
			return m, nil
		}
		if archivePath != "" && localDestExists(m, archivePath) {
			return openOverwriteConfirm(m, "container-commit-export", "resource.container.commit", "Commit container "+name+" and export image")
		}
		trace := beginAudit(m, "resource.container.commit", containerTarget(m, id), "Commit container "+name)
		return executeContainerCommitForm(m, opts, archivePath, trace)

	case state.FormImageSave:
		path := m.Form.Get(fieldImagePath)
		if m.Navigation.ActivePanel != state.PanelImages || id == "" || path == nil || path.Text() == "" {
			ShowToastWarn(m, i18n.T("image.transfer.required"))
			return m, nil
		}
		destination, err := normalizeSaveDestination(path.Text(), m.Form.CWD)
		if err != nil {
			ShowToastWarn(m, i18n.T("form.path.invalid", err.Error()))
			return m, nil
		}
		path.Input.Set(destination)
		if localDestExists(m, destination) {
			return openImageSaveOverwriteConfirm(m)
		}
		request := runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferSave, Source: id, Path: destination}
		clearContainerForm(m)
		return beginImageTransfer(m, request)

	case state.FormImageLoad:
		path := m.Form.Get(fieldImagePath)
		if m.Navigation.ActivePanel != state.PanelImages || path == nil || path.Text() == "" {
			ShowToastWarn(m, i18n.T("image.transfer.required"))
			return m, nil
		}
		source, err := normalizeLoadSource(path.Text(), m.Form.CWD)
		if err != nil {
			ShowToastWarn(m, i18n.T("form.path.invalid", err.Error()))
			return m, nil
		}
		request := runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferLoad, Path: source}
		clearContainerForm(m)
		return beginImageTransfer(m, request)
	}
	clearContainerForm(m)
	return m, nil
}

func containerCommitFormRequest(m *state.AppModel) (runtimeapi.ContainerCommitOptions, string, error) {
	repo := m.Form.Get(fieldRepository)
	if repo == nil || repo.Text() == "" {
		return runtimeapi.ContainerCommitOptions{}, "", fmt.Errorf("repository is required")
	}
	tag := m.Form.Get(fieldTag).Text()
	if tag == "" {
		tag = "latest"
	}
	opts := runtimeapi.ContainerCommitOptions{
		Repository: repo.Text(), Tag: tag, Author: m.Form.Get(fieldAuthor).Text(),
		Comment: m.Form.Get(fieldComment).Text(), Pause: m.Form.Get(fieldPause).Toggle,
	}
	archivePath := ""
	if export := m.Form.Get(fieldExportTar); export != nil && export.Toggle {
		archive := m.Form.Get(fieldArchivePath)
		if archive == nil || archive.Text() == "" {
			return runtimeapi.ContainerCommitOptions{}, "", fmt.Errorf("archive path is required")
		}
		destination, err := normalizeSaveDestination(archive.Text(), m.Form.CWD)
		if err != nil {
			return runtimeapi.ContainerCommitOptions{}, "", err
		}
		archive.Input.Set(destination)
		archivePath = destination
	}
	return opts, archivePath, nil
}

func executeContainerCommitForm(m *state.AppModel, opts runtimeapi.ContainerCommitOptions, archivePath string, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	id := m.Form.TargetID
	clearContainerForm(m)
	return m, withAdvancedAudit(containerCommitCmd(m.Connection.Engine, id, opts, archivePath), trace)
}

// clearContainerForm resets the mode and closes the active form.
func clearContainerForm(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Form.Close()
}

// workingDir returns the dtui process working directory, falling back to ".".
func workingDir() string {
	wd, err := os.Getwd()
	if err != nil || wd == "" {
		return "."
	}
	return wd
}

// shortContainerID abbreviates a container ID to its first 12 characters.
func shortContainerID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// prefillDefaultDestination sets the default local tar name for the Copy form
// when the destination has not been edited. Re-generating on a changed source
// is deferred to completeForField; here only the initial value is applied.
func prefillDefaultDestination(m *state.AppModel, cwd, shortID string) {
	dst := m.Form.Get(fieldDestinationPath)
	src := m.Form.Get(fieldSourcePath)
	if dst == nil || dst.Touched {
		return
	}
	base := ""
	if src != nil {
		base = state.PathBase(src.Text())
	}
	if base == "" {
		base = "container"
	}
	dst.Input.Set(state.DefaultCopyName(cwd, m.Form.TargetName, base, shortID, time.Now()))
}

// completeForField refreshes path completion candidates for a focused local
// path field, using the form CWD. On failure candidates are left empty so the
// user can keep typing manually (BR-041 §6.3).
func completeForField(m *state.AppModel, f *state.FormField) {
	if f == nil || f.Kind != state.FormPath || f.PathSource != state.PathLocal {
		return
	}
	provider := state.LocalPathProvider{CWD: m.Form.CWD}
	cands, err := provider.Complete(state.PathCompletionRequest{
		Path: f.Input.Text,
		Mode: f.PathMode,
	})
	if err != nil {
		f.Suggestions = nil
		f.Error = err.Error()
		return
	}
	f.Error = ""
	f.Suggestions = cands
}

// openPathPopup opens the candidate popup for the focused local path field.
func openPathPopup(m *state.AppModel) {
	f := m.Form.Field()
	if f == nil || f.Kind != state.FormPath || f.PathSource != state.PathLocal {
		return
	}
	completeForField(m, f)
	if len(f.Suggestions) > 0 {
		m.Form.OpenPopup()
	}
}

func commonPathPrefix(entries []state.PathEntry) string {
	if len(entries) == 0 {
		return ""
	}
	prefix := []rune(entries[0].Path)
	for _, entry := range entries[1:] {
		candidate := []rune(entry.Path)
		n := min(len(prefix), len(candidate))
		i := 0
		for i < n && prefix[i] == candidate[i] {
			i++
		}
		prefix = prefix[:i]
		if len(prefix) == 0 {
			break
		}
	}
	return string(prefix)
}

// applyPathCandidate fills the focused path field from its suggestions when
// there is exactly one candidate (Tab completion). Multi-candidate selection
// is delegated to the popup via confirmFormPopup.
func applyPathCandidate(m *state.AppModel) {
	f := m.Form.Field()
	if f == nil || f.Kind != state.FormPath || f.PathSource != state.PathLocal {
		return
	}
	if len(f.Suggestions) == 1 {
		applyPathEntry(f, f.Suggestions[0])
	}
}

func applyPathEntry(f *state.FormField, entry state.PathEntry) {
	value := entry.Path
	separator := string(os.PathSeparator)
	if f.PathSource == state.PathContainer {
		separator = "/"
	}
	if entry.IsDir && !strings.HasSuffix(value, separator) {
		value += separator
	}
	f.Input.Set(value)
	f.Touched = true
}

// localDestExists reports whether the local destination path already exists.
// Used to gate overwrite confirmation for Copy/Export (BR-041 §10).
func localDestExists(m *state.AppModel, dst string) bool {
	if dst == "" {
		return false
	}
	_, err := os.Stat(state.Absolute(dst, m.Form.CWD))
	return err == nil
}

func normalizeSaveDestination(input, cwd string) (string, error) {
	home, _ := os.UserHomeDir()
	expanded, err := state.ExpandPath(input, home)
	if err != nil {
		return "", err
	}
	destination := state.Absolute(expanded, cwd)
	parent := state.Absolute(state.JoinPath(destination, ".."), cwd)
	info, err := os.Stat(parent)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("parent is not a directory: %s", parent)
	}
	return destination, nil
}

func normalizeLoadSource(input, cwd string) (string, error) {
	home, _ := os.UserHomeDir()
	expanded, err := state.ExpandPath(input, home)
	if err != nil {
		return "", err
	}
	source := state.Absolute(expanded, cwd)
	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", source)
	}
	return source, nil
}

// openOverwriteConfirm raises the BR-041 §10 confirmation when the Copy/Export
// destination already exists: default focus on Cancel, with a separate Force
// option. Cancelling returns to the still-open form.
func openOverwriteConfirm(m *state.AppModel, action, auditAction, message string) (*state.AppModel, tea.Cmd) {
	trace := beginAudit(m, auditAction, containerTarget(m, m.Form.TargetID), message)
	m.Confirm.Open(action, m.Form.TargetID, i18n.T("form.overwrite.confirm"), trace)
	m.Confirm.Options = []state.ChoiceOption{
		{ID: "cancel", Label: i18n.T("key.cancel")},
		{ID: "force", Label: i18n.T("form.overwrite.force"), Description: i18n.T("form.overwrite.force_desc")},
	}
	m.Confirm.ReturnMode = state.ModeContainerForm
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func openImageSaveOverwriteConfirm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Confirm.Open("image-save", m.Form.TargetID, i18n.T("form.overwrite.confirm"), audit.Trace{})
	m.Confirm.Options = []state.ChoiceOption{
		{ID: "cancel", Label: i18n.T("key.cancel")},
		{ID: "force", Label: i18n.T("form.overwrite.force"), Description: i18n.T("form.overwrite.force_desc")},
	}
	m.Confirm.ReturnMode = state.ModeContainerForm
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}
