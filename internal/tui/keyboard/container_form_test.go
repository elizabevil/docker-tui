package keyboard

import (
	"context"
	"io"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
)

type pathExecService struct {
	output string
	err    error
}

func (s pathExecService) Open(_ context.Context, _ string, _ runtimeapi.ExecOptions) (runtimeapi.ExecSession, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &pathExecSession{Reader: strings.NewReader(s.output)}, nil
}

type pathExecSession struct{ io.Reader }

func (s *pathExecSession) ID() string                                            { return "path-exec" }
func (s *pathExecSession) Write(p []byte) (int, error)                           { return len(p), nil }
func (s *pathExecSession) Close() error                                          { return nil }
func (s *pathExecSession) Resize(context.Context, runtimeapi.TerminalSize) error { return nil }

type pathCompletionEngine struct {
	*advancedEngine
	exec runtimeapi.ExecService
}

func (e *pathCompletionEngine) Exec() runtimeapi.ExecService { return e.exec }

func formModelWithPathExec(t *testing.T, output string) *state.AppModel {
	t.Helper()
	m := formModel(t, &stubContainerService{})
	m.Connection.Engine = &pathCompletionEngine{advancedEngine: newAdvancedEngine(&stubContainerService{}), exec: pathExecService{output: output}}
	return m
}

// formModel returns a model wired to the given container service with a
// selected container in the containers panel, ready for form interactions.
func formModel(t *testing.T, svc runtimeapi.ContainerService) *state.AppModel {
	t.Helper()
	m := newAppModelWithEngine(newAdvancedEngine(svc))
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c1", Name: "web", Image: "nginx"}}
	return m
}

// requireForm asserts the expected form is open with the given field count
// and the default focus matches its FormSpec.Dangerous flag — Cancel slot
// for destructive forms, first field otherwise.
func requireForm(t *testing.T, m *state.AppModel, kind state.FormKind, fields int) {
	t.Helper()
	requireFormWithFocus(t, m, kind, fields, 0)
}

func requireFormWithFocus(t *testing.T, m *state.AppModel, kind state.FormKind, fields int, wantFocus int) {
	t.Helper()
	if m.Navigation.Mode != state.ModeContainerForm {
		t.Fatalf("mode = %v, want ModeContainerForm", m.Navigation.Mode)
	}
	if m.Form.Kind != kind {
		t.Fatalf("form kind = %v, want %v", m.Form.Kind, kind)
	}
	if len(m.Form.Fields) != fields {
		t.Fatalf("form fields = %d, want %d", len(m.Form.Fields), fields)
	}
	expected := wantFocus
	if wantFocus < 0 {
		expected = m.Form.CancelSlot()
	}
	if m.Form.FieldFocus != expected {
		t.Fatalf("default focus = %d, want %d", m.Form.FieldFocus, expected)
	}
}

// captureService records the Update / Commit options handed to the runtime so
// tests can assert exactly what the form submitted.
type captureService struct {
	*stubContainerService
	updateOpts runtimeapi.ContainerUpdateOptions
	commitOpts runtimeapi.ContainerCommitOptions
}

func (s *captureService) Update(ctx context.Context, id string, opts runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error) {
	s.updateOpts = opts
	return runtimeapi.ContainerUpdateResult{}, nil
}

func (s *captureService) Commit(ctx context.Context, id string, opts runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error) {
	s.commitOpts = opts
	return runtimeapi.ContainerCommitResult{ID: "img-new"}, nil
}

func TestOpenContainerUpdateForm(t *testing.T) {
	svc := &stubContainerService{inspectFn: func(_ context.Context, id string) (*runtimeapi.ContainerDetail, error) {
		if id != "c1" {
			t.Fatalf("inspect id = %q", id)
		}
		return &runtimeapi.ContainerDetail{Resources: runtimeapi.ContainerResources{
			Memory: 512 * 1024 * 1024, NanoCPUs: 1_500_000_000,
			RestartPolicy: "on-failure", MaximumRetryCount: 4,
		}}, nil
	}}
	m := formModel(t, svc)
	updated, cmd := openContainerUpdateForm(m)
	if cmd == nil || !updated.Form.Loading {
		t.Fatalf("open must inspect current configuration: loading=%v cmd=%v", updated.Form.Loading, cmd)
	}
	requireForm(t, updated, state.FormContainerUpdate, 4)
	msg := cmd().(state.ContainerUpdateConfigLoaded)
	updated, _ = HandleContainerUpdateConfigLoaded(updated, msg)
	if updated.Form.Loading {
		t.Fatal("inspect result must clear loading")
	}
	if got := updated.Form.Get(fieldMemory).Text(); got != "512" {
		t.Fatalf("memory = %q, want 512", got)
	}
	if got := updated.Form.Get(fieldCPUs).Text(); got != "1.5" {
		t.Fatalf("CPUs = %q, want 1.5", got)
	}
	restart := updated.Form.Get(fieldRestartPolicy)
	if restart == nil || restart.Kind() != state.FormSelect || len(restart.Options()) != len(restartPolicyChoices) {
		t.Fatalf("update restart field wrong: %#v", restart)
	}
	if restart.Value().(string) != "on-failure" || updated.Form.Get(fieldMaxRetries).Text() != "4" {
		t.Fatalf("restart config = %q/%q", restart.Value(), updated.Form.Get(fieldMaxRetries).Text())
	}
}

