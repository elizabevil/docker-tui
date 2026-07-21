package keyboard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func ImagePullCmd(client *docker.Client, ref string) tea.Cmd {
	return func() tea.Msg {
		err := client.PullImage(ref)
		return state.ImageActioned{Action: state.ActionPulled, Ref: ref, Success: err == nil, Error: err}
	}
}

func ImagePullCmdWithAudit(client *docker.Client, ref string, trace audit.Trace) tea.Cmd {
	return withImageAudit(ImagePullCmd(client, ref), trace)
}

func imagePruneCmd(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		reclaimed, err := client.PruneImages()
		return state.ImageActioned{Action: state.ActionPruned, Ref: fmt.Sprintf("%d bytes reclaimed", reclaimed), Success: err == nil, Error: err}
	}
}

func imageRemoveCmd(client *docker.Client, id string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveImage(id, force)
		return state.ImageActioned{Action: state.ActionRemoved, Ref: id, Success: err == nil, Error: err}
	}
}

func doImagePull(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	m.Mode = state.ModeImagePull
	m.FilterText = ""
	m.FilterCursor = 0
	m.InfoMessage = "Type image name (e.g. nginx:latest) and press Enter to pull"
	return m, nil
}

func handleImagePullInput(key string, m *state.AppModel) *state.AppModel {
	switch key {
	case keys.KeyEnter:
		ref := strings.TrimSpace(m.FilterText)
		if ref == "" {
			ShowToastWarn(m, "Image reference is required")
			return m
		}
		m.PendingImagePull = ref
		m.PendingImagePullAudit = beginAudit(m, "resource.image.pull", audit.ImageTarget{ID: ref, Name: ref}, "Pulling image "+ref)
		m.FilterText = ""
		m.FilterCursor = 0
		m.Mode = state.ModeNormal
	case keys.KeyEsc:
		m.FilterText = ""
		m.FilterCursor = 0
		m.Mode = state.ModeNormal
	default:
		editTextInput(key, &m.FilterText, &m.FilterCursor)
	}
	return m
}

func doImagePrune(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	trace := beginAudit(m, "resource.image.prune", audit.ImageTarget{ID: "unused", Name: "unused images"}, "Pruning unused images")
	return m, withImageAudit(imagePruneCmd(m.Docker), trace)
}

func doImageRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || m.ActivePanel != state.PanelImages {
		return m, nil
	}
	img := m.Images.Selected()
	if img == nil {
		return m, nil
	}
	tag := ""
	if len(img.RepoTags) > 0 {
		tag = img.RepoTags[0]
	}
	confirmAction(m, "image-remove", img.ID, fmt.Sprintf("Remove image %s?", tag))
	m.ConfirmAudit = beginAudit(m, "resource.image.delete", imageTarget(m, img.ID), "Remove image "+tag)
	return m, nil
}

func doImageDetail(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	img := m.Images.Selected()
	if img == nil {
		return m, nil
	}
	m.ImageDetailID = img.ID
	m.ImageDetailContent = ""
	m.ImageDetailData = docker.NewImageDetailData(*img)
	titleID := img.ID
	if len(titleID) > 12 {
		titleID = titleID[:12]
	}
	m.DetailTitle = "Image Detail: " + titleID
	m.DetailHint = ""
	m.DetailOffset = 0
	m.PrevPanel = m.ActivePanel
	m.Mode = state.ModeDetail
	return m, inspectImageCmd(m.Docker, *img)
}

func inspectImageCmd(client *docker.Client, image docker.ImageSummary) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.InspectImageDetail(image)
		return state.ImageDetailLoaded{
			ImageID: image.ID,
			Detail:  detail,
			Error:   err,
		}
	}
}

func doImageSort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ActivePanel != state.PanelImages {
		return m, nil
	}
	if m.Images.SortBy >= state.ImageSortByCreated {
		m.Images.SortBy = state.ImageSortByRepo
		m.Images.SortAsc = !m.Images.SortAsc
	} else {
		m.Images.SortBy++
	}
	m.Images.Cursor = 0
	m.Images.ViewOffset = 0
	m.InfoMessage = fmt.Sprintf("Sort by %s (%v)", imageSortColLabel(state.ImageSortColumn(m.Images.SortBy)), m.Images.SortAsc)
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
	img := m.Images.Selected()
	if img == nil {
		return m, nil
	}
	if !hasUsingContainers(m.Containers, img.ID, img.RepoTags) {
		ShowToastWarn(m, "no containers using "+shortID(img.ID))
		return m, nil
	}
	ToImageContainers(m, img.ID)
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

func doImageExport(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	img := m.Images.Selected()
	if img == nil || m.Docker == nil {
		return m, nil
	}
	tag := tagName(img)
	home, _ := os.UserHomeDir()
	exportDir := filepath.Join(home, "dtui-exports")
	os.MkdirAll(exportDir, 0755)
	path := filepath.Join(exportDir, tag+".tar")

	engine := "docker"
	if m.Docker.RuntimeType == "podman" {
		engine = "podman"
	}
	ref := img.ID[:20]
	if len(img.RepoTags) > 0 {
		ref = img.RepoTags[0]
	}

	cmd := fmt.Sprintf("%s save -o %s %s", engine, path, ref)
	m.Mode = state.ModeExport
	m.DialogTitle = "Export Image"
	m.DialogBody = fmt.Sprintf("Image: %s\nOutput: %s", ref, path)
	m.DialogPreview = cmd
	m.DialogAction = "Export command copied to preview"
	return m, nil
}

func doImageDebug(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	img := m.Images.Selected()
	if img == nil || m.Docker == nil {
		return m, nil
	}
	engine := "docker"
	if m.Docker.RuntimeType == "podman" {
		engine = "podman"
	}
	ref := img.ID[:20]
	if len(img.RepoTags) > 0 {
		ref = img.RepoTags[0]
	}

	cmd := fmt.Sprintf("%s run --rm -it --name debug-%s %s sh",
		engine, tagName(img), ref)
	m.Mode = state.ModeDebug
	m.DialogTitle = "Debug Run"
	m.DialogBody = fmt.Sprintf("Image: %s\nContainer: debug-%s", ref, tagName(img))
	m.DialogPreview = cmd
	m.DialogAction = "Debug command copied to preview"
	return m, nil
}

func fullImageRef(img *docker.ImageSummary) string {
	if len(img.RepoTags) > 0 && img.RepoTags[0] != "<none>:<none>" {
		return img.RepoTags[0]
	}
	if len(img.ID) >= 12 {
		return img.ID[:12]
	}
	return img.ID
}

func doImageCopyRef(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ActivePanel != state.PanelImages {
		return m, nil
	}
	img := m.Images.Selected()
	if img == nil {
		return m, nil
	}
	ref := fullImageRef(img)
	ShowToastNow(m, "✓ Copied: "+ref)
	return m, func() tea.Msg {
		var cmd *exec.Cmd
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("pbcopy"); err == nil {
			cmd = exec.Command("pbcopy")
		} else {
			// Fallback: write to temp file
			os.WriteFile("/tmp/dtui-clipboard.txt", []byte(ref), 0644)
			return nil
		}
		cmd.Stdin = strings.NewReader(ref)
		cmd.Run()
		return nil
	}
}

func tagName(img *docker.ImageSummary) string {
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
