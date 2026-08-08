package keyboard

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
	"github.com/elizabevil/docker-tui/internal/utils"

	tea "charm.land/bubbletea/v2"
)

const (
	fieldSourcePath               = "source"
	fieldDestinationPath          = "destination"
	fieldMemory                   = "memory"
	fieldCPUs                     = "cpus"
	fieldRestartPolicy            = "restart"
	fieldMaxRetries               = "maxRetries"
	fieldRepository               = "repository"
	fieldTag                      = "tag"
	fieldAuthor                   = "author"
	fieldComment                  = "comment"
	fieldPause                    = "pause"
	fieldExportTar                = "exportTar"
	fieldArchivePath              = "archivePath"
	fieldImagePath                = "imagePath"
	fieldImageRemoveForce         = "force"
	fieldImageRemovePruneChildren = "pruneChildren"
	fieldImageRemovePlatforms     = "platforms"
	fieldContainerRemoveForce     = "force"
	fieldContainerRemoveVolumes   = "removeVolumes"
	fieldContainerRemoveLinks     = "removeLinks"
	fieldVolumeRemoveForce        = "force"
	fieldComposeScaleReplicas     = "replicas"
	fieldComposeScaleNoDeps       = "noDeps"
	// FormNetworkRemove has no editable fields (network remove does
	// not support --force), so no fieldNetworkRemoveForce constant.
)

const restartPolicyUnchanged = "unchanged"

type restartChoice struct {
	Key   string
	Label string
}

var restartPolicyChoices = []restartChoice{
	{Key: "unchanged", Label: i18n.T("container.update.form.restart.unchanged")},
	{Key: "no", Label: i18n.T("container.update.form.restart.no")},
	{Key: "always", Label: i18n.T("container.update.form.restart.always")},
	{Key: "unless-stopped", Label: i18n.T("container.update.form.restart.unless_stopped")},
	{Key: "on-failure", Label: i18n.T("container.update.form.restart.on_failure")},
}

func buildRestartField() state.FormField {
	options := make([]string, len(restartPolicyChoices))
	display := make([]string, len(restartPolicyChoices))
	for i, c := range restartPolicyChoices {
		options[i] = c.Key
		display[i] = c.Label
	}
	return dialog.NewSelectField(dialog.SelectFieldConfig{
		Key:            fieldRestartPolicy,
		Label:          i18n.T("container.update.form.restart"),
		Options:        options,
		DisplayOptions: display,
	})
}

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
			dialog.NewPathField(dialog.PathFieldConfig{
				Key:        fieldSourcePath,
				Label:      i18n.T("container.copy.form.source"),
				PathSource: state.PathContainer,
				PathMode:   state.PathAny,
				Required:   true,
			}),
			dialog.NewPathField(dialog.PathFieldConfig{
				Key:        fieldDestinationPath,
				Label:      i18n.T("container.copy.form.destination"),
				PathSource: state.PathLocal,
				PathMode:   state.PathSaveFile,
				Required:   true,
			}),
		},
	})
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
	shortID := utils.ShortID(image.ID)
	path := state.DefaultImageSaveName(cwd, ref, shortID, time.Now())
	field := dialog.NewPathField(dialog.PathFieldConfig{
		Key:        fieldImagePath,
		Label:      i18n.T("image.save.form.path"),
		PathSource: state.PathLocal,
		PathMode:   state.PathSaveFile,
		Required:   true,
		Text:       path,
	})
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
		Fields: []state.FormField{dialog.NewPathField(dialog.PathFieldConfig{
			Key:        fieldImagePath,
			Label:      i18n.T("image.load.form.path"),
			PathSource: state.PathLocal,
			PathMode:   state.PathFile,
			Required:   true,
		})},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

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
		Fields: []state.FormField{dialog.NewPathField(dialog.PathFieldConfig{
			Key:        fieldDestinationPath,
			Label:      i18n.T("container.export.form.destination"),
			PathSource: state.PathLocal,
			PathMode:   state.PathSaveFile,
			Required:   true,
		})},
	})
	if dst := m.Form.Get(fieldDestinationPath); dst != nil && dst.Text() == "" {
		if pf, ok := dst.(*dialog.PathField); ok {
			pf.SetText(state.DefaultExportName(cwd, ctr.Name, shortContainerID(ctr.ID), time.Now()))
		}
	}
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

func openContainerUpdateForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	restart := buildRestartField()
	m.Form.Open(state.FormSpec{
		Kind:       state.FormContainerUpdate,
		Title:      i18n.T("container.update.form.title"),
		TargetID:   ctr.ID,
		TargetName: ctr.Name,
		Fields: []state.FormField{
			dialog.NewIntField(dialog.IntFieldConfig{
				Key:        fieldMemory,
				Label:      i18n.T("container.update.form.memory"),
				HelperText: "MB",
				Unit:       "MB",
				Min:        state.FormMin(0),
				Max:        state.FormMax(1e15),
			}),
			dialog.NewIntField(dialog.IntFieldConfig{
				Key:        fieldCPUs,
				Label:      i18n.T("container.update.form.cpus"),
				HelperText: "cores",
				Unit:       "cores",
				Min:        state.FormMin(0),
				Max:        state.FormMax(1e15),
			}),
			restart,
			dialog.NewIntField(dialog.IntFieldConfig{
				Key:   fieldMaxRetries,
				Label: i18n.T("container.update.form.max_retries"),
			}),
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
	if restart := m.Form.Get(fieldRestartPolicy); restart != nil && !restart.Touched() {
		policy := resources.RestartPolicy
		if policy == "" {
			policy = "no"
		}
		options := restart.Options()
		for i, option := range options {
			if option == policy {
				restart.SetIndex(i)
				break
			}
		}
	}
	return m, nil
}

func setUntouchedFormText(field state.FormField, value string) {
	if field == nil || field.Touched() {
		return
	}
	field.SetText(value)
}

func formatResourceValue(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func buildImageRemoveForceField() state.FormField {
	return dialog.NewBoolField(dialog.BoolFieldConfig{
		Key:        fieldImageRemoveForce,
		Label:      i18n.T("image.remove.form.force"),
		HelperText: i18n.T("image.remove.form.force_desc"),
	})
}

func buildImageRemovePruneChildrenField() state.FormField {
	return dialog.NewBoolField(dialog.BoolFieldConfig{
		Key:        fieldImageRemovePruneChildren,
		Label:      i18n.T("image.remove.form.prune_children"),
		HelperText: i18n.T("image.remove.form.prune_children_desc"),
		DependsOn:  fieldImageRemoveForce,
		DependsEq:  true,
	})
}

func buildImageRemovePlatformsField() state.FormField {
	return dialog.NewTextField(dialog.TextFieldConfig{
		Key:        fieldImageRemovePlatforms,
		Label:      i18n.T("image.remove.form.platforms"),
		HelperText: i18n.T("image.remove.form.platforms_desc"),
	})
}

func buildContainerRemoveForceField() state.FormField {
	return dialog.NewBoolField(dialog.BoolFieldConfig{
		Key:        fieldContainerRemoveForce,
		Label:      i18n.T("container.remove.form.force"),
		HelperText: i18n.T("container.remove.form.force_desc"),
	})
}

func buildContainerRemoveVolumesField() state.FormField {
	return dialog.NewBoolField(dialog.BoolFieldConfig{
		Key:        fieldContainerRemoveVolumes,
		Label:      i18n.T("container.remove.form.volumes"),
		HelperText: i18n.T("container.remove.form.volumes_desc"),
		DependsOn:  fieldContainerRemoveForce,
		DependsEq:  true,
	})
}

func buildContainerRemoveLinksField() state.FormField {
	return dialog.NewBoolField(dialog.BoolFieldConfig{
		Key:        fieldContainerRemoveLinks,
		Label:      i18n.T("container.remove.form.links"),
		HelperText: i18n.T("container.remove.form.links_desc"),
		DependsOn:  fieldContainerRemoveForce,
		DependsEq:  true,
	})
}

func buildVolumeRemoveForceField() state.FormField {
	return dialog.NewBoolField(dialog.BoolFieldConfig{
		Key:        fieldVolumeRemoveForce,
		Label:      i18n.T("volume.remove.form.force"),
		HelperText: i18n.T("volume.remove.form.force_desc"),
	})
}

func openImageRemoveForm(m *state.AppModel, img *runtimeapi.ImageSummary) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelImages || img == nil {
		return m, nil
	}
	m.Form.Open(state.FormSpec{
		Kind:       state.FormImageRemove,
		Title:      i18n.T("image.remove.form.title"),
		TargetID:   img.ID,
		TargetName: tagName(img),
		CWD:        workingDir(),
		Dangerous:  true,
		Fields: []state.FormField{
			buildImageRemoveForceField(),
			buildImageRemovePruneChildrenField(),
			buildImageRemovePlatformsField(),
		},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

// openComposeScaleForm opens the R08-09 scale form for the currently
// selected service. The form carries one Int spinner (target replicas)
// and one Bool toggle (--no-deps).
//
// R08-09 × R08-14: on Docker, a service whose containers belong to a
// CoLocated group (network_mode/pid/ipc service: references) is pinned
// to 1 replica. The scale action is blocked with a toast instead of
// opening the form. Podman keeps its own group semantics (1 project =
// 1 pod), so scaling a group member is allowed there.
func openComposeScaleForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	project := currentComposeProject(m)
	service := selectedDetailComposeService(m, project)
	if m.Connection.Engine == nil || project == "" || service == "" {
		ShowToastWarn(m, "✕ compose scale: pick a service in the compose panel")
		return m, nil
	}
	if m.Connection.Engine.Identity().Type == runtimeapi.Docker {
		for _, c := range state.ComposeProjectContainers(m, project) {
			if c.ComposeService == service && c.CoLocatedGroupID != "" {
				ShowToastWarn(m, "✕ Docker compose 限制 network_mode: service: 的 service 不能 scale>1")
				return m, nil
			}
		}
	}
	current := 0
	for _, c := range state.ComposeProjectContainers(m, project) {
		if c.ComposeService == service {
			current++
		}
	}
	m.Form.Open(state.FormSpec{
		Kind:       state.FormComposeScale,
		Title:      "Scale " + project + "/" + service,
		TargetID:   project,
		TargetName: service,
		Fields: []state.FormField{
			dialog.NewIntField(dialog.IntFieldConfig{
				Key:   fieldComposeScaleReplicas,
				Label: "Target replicas (current: " + strconv.Itoa(current) + ")",
				Min:   state.FormMin(0),
				Max:   state.FormMax(100),
				Text:  strconv.Itoa(current),
			}),
			dialog.NewBoolField(dialog.BoolFieldConfig{
				Key:   fieldComposeScaleNoDeps,
				Label: "--no-deps",
			}),
		},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

func openContainerRemoveForm(m *state.AppModel, ctr *runtimeapi.ContainerSummary) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	m.Form.Open(state.FormSpec{
		Kind:       state.FormContainerRemove,
		Title:      i18n.T("container.remove.form.title"),
		TargetID:   ctr.ID,
		TargetName: ctr.Name,
		CWD:        workingDir(),
		Dangerous:  true,
		Fields: []state.FormField{
			buildContainerRemoveForceField(),
			buildContainerRemoveVolumesField(),
			buildContainerRemoveLinksField(),
		},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

func openVolumeRemoveForm(m *state.AppModel, vol *runtimeapi.Volume) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelVolumes || vol == nil {
		return m, nil
	}
	m.Form.Open(state.FormSpec{
		Kind:       state.FormVolumeRemove,
		Title:      i18n.T("volume.remove.form.title"),
		TargetID:   vol.Name,
		TargetName: vol.Name,
		CWD:        workingDir(),
		Dangerous:  true,
		Fields:     []state.FormField{buildVolumeRemoveForceField()},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

// openNetworkRemoveForm is the network-scope twin of openVolumeRemoveForm.
// Unlike volume remove (which has a Force field because docker /
// podman support `--force`), networks cannot be force-removed — the
// backend NetworkService.Remove takes only id. So the form here opens
// with no editable fields; the user just sees the target name and
// presses Confirm or Cancel. R06-08 state consolidation may fold both
// into a generic openResourceRemoveForm(scope, summary).
func openNetworkRemoveForm(m *state.AppModel, net *runtimeapi.Network) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelNetworks || net == nil {
		return m, nil
	}
	m.Form.Open(state.FormSpec{
		Kind:       state.FormNetworkRemove,
		Title:      i18n.T("network.remove.form.title"),
		TargetID:   net.ID,
		TargetName: net.Name,
		CWD:        workingDir(),
		Dangerous:  true,
		Fields:     []state.FormField{},
	})
	m.Navigation.Mode = state.ModeContainerForm
	return m, nil
}

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
		dialog.NewTextField(dialog.TextFieldConfig{Key: fieldRepository, Label: i18n.T("container.commit.form.repository"), Text: repository}),
		dialog.NewTextField(dialog.TextFieldConfig{Key: fieldTag, Label: i18n.T("container.commit.form.tag"), Text: tag}),
		dialog.NewTextField(dialog.TextFieldConfig{Key: fieldAuthor, Label: i18n.T("container.commit.form.author"), Text: author}),
		dialog.NewTextField(dialog.TextFieldConfig{Key: fieldComment, Label: i18n.T("container.commit.form.comment"), Text: i18n.T("container.commit.form.comment_default", ctr.Name)}),
		dialog.NewBoolField(dialog.BoolFieldConfig{Key: fieldPause, Label: i18n.T("container.commit.form.pause"), Toggle: true}),
		dialog.NewBoolField(dialog.BoolFieldConfig{Key: fieldExportTar, Label: i18n.T("container.commit.form.export_tar")}),
		dialog.NewPathField(dialog.PathFieldConfig{
			Key:        fieldArchivePath,
			Label:      i18n.T("container.commit.form.archive"),
			PathSource: state.PathLocal,
			PathMode:   state.PathSaveFile,
			Text:       state.DefaultImageSaveName(cwd, repository+":"+tag, shortContainerID(ctr.ID), time.Now()),
		}),
	}
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

// handleContainerFormKey drives the container-action form overlay.
func handleContainerFormKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Form.Popup.Open {
		return handleFormPopupKey(key, m)
	}

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
			if m.Form.FocusedButton() == keys.ShowOptionConfirm {
				return submitContainerForm(m)
			}
			clearContainerForm(m)
		case keys.KeyEsc:
			clearContainerForm(m)
		}
		return m, nil
	}

	f := m.Form.Field()
	prevFocus := m.Form.FieldFocus

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
		} else if f != nil && f.Kind() == state.FormPath {
			return m, pathTabCycle(m, f)
		}
	case keys.KeyShiftTab:
		if m.Form.Kind == state.FormContainerUpdate {
			m.Form.MoveField(-1)
		}
	case keys.KeyCtrlSpace:
		if f != nil && f.Kind() == state.FormPath {
			if pf, _ := f.(*dialog.PathField); pf != nil && pf.PathSource() == state.PathContainer {
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
		switch f.Kind() {
		case state.FormSelect, state.FormMultiSelect:
			m.Form.OpenPopup()
		case state.FormPath:
			if m.Form.Popup.Open {
				return confirmFormPopup(m)
			}
			if pf, _ := f.(*dialog.PathField); pf != nil && len(pf.Suggestions()) == 1 {
				applyPathCandidate(m)
				return m, nil
			}
			m.Form.MoveField(1)
		default:
			m.Form.MoveField(1)
		}
	default:
		if f != nil && (f.Kind() == state.FormText || f.Kind() == state.FormInt || f.Kind() == state.FormPath) {
			before := f.Text()
			input := state.QueryInputState{Text: before, Cursor: f.Cursor()}
			editQueryInput(key, &input)
			if input.Text != before {
				f.SetText(input.Text)
				f.SetCursor(input.Cursor)
				handleFormFieldChanged(m, f)
			}
		}
	}

	if prevFocus != m.Form.FieldFocus && prevFocus >= 0 && prevFocus < len(m.Form.Fields) {
		absolutizePathFieldOnBlur(m.Form.Fields[prevFocus], m.Form.CWD)
	}
	return m, nil
}

// handleFormFieldEditKey routes per-kind edit keys through the dialog-side
// FormField interface. The keyboard layer keeps navigation, popup activation,
// and post-edit side effects.
func handleFormFieldEditKey(key string, m *state.AppModel, f state.FormField) bool {
	if f == nil {
		return false
	}
	df, ok := f.(dialog.FormField)
	if !ok {
		return false
	}
	handled, _ := df.HandleKey(key)
	if handled {
		handleFormFieldChanged(m, f)
	}
	return handled
}

func handleFormFieldChanged(m *state.AppModel, f state.FormField) {
	f.SetTouched(true)
	if f.Kind() == state.FormPath {
		if pf, ok := f.(*dialog.PathField); ok {
			pf.SetPathLoading(false)
			pf.SetPathTabInput("")
			if pf.PathSource() == state.PathContainer {
				pf.SetSuggestions(nil)
			} else {
				completeForField(m, pf)
			}
		}
	}
	if f.Kind() == state.FormBool {
		m.Form.RecomputeVisibility()
	}
	if m.Form.Kind == state.FormContainerCopy && f.Key() == fieldSourcePath {
		prefillDefaultDestination(m, m.Form.CWD, shortContainerID(m.Form.TargetID))
	}
}

// absolutizePathFieldOnBlur rewrites a local FormPath field's text to its
// absolute form when focus moves away.
func absolutizePathFieldOnBlur(f state.FormField, cwd string) {
	if f == nil || f.Kind() != state.FormPath {
		return
	}
	pf, ok := f.(*dialog.PathField)
	if !ok || pf.PathSource() != state.PathLocal || pf.TextRaw() == "" {
		return
	}
	home, _ := os.UserHomeDir()
	expanded, err := state.ExpandPath(pf.TextRaw(), home)
	if err != nil {
		return
	}
	abs := state.Absolute(expanded, cwd)
	if abs == pf.TextRaw() {
		return
	}
	cursor := pf.Cursor()
	pf.SetText(abs)
	runes := []rune(abs)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	pf.SetCursor(cursor)
}

// pathTabCycle implements the path-field Tab state machine.
func pathTabCycle(m *state.AppModel, f state.FormField) tea.Cmd {
	if m.Form.Popup.Open {
		m.Form.PopupCursor(1)
		return nil
	}
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return nil
	}
	if pf.PathSource() == state.PathContainer {
		return requestContainerPathCompletion(m, f, false)
	}
	completeForField(m, pf)
	applyPathSuggestions(m, f, false)
	return nil
}

func applyPathSuggestions(m *state.AppModel, f state.FormField, openPopup bool) {
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return
	}
	switch len(pf.Suggestions()) {
	case 1:
		applyPathEntry(pf, pf.Suggestions()[0])
		pf.SetPathTabInput("")
	case 0:
		pf.SetPathTabInput("")
		ShowToastWarn(m, i18n.T("form.path.no_matches"))
	default:
		prefix := commonPathPrefix(pf.Suggestions())
		if !openPopup && len(prefix) > len(pf.TextRaw()) {
			pf.SetText(prefix)
			pf.SetTouched(true)
			pf.SetPathTabInput(prefix)
			if pf.PathSource() == state.PathLocal {
				completeForField(m, pf)
			}
			return
		}
		if openPopup || pf.PathTabInput() == pf.TextRaw() {
			m.Form.OpenPopup()
			pf.SetPathTabInput("")
		} else {
			pf.SetPathTabInput(pf.TextRaw())
		}
	}
}

func requestContainerPathCompletion(m *state.AppModel, f state.FormField, openPopup bool) tea.Cmd {
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return nil
	}
	if pf.PathLoading() {
		return nil
	}
	if m.Connection.Engine == nil || m.Connection.Engine.Exec() == nil || m.Form.TargetID == "" {
		ShowToastWarn(m, i18n.T("form.path.container_unavailable"))
		return nil
	}
	pf.SetPathLoading(true)
	pf.SetSuggestions(nil)
	return containerPathCompletionCmd(m.Connection.Engine.Exec(), m.Form.TargetID, f.Key(), pf.TextRaw(), pf.PathMode(), pf.ShowHidden(), openPopup)
}

func containerPathCompletionCmd(service runtimeapi.ExecService, containerID, fieldKey, input string, mode state.PathMode, showHidden, openPopup bool) tea.Cmd {
	return func() tea.Msg {
		entries, err := listContainerPath(context.Background(), service, containerID, input, mode, showHidden)
		return state.ContainerPathCompleted{ContainerID: containerID, FieldKey: fieldKey, Input: input, Entries: entries, OpenPopup: openPopup, Error: err}
	}
}

const containerPathListScript = `dir=$1
showhidden=$2
hidden_glob=
if [ "$showhidden" = "1" ]; then hidden_glob='.[!.]* ..?*'; fi
for entry in "$dir"/* $hidden_glob; do
  [ -e "$entry" ] || [ -L "$entry" ] || continue
  name=${entry##*/}
  if [ -L "$entry" ]; then
    kind=l; target=$(readlink "$entry" 2>/dev/null)
  elif [ -d "$entry" ]; then
    kind=d; target=
  else
    kind=f; target=
  fi
  stat -c "$kind|$name|%A|%U|%G|%s|%Y|$target" "$entry" 2>/dev/null
done`

func listContainerPath(ctx context.Context, service runtimeapi.ExecService, containerID, input string, mode state.PathMode, showHidden bool) ([]state.PathEntry, error) {
	dir, prefix := splitContainerPath(input)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	hiddenArg := "0"
	if showHidden {
		hiddenArg = "1"
	}
	session, err := service.Open(ctx, containerID, runtimeapi.ExecOptions{
		Command: []string{config.DefaultShell, "-c", containerPathListScript, "dtui-path", dir, hiddenArg},
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
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 8)
		if len(parts) < 7 || parts[1] == "" {
			continue
		}
		kind, name := parts[0], parts[1]
		if !state.PrefixMatch(name, prefix) {
			continue
		}
		entry := state.PathEntry{
			Name:  name,
			Path:  path.Join(dir, name),
			IsDir: kind == "d",
			Mode:  parts[2],
			Owner: parts[3],
			Group: parts[4],
		}
		entry.Size, _ = strconv.ParseInt(parts[5], 10, 64)
		if ts, err := strconv.ParseInt(parts[6], 10, 64); err == nil {
			entry.Mtime = time.Unix(ts, 0)
		}
		switch kind {
		case "d":
			entry.Type = state.PathEntryDir
		case "l":
			entry.Type = state.PathEntryLink
			if len(parts) == 8 {
				entry.LinkTarget = parts[7]
			}
		default:
			entry.Type = state.PathEntryFile
		}
		if mode == state.PathDirectory && !entry.IsDir {
			continue
		}
		entries = append(entries, entry)
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

func HandleContainerPathCompleted(m *state.AppModel, msg state.ContainerPathCompleted) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode != state.ModeContainerForm || m.Form.TargetID != msg.ContainerID {
		return m, nil
	}
	f := m.Form.Get(msg.FieldKey)
	if f == nil {
		return m, nil
	}
	pf, ok := f.(*dialog.PathField)
	if !ok || pf.PathSource() != state.PathContainer || pf.TextRaw() != msg.Input || !pf.PathLoading() {
		return m, nil
	}
	pf.SetPathLoading(false)
	if msg.Error != nil {
		pf.SetSuggestions(nil)
		pf.SetPathError(msg.Error.Error())
		ShowToastWarn(m, i18n.T("form.path.container_failed", msg.Error.Error()))
		return m, nil
	}
	pf.SetPathError("")
	pf.SetSuggestions(msg.Entries)
	applyPathSuggestions(m, f, msg.OpenPopup)
	return m, nil
}

// handleFormPopupKey routes keys while a selection or path popup is open.
func handleFormPopupKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEsc:
		if m.Form.Popup.Kind == state.PopupPath {
			if f := m.Form.PopupField(); f != nil {
				if pf, ok := f.(*dialog.PathField); ok {
					pf.SetPathLoading(false)
				}
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
			if df, ok := f.(dialog.FormField); ok {
				if idx := m.Form.Popup.Cursor; idx >= 0 && idx < len(df.Options()) {
					switch f.Kind() {
					case state.FormMultiSelect:
						m.Form.TogglePopupMulti(df.Options()[idx])
					case state.FormSelect:
						df.SetIndex(idx)
						df.SetTouched(true)
						m.Form.ClosePopup()
					}
				}
			}
		}
	default:
		if m.Form.Popup.Kind == state.PopupPath {
			if f := m.Form.PopupField(); f != nil && f.Kind() == state.FormPath {
				if pf, ok := f.(*dialog.PathField); ok {
					if isPathEditKey(key) {
						before := pf.TextRaw()
						input := state.QueryInputState{Text: before, Cursor: pf.Cursor()}
						_, changed := editQueryInput(key, &input)
						if changed {
							pf.SetText(input.Text)
							pf.SetCursor(input.Cursor)
							pf.SetPathError("")
							return m, triggerPathBrowseCompletion(m, f)
						}
					}
				}
			}
		}
	}
	return m, nil
}

func isPathEditKey(key string) bool {
	switch key {
	case keys.KeyBackspace, keys.KeyDelete:
		return true
	}
	if len([]rune(key)) == 1 {
		return true
	}
	return false
}

func triggerPathBrowseCompletion(m *state.AppModel, f state.FormField) tea.Cmd {
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return nil
	}
	if pf.PathSource() == state.PathContainer {
		return requestContainerPathCompletion(m, f, true)
	}
	completeForField(m, pf)
	applyPathSuggestions(m, f, true)
	m.Form.Popup.Cursor = 0
	return nil
}

const popupVisibleRows = state.FormPopupVisibleRows

func confirmFormPopup(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	f := m.Form.PopupField()
	if f == nil {
		m.Form.ClosePopup()
		return m, nil
	}
	switch f.Kind() {
	case state.FormSelect:
		df, ok := f.(dialog.FormField)
		if !ok {
			m.Form.ClosePopup()
			return m, nil
		}
		options := df.Options()
		if idx := m.Form.Popup.Cursor; idx >= 0 && idx < len(options) {
			df.SetIndex(idx)
			df.SetTouched(true)
		}
		m.Form.ClosePopup()
	case state.FormMultiSelect:
		m.Form.CommitPopupMulti()
	case state.FormPath:
		pf, ok := f.(*dialog.PathField)
		if !ok {
			m.Form.ClosePopup()
			return m, nil
		}
		idx := m.Form.Popup.Cursor
		if idx < 0 || idx >= len(pf.Suggestions()) {
			m.Form.ClosePopup()
			return m, nil
		}
		entry := pf.Suggestions()[idx]
		applyPathEntry(pf, entry)
		if entry.IsDir {
			text := pf.TextRaw()
			if !strings.HasSuffix(text, "/") {
				text += "/"
			}
			pf.SetText(text)
			pf.SetCursor(len([]rune(text)))
			pf.SetSuggestions(nil)
			m.Form.Popup.Cursor = 0
			m.Form.ClosePopup()
			return m, requestContainerPathCompletion(m, f, true)
		}
		m.Form.ClosePopup()
	}
	return m, nil
}

func navigatePathPopupChild(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	f := m.Form.PopupField()
	if f == nil {
		return m, nil
	}
	pf, ok := f.(*dialog.PathField)
	if !ok || m.Form.Popup.Cursor < 0 || m.Form.Popup.Cursor >= len(pf.Suggestions()) {
		return m, nil
	}
	entry := pf.Suggestions()[m.Form.Popup.Cursor]
	if !entry.IsDir {
		return m, nil
	}
	applyPathEntry(pf, entry)
	return reloadPathPopup(m, f)
}

func navigatePathPopupParent(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	f := m.Form.PopupField()
	if f == nil {
		return m, nil
	}
	pf, ok := f.(*dialog.PathField)
	if !ok || pf.PathLoading() {
		return m, nil
	}
	separator := string(os.PathSeparator)
	current := pf.TextRaw()
	if pf.PathSource() == state.PathContainer {
		separator = "/"
		current = strings.ReplaceAll(current, "\\", "/")
	}
	hadTrailingSeparator := strings.HasSuffix(current, separator)
	current = strings.TrimSuffix(current, separator)
	var parent string
	if pf.PathSource() == state.PathContainer {
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
	pf.SetText(parent)
	pf.SetTouched(true)
	return reloadPathPopup(m, f)
}

func reloadPathPopup(m *state.AppModel, f state.FormField) (*state.AppModel, tea.Cmd) {
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return m, nil
	}
	pf.SetPathTabInput("")
	if pf.PathSource() == state.PathContainer {
		return m, requestContainerPathCompletion(m, f, true)
	}
	m.Form.ClosePopup()
	completeForField(m, pf)
	if len(pf.Suggestions()) > 0 {
		m.Form.OpenPopup()
	}
	return m, nil
}

func validateFloatField(m *state.AppModel, f state.FormField) (float64, bool) {
	value, err := strconv.ParseFloat(f.Text(), 64)
	if err != nil || value < 0 {
		ShowToastWarn(m, i18n.T("container.update.form.invalid"))
		return 0, false
	}
	if intf, ok := f.(*dialog.IntField); ok {
		if min, max := intf.Min(), intf.Max(); min != nil && value < *min {
			ShowToastWarn(m, i18n.T("container.update.form.invalid_range"))
			return 0, false
		} else if max != nil && value > *max {
			ShowToastWarn(m, i18n.T("container.update.form.invalid_range"))
			return 0, false
		}
	}
	return value, true
}

func submitContainerForm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	if m.Form.FocusedButton() != keys.ShowOptionConfirm {
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
		dst.SetText(destination)
		if localDestExists(m, destination) {
			return openOverwriteConfirm(m, keys.ShowContainerCopy, "resource.container.copy", "Copy file from container "+name)
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
		dst.SetText(destination)
		if localDestExists(m, destination) {
			return openOverwriteConfirm(m, keys.ShowContainerExport, "resource.container.export", "Export container "+name)
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
		if mem := m.Form.Get(fieldMemory); mem != nil && mem.Touched() && mem.Text() != "" {
			mb, ok := validateFloatField(m, mem)
			if !ok {
				return m, nil
			}
			bytes := int64(mb * 1024 * 1024)
			opts.Memory = &bytes
			changed = true
		}
		if cpu := m.Form.Get(fieldCPUs); cpu != nil && cpu.Touched() && cpu.Text() != "" {
			cores, ok := validateFloatField(m, cpu)
			if !ok {
				return m, nil
			}
			nano := int64(cores * 1e9)
			opts.NanoCPUs = &nano
			changed = true
		}
		if restart := m.Form.Get(fieldRestartPolicy); restart != nil && restart.Touched() {
			policy := restart.Value().(string)
			if policy != restartPolicyUnchanged {
				opts.RestartPolicy = &policy
				changed = true
				if policy == "on-failure" {
					if retries := m.Form.Get(fieldMaxRetries); retries != nil && retries.Touched() && retries.Text() != "" {
						val, err := strconv.Atoi(retries.Text())
						if err != nil || val < 0 {
							ShowToastWarn(m, i18n.T("container.update.form.invalid"))
							return m, nil
						}
						opts.RestartMaxRetries = &val
					}
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
			return openOverwriteConfirm(m, keys.ShowContainerCommitExport, "resource.container.commit", "Commit container "+name+" and export image")
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
		path.SetText(destination)
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

	case state.FormImageRemove:
		if m.Navigation.ActivePanel != state.PanelImages || id == "" {
			clearContainerForm(m)
			return m, nil
		}
		force := formBoolValue(m.Form.Get(fieldImageRemoveForce))
		pruneChildren := formBoolValue(m.Form.Get(fieldImageRemovePruneChildren))
		platforms := parsePlatformInput(m.Form.Get(fieldImageRemovePlatforms))
		trace := beginAudit(m, "resource.image.delete", imageTarget(m, id), "Remove image "+name)
		clearContainerForm(m)
		return m, withImageAudit(imageRemoveCmd(m.Connection.Engine, id, force, pruneChildren, platforms), trace)

	case state.FormContainerRemove:
		if m.Navigation.ActivePanel != state.PanelContainers || id == "" {
			clearContainerForm(m)
			return m, nil
		}
		force := formBoolValue(m.Form.Get(fieldContainerRemoveForce))
		removeVolumes := formBoolValue(m.Form.Get(fieldContainerRemoveVolumes))
		removeLinks := formBoolValue(m.Form.Get(fieldContainerRemoveLinks))
		trace := beginAudit(m, "resource.container.delete", containerTarget(m, id), "Remove container "+name)
		clearContainerForm(m)
		return m, withContainerAudit(containerRemoveCmd(m.Connection.Engine, id, runtimeapi.LifecycleOptions{
			Force:         force,
			RemoveVolumes: removeVolumes,
			RemoveLinks:   removeLinks,
		}), trace)

	case state.FormVolumeRemove:
		if m.Navigation.ActivePanel != state.PanelVolumes || id == "" {
			clearContainerForm(m)
			return m, nil
		}
		force := formBoolValue(m.Form.Get(fieldVolumeRemoveForce))
		trace := beginAudit(m, "resource.volume.delete", audit.VolumeTarget{Name: name, Meta: audit.VolumeMeta{Driver: volumeDriver(m, name)}}, "Remove volume "+name)
		clearContainerForm(m)
		return m, withGenericAudit(volumeRemoveCmd(m.Connection.Engine, id, runtimeapi.LifecycleOptions{Force: force}), trace)

	case state.FormNetworkRemove:
		if m.Navigation.ActivePanel != state.PanelNetworks || id == "" {
			clearContainerForm(m)
			return m, nil
		}
		trace := beginAudit(m, "resource.network.delete", audit.NetworkTarget{ID: id, Name: name}, "Remove network "+name)
		clearContainerForm(m)
		return m, withGenericAudit(networkRemoveCmd(m.Connection.Engine, id), trace)
	case state.FormComposeScale:
		project := id
		service := name
		replicas := 1
		if f := m.Form.Get(fieldComposeScaleReplicas); f != nil {
			if v, ok := parseScaleReplicas(f.Text()); ok {
				replicas = v
			} else {
				ShowToastWarn(m, i18n.T("compose.scale.form.invalid"))
				clearContainerForm(m)
				return m, nil
			}
		}
		noDeps := false
		if f := m.Form.Get(fieldComposeScaleNoDeps); f != nil {
			noDeps = f.Toggle()
		}
		clearContainerForm(m)
		return executeComposeScale(m, project, service, replicas, noDeps)
	}
	clearContainerForm(m)
	return m, nil
}

func formBoolValue(f state.FormField) bool {
	if f == nil {
		return false
	}
	return f.Toggle()
}

func parsePlatformInput(f state.FormField) []string {
	if f == nil {
		return nil
	}
	raw := strings.TrimSpace(f.Text())
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func volumeDriver(m *state.AppModel, name string) string {
	if m == nil || m.Resources.Volumes == nil {
		return ""
	}
	for _, v := range m.Resources.Volumes.Items {
		if v.Name == name {
			return v.Driver
		}
	}
	return ""
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
	pause := false
	if p := m.Form.Get(fieldPause); p != nil {
		pause = p.Toggle()
	}
	opts := runtimeapi.ContainerCommitOptions{
		Repository: repo.Text(),
		Tag:        tag,
		Author:     m.Form.Get(fieldAuthor).Text(),
		Comment:    m.Form.Get(fieldComment).Text(),
		Pause:      pause,
	}
	archivePath := ""
	if export := m.Form.Get(fieldExportTar); export != nil && export.Toggle() {
		archive := m.Form.Get(fieldArchivePath)
		if archive == nil || archive.Text() == "" {
			return runtimeapi.ContainerCommitOptions{}, "", fmt.Errorf("archive path is required")
		}
		destination, err := normalizeSaveDestination(archive.Text(), m.Form.CWD)
		if err != nil {
			return runtimeapi.ContainerCommitOptions{}, "", err
		}
		archive.SetText(destination)
		archivePath = destination
	}
	return opts, archivePath, nil
}

func executeContainerCommitForm(m *state.AppModel, opts runtimeapi.ContainerCommitOptions, archivePath string, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	id := m.Form.TargetID
	clearContainerForm(m)
	return m, withAdvancedAudit(containerCommitCmd(m.Connection.Engine, id, opts, archivePath), trace)
}

func clearContainerForm(m *state.AppModel) {
	m.Navigation.Mode = state.ModeNormal
	m.Form.Close()
}

func workingDir() string {
	wd, err := os.Getwd()
	if err != nil || wd == "" {
		return "."
	}
	return wd
}

func shortContainerID(id string) string {
	return utils.ShortID(id)
}

func prefillDefaultDestination(m *state.AppModel, cwd, shortID string) {
	dst := m.Form.Get(fieldDestinationPath)
	src := m.Form.Get(fieldSourcePath)
	if dst == nil || dst.Touched() {
		return
	}
	base := ""
	if src != nil {
		base = state.PathBase(src.Text())
	}
	if base == "" {
		base = "container"
	}
	dst.SetText(state.DefaultCopyName(cwd, m.Form.TargetName, base, shortID, time.Now()))
}

func completeForField(m *state.AppModel, pf *dialog.PathField) {
	if pf == nil || pf.Kind() != state.FormPath || pf.PathSource() != state.PathLocal {
		return
	}
	provider := state.LocalPathProvider{CWD: m.Form.CWD}
	cands, err := provider.Complete(state.PathCompletionRequest{
		Path: pf.TextRaw(),
		Mode: pf.PathMode(),
	})
	if err != nil {
		pf.SetSuggestions(nil)
		pf.SetError(err.Error())
		return
	}
	pf.SetError("")
	pf.SetSuggestions(cands)
}

func openPathPopup(m *state.AppModel) {
	f := m.Form.Field()
	if f == nil || f.Kind() != state.FormPath {
		return
	}
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return
	}
	completeForField(m, pf)
	if len(pf.Suggestions()) > 0 {
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

func applyPathCandidate(m *state.AppModel) {
	f := m.Form.Field()
	if f == nil || f.Kind() != state.FormPath {
		return
	}
	pf, ok := f.(*dialog.PathField)
	if !ok {
		return
	}
	if len(pf.Suggestions()) == 1 {
		applyPathEntry(pf, pf.Suggestions()[0])
	}
}

func applyPathEntry(pf *dialog.PathField, entry state.PathEntry) {
	value := entry.Path
	separator := string(os.PathSeparator)
	if pf.PathSource() == state.PathContainer {
		separator = "/"
	}
	if entry.IsDir && !strings.HasSuffix(value, separator) {
		value += separator
	}
	pf.SetText(value)
	pf.SetTouched(true)
}

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

func openOverwriteConfirm(m *state.AppModel, action, auditAction, message string) (*state.AppModel, tea.Cmd) {
	trace := beginAudit(m, auditAction, containerTarget(m, m.Form.TargetID), message)
	m.Confirm.Open(action, m.Form.TargetID, i18n.T("form.overwrite.confirm"), trace)
	m.Confirm.Options = []state.ChoiceOption{
		{ID: keys.ShowOptionCancel, Label: i18n.T("key.cancel")},
		{ID: keys.ShowOptionForce, Label: i18n.T("form.overwrite.force"), Description: i18n.T("form.overwrite.force_desc")},
	}
	m.Confirm.ReturnMode = state.ModeContainerForm
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func openImageSaveOverwriteConfirm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Confirm.Open(keys.ShowImageSave, m.Form.TargetID, i18n.T("form.overwrite.confirm"), audit.Trace{})
	m.Confirm.Options = []state.ChoiceOption{
		{ID: keys.ShowOptionCancel, Label: i18n.T("key.cancel")},
		{ID: keys.ShowOptionForce, Label: i18n.T("form.overwrite.force"), Description: i18n.T("form.overwrite.force_desc")},
	}
	m.Confirm.ReturnMode = state.ModeContainerForm
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}
