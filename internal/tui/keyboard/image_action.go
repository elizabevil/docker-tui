package keyboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func ImagePullCmd(client runtimeapi.Engine, ref string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: ref}, runtimeapi.ActionPull, runtimeapi.ActionOptions{})
		return state.ImageActioned{Action: state.ActionPulled, Ref: ref, Success: err == nil, Error: err}
	}
}

func ImagePullCmdWithAudit(client runtimeapi.Engine, ref string, trace audit.Trace) tea.Cmd {
	return withImageAudit(ImagePullCmd(client, ref), trace)
}

func imagePruneCmd(client runtimeapi.Engine) tea.Cmd {
	return func() tea.Msg {
		result, err := client.Actions().Execute(context.Background(), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage}, runtimeapi.ActionPrune, runtimeapi.ActionOptions{})
		return state.ImageActioned{Action: state.ActionPruned, Ref: fmt.Sprintf("%d bytes reclaimed", result.SpaceReclaimed), Success: err == nil, Error: err}
	}
}

func imageRemoveCmd(client runtimeapi.Engine, id string, force bool) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: id}, runtimeapi.ActionRemove,
			runtimeapi.ActionOptions{Lifecycle: runtimeapi.LifecycleOptions{Force: force}})
		return state.ImageActioned{Action: state.ActionRemoved, Ref: id, Success: err == nil, Error: err}
	}
}

func doImagePull(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	m.Dialog.Open(state.DialogSpec{Kind: state.DialogImagePull})
	m.Navigation.Mode = m.Dialog.Kind.Mode()
	m.Feedback.InfoMessage = "Type image name (e.g. nginx:latest) and press Enter to pull"
	return m, nil
}

func handleImagePullInput(key string, m *state.AppModel) *state.AppModel {
	switch key {
	case keys.KeyEnter:
		ref := strings.TrimSpace(m.Dialog.Input.Text)
		if ref == "" {
			ShowToastWarn(m, "Image reference is required")
			return m
		}
		trace := beginAudit(m, "resource.image.pull", audit.ImageTarget{ID: ref, Name: ref}, "Pulling image "+ref)
		m.Selection.QueueImagePull(ref, trace)
		clearDialogState(m)
	case keys.KeyEsc:
		clearDialogState(m)
	default:
		editQueryInput(key, &m.Dialog.Input)
	}
	return m
}

func doImagePrune(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	trace := beginAudit(m, "resource.image.prune", audit.ImageTarget{ID: "unused", Name: "unused images"}, "Pruning unused images")
	return m, withImageAudit(imagePruneCmd(m.Connection.Engine), trace)
}

func doImageRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelImages {
		return m, nil
	}
	img := m.Resources.Images.Selected()
	if img == nil {
		return m, nil
	}
	tag := ""
	if len(img.RepoTags) > 0 {
		tag = img.RepoTags[0]
	}
	confirmAction(m, keys.ShowImageRemove, img.ID, fmt.Sprintf("Remove image %s?", tag))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource.image.delete", imageTarget(m, img.ID), "Remove image "+tag)
	return m, nil
}

func doImageDetail(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	img := m.Resources.Images.Selected()
	if img == nil {
		return m, nil
	}
	titleID := img.ID
	if len(titleID) > 12 {
		titleID = titleID[:12]
	}
	m.Detail.OpenImage(img.ID, "Image Detail: "+titleID, docker.NewImageDetailData(*img))
	m.Navigation.PrevPanel = m.Navigation.ActivePanel
	m.Navigation.Mode = state.ModeDetail
	return m, inspectImageCmd(m.Connection.Engine, *img)
}

func inspectImageCmd(client runtimeapi.Engine, image runtimeapi.ImageSummary) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.Images().Inspect(context.Background(), image)
		return state.ImageDetailLoaded{
			ImageID: image.ID,
			Detail:  detail,
			Error:   err,
		}
	}
}

func doImageSort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.ActivePanel != state.PanelImages {
		return m, nil
	}
	if m.Resources.Images.SortBy >= state.ImageSortByCreated {
		m.Resources.Images.SortBy = state.ImageSortByRepo
		m.Resources.Images.SortAsc = !m.Resources.Images.SortAsc
	} else {
		m.Resources.Images.SortBy++
	}
	m.Resources.Images.Cursor = 0
	m.Resources.Images.ViewOffset = 0
	m.Feedback.InfoMessage = fmt.Sprintf("Sort by %s (%v)", imageSortColLabel(state.ImageSortColumn(m.Resources.Images.SortBy)), m.Resources.Images.SortAsc)
	return m, nil
}

func imageSortColLabel(col state.ImageSortColumn) string {
	switch col {
	case state.ImageSortByRepo:
		return "name"
	case state.ImageSortByID:
		return "id"
	case state.ImageSortBySize:
		return "size"
	case state.ImageSortByCreated:
		return "created"
	}
	return ""
}

func doImageExpand(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	img := m.Resources.Images.Selected()
	if img == nil {
		return m, nil
	}
	if !hasUsingContainers(m.Resources.Containers, img.ID, img.RepoTags) {
		ShowToastWarn(m, "no containers using "+shortID(img.ID))
		return m, nil
	}
	ToImageContainers(m, img.ID, fullImageRef(img))
	return m, nil
}

// hasUsingContainers checks if any container is using the given image.
func hasUsingContainers(cm *state.ContainerListModel, imgID string, tags []string) bool {
	if cm == nil {
		return false
	}
	short := shortID(imgID)
	for _, c := range cm.Items {
		if strings.Contains(c.Image, short) {
			return true
		}
		for _, tag := range tags {
			if strings.Contains(c.Image, tag) {
				return true
			}
		}
	}
	return false
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func doImageCollapse(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	BackFromImageContainers(m)
	return m, nil
}

func doImageDebug(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	img := m.Resources.Images.Selected()
	if img == nil || m.Connection.Engine == nil {
		return m, nil
	}
	engine := "docker"
	if m.Connection.Engine.Identity().Type == runtimeapi.Podman {
		engine = "podman"
	}
	ref := img.ID[:20]
	if len(img.RepoTags) > 0 {
		ref = img.RepoTags[0]
	}

	cmd := fmt.Sprintf("%s run --rm -it --name debug-%s %s sh",
		engine, tagName(img), ref)
	m.Dialog.Open(state.DialogSpec{
		Kind:    state.DialogImageDebug,
		Title:   "Debug Run",
		Body:    fmt.Sprintf("Image: %s\nContainer: debug-%s", ref, tagName(img)),
		Preview: cmd,
		Action:  "Debug command copied to preview",
	})
	m.Navigation.Mode = m.Dialog.Kind.Mode()
	return m, nil
}

func fullImageRef(img *runtimeapi.ImageSummary) string {
	if len(img.RepoTags) > 0 && img.RepoTags[0] != "<none>:<none>" {
		return img.RepoTags[0]
	}
	if len(img.ID) >= 12 {
		return img.ID[:12]
	}
	return img.ID
}

func doImageCopyRef(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.ActivePanel != state.PanelImages {
		return m, nil
	}
	img := m.Resources.Images.Selected()
	if img == nil {
		return m, nil
	}
	ref := fullImageRef(img)
	ShowToastNow(m, "✓ Copied: "+ref)
	return m, clipboardCmd(ref)
}

func tagName(img *runtimeapi.ImageSummary) string {
	if len(img.RepoTags) > 0 {
		name := strings.ReplaceAll(img.RepoTags[0], ":", "-")
		name = strings.ReplaceAll(name, "/", "_")
		if len(name) > 20 {
			name = name[:20]
		}
		return name
	}
	return "debug"
}