func TestContainerUpdateInspectDoesNotOverwriteTouchedField(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	memory := m.Form.Get(fieldMemory)
	memory.SetText("256")
	memory.SetTouched(true)
	msg := state.ContainerUpdateConfigLoaded{ContainerID: "c1", Detail: &runtimeapi.ContainerDetail{
		Resources: runtimeapi.ContainerResources{Memory: 512 * 1024 * 1024, NanoCPUs: 2_000_000_000, RestartPolicy: "always"},
	}}
	HandleContainerUpdateConfigLoaded(m, msg)
	if memory.Text() != "256" || m.Form.Get(fieldCPUs).Text() != "2" {
		t.Fatalf("inspect overwrote edited value or missed untouched value: memory=%q cpus=%q", memory.Text(), m.Form.Get(fieldCPUs).Text())
	}
}

func TestContainerUpdateInspectLeavesUnlimitedResourcesBlank(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	msg := state.ContainerUpdateConfigLoaded{ContainerID: "c1", Detail: &runtimeapi.ContainerDetail{
		Resources: runtimeapi.ContainerResources{RestartPolicy: "no"},
	}}
	HandleContainerUpdateConfigLoaded(m, msg)
	if memory, cpus := m.Form.Get(fieldMemory).Text(), m.Form.Get(fieldCPUs).Text(); memory != "" || cpus != "" {
		t.Fatalf("unlimited resources must stay blank: memory=%q cpus=%q", memory, cpus)
	}
	if retries := m.Form.Get(fieldMaxRetries).Text(); retries != "" {
		t.Fatalf("non-on-failure retries = %q, want blank", retries)
	}
}

func TestHandleContainerFormKeyEscCloses(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	updated, cmd := handleContainerFormKey(keys.KeyEsc, m)
	if cmd != nil {
		t.Fatalf("Esc returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeNormal || updated.Form.Kind != state.FormNone {
		t.Fatalf("Esc must close the form: mode=%v kind=%v", updated.Navigation.Mode, updated.Form.Kind)
	}
}

func TestHandleContainerFormKeyEnterOnCancelCloses(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	// Move focus to Cancel; Enter must dismiss the form without issuing a
	// runtime command.
	m.Form.FieldFocus = m.Form.CancelSlot()
	updated, cmd := handleContainerFormKey(keys.KeyEnter, m)
	if cmd != nil {
		t.Fatalf("Enter on Cancel returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeNormal || updated.Form.Kind != state.FormNone {
		t.Fatalf("Enter on Cancel must close: mode=%v kind=%v", updated.Navigation.Mode, updated.Form.Kind)
	}
}

func TestHandleContainerFormKeyEnterAdvancesFromField(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	// source field is the default focus; Enter advances to destination.
	updated, cmd := handleContainerFormKey(keys.KeyEnter, m)
	if cmd != nil {
		t.Fatalf("Enter on field returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Form.FieldFocus != 1 {
		t.Fatalf("Enter should advance to destination: mode=%v focus=%d", updated.Navigation.Mode, updated.Form.FieldFocus)
	}
}

func TestSubmitContainerCopyFormValidation(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	// Focus on Confirm then submit with empty required fields.
	m.Form.FieldFocus = m.Form.ConfirmSlot()
	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("empty copy form returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm {
		t.Fatal("validation failure must keep the form open")
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("validation failure must surface a warning toast")
	}
	if updated.Feedback.ToastLevel != state.NotificationWarning {
		t.Fatalf("toast level = %v, want warning", updated.Feedback.ToastLevel)
	}
}

func TestSubmitContainerCommitFormSuccess(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerCommitForm(m)
	m.Form.Get(fieldRepository).SetText("registry.example/app")
	m.Form.Get(fieldTag).SetText("v2")
	m.Form.Get(fieldComment).SetText("pinned build")
	// Pause stays at its default true.
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd == nil {
		t.Fatal("valid commit form must return a command")
	}
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatal("successful commit must close the form")
	}
	msg := cmd()
	done, ok := msg.(ContainerCommitDone)
	if !ok {
		t.Fatalf("commit cmd returned %T, want ContainerCommitDone", msg)
	}
	if done.Error != nil {
		t.Fatalf("unexpected commit error: %v", done.Error)
	}
	opts := capture.commitOpts
	if opts.Repository != "registry.example/app" || opts.Tag != "v2" || opts.Comment != "pinned build" {
		t.Fatalf("commit opts = %#v", opts)
	}
	if !opts.Pause {
		t.Fatal("commit must default to pausing the container")
	}
}

func TestSubmitContainerCommitFormDefaultsTagToLatest(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerCommitForm(m)
	m.Form.Get(fieldRepository).SetText("registry.example/app")
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	_, cmd := submitContainerForm(m)
	if cmd == nil {
		t.Fatal("commit with an omitted tag must return a command")
	}
	cmd()
	if capture.commitOpts.Tag != "latest" {
		t.Fatalf("default tag = %q, want latest", capture.commitOpts.Tag)
	}
}

func TestContainerPathTabCompletesSingleDirectory(t *testing.T) {
	m := formModelWithPathExec(t, "d|etc|drwxr-xr-x|root|root|4096|1700000000|\r\nf|entrypoint.sh|-rw-r--r--|root|root|100|1700000000|\r\n")
	openContainerCopyForm(m)
	source := m.Form.Get(fieldSourcePath)
	pf := source.(*dialog.PathField)
	pf.SetText("/et")

	updated, cmd := handleContainerFormKey(keys.KeyTab, m)
	if cmd == nil || !pf.PathLoading() {
		t.Fatalf("Tab must start async completion: loading=%v cmd=%v", pf.PathLoading(), cmd)
	}
	msg, ok := cmd().(state.ContainerPathCompleted)
	if !ok {
		t.Fatalf("completion command returned %T", cmd())
	}
	updated, _ = HandleContainerPathCompleted(updated, msg)
	if pf.TextRaw() != "/etc/" || pf.PathLoading() || updated.Form.Popup.Open {
		t.Fatalf("single directory completion: input=%q loading=%v popup=%#v", pf.TextRaw(), pf.PathLoading(), updated.Form.Popup)
	}
}

func TestContainerPathCtrlSpaceOpensDirectoryPopup(t *testing.T) {
	m := formModelWithPathExec(t, "d|nginx|drwxr-xr-x|root|root|4096|1700000000|\nf|hosts|-rw-r--r--|root|root|200|1700000000|\nf|resolv.conf|-rw-r--r--|root|root|100|1700000000|\n")
	openContainerCopyForm(m)
	source := m.Form.Get(fieldSourcePath)
	pf := source.(*dialog.PathField)
	pf.SetText("/etc/")

	updated, cmd := handleContainerFormKey(keys.KeyCtrlSpace, m)
	if cmd == nil {
		t.Fatal("Ctrl+Space must start container completion")
	}
	msg := cmd().(state.ContainerPathCompleted)
	updated, _ = HandleContainerPathCompleted(updated, msg)
	if !updated.Form.Popup.Open || updated.Form.Popup.Kind != state.PopupPath || len(pf.Suggestions()) != 3 {
		t.Fatalf("container popup = %#v suggestions=%#v", updated.Form.Popup, pf.Suggestions())
	}
	if !pf.Suggestions()[0].IsDir || pf.Suggestions()[0].Path != "/etc/nginx" {
		t.Fatalf("directory must sort first: %#v", pf.Suggestions())
	}
}

func TestContainerPathCompletionIgnoresStaleInput(t *testing.T) {
	m := formModelWithPathExec(t, "d\tetc\n")
	openContainerCopyForm(m)
	source := m.Form.Get(fieldSourcePath)
	pf := source.(*dialog.PathField)
	pf.SetText("/et")
	_, cmd := handleContainerFormKey(keys.KeyTab, m)
	msg := cmd().(state.ContainerPathCompleted)
	pf.SetText("/var")
	pf.SetPathLoading(false)
	HandleContainerPathCompleted(m, msg)
	if pf.TextRaw() != "/var" || len(pf.Suggestions()) != 0 {
		t.Fatalf("stale completion changed field: %#v", pf)
	}
}
