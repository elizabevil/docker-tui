package keyboard

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
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

// requireForm asserts the expected form is open with the default focus on the
// Cancel slot (the safest default) and the given field count.
func requireForm(t *testing.T, m *state.AppModel, kind state.FormKind, fields int) {
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
	if m.Form.FieldFocus != m.Form.CancelSlot() {
		t.Fatalf("default focus must be Cancel: FieldFocus=%d", m.Form.FieldFocus)
	}
}

func TestOpenContainerCopyForm(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	updated, cmd := openContainerCopyForm(m)
	if cmd != nil {
		t.Fatalf("open returned a cmd %T, want nil", cmd)
	}
	requireForm(t, updated, state.FormContainerCopy, 2)
	if updated.Form.Get(fieldSourcePath) == nil || updated.Form.Get(fieldDestinationPath) == nil {
		t.Fatalf("copy form missing source/destination fields: %#v", updated.Form.Fields)
	}
	if source := updated.Form.Get(fieldSourcePath); source.Kind != state.FormPath || source.PathSource != state.PathContainer {
		t.Fatalf("copy source must be a container path: %#v", source)
	}
}

func TestOpenContainerExportForm(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	updated, cmd := openContainerExportForm(m)
	if cmd != nil {
		t.Fatalf("open returned a cmd %T, want nil", cmd)
	}
	requireForm(t, updated, state.FormContainerExport, 1)
	if updated.Form.Get(fieldDestinationPath) == nil {
		t.Fatal("export form missing destination field")
	}
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
	if restart == nil || restart.Kind != state.FormSelect || len(restart.Options) != len(restartPolicyChoices) {
		t.Fatalf("update restart field wrong: %#v", restart)
	}
	if restart.Option() != "on-failure" || updated.Form.Get(fieldMaxRetries).Text() != "4" {
		t.Fatalf("restart config = %q/%q", restart.Option(), updated.Form.Get(fieldMaxRetries).Text())
	}
}

func TestContainerUpdateInspectDoesNotOverwriteTouchedField(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	memory := m.Form.Get(fieldMemory)
	memory.Input.Set("256")
	memory.Touched = true
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

func TestContainerUpdateTabCyclesFocus(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	if m.Form.FocusedButton() != keys.ShowOptionCancel {
		t.Fatal("setup must start on Cancel")
	}
	handleContainerFormKey(keys.KeyTab, m)
	if m.Form.FieldFocus != 0 {
		t.Fatalf("Tab from Cancel = field %d, want 0", m.Form.FieldFocus)
	}
	handleContainerFormKey(keys.KeyTab, m)
	if m.Form.FieldFocus != 1 {
		t.Fatalf("second Tab = field %d, want 1", m.Form.FieldFocus)
	}
	handleContainerFormKey(keys.KeyShiftTab, m)
	if m.Form.FieldFocus != 0 {
		t.Fatalf("Shift+Tab = field %d, want 0", m.Form.FieldFocus)
	}
}

func TestOpenContainerCommitForm(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	updated, cmd := openContainerCommitForm(m)
	if cmd != nil {
		t.Fatalf("open returned a cmd %T, want nil", cmd)
	}
	requireForm(t, updated, state.FormContainerCommit, 7)
	pause := updated.Form.Get(fieldPause)
	if pause == nil || pause.Kind != state.FormBool || !pause.Toggle {
		t.Fatalf("commit pause field must default to true: %#v", pause)
	}
	if updated.Form.Get(fieldRepository).Text() != "nginx" || updated.Form.Get(fieldTag).Text() != "latest" {
		t.Fatalf("commit image defaults = %q:%q", updated.Form.Get(fieldRepository).Text(), updated.Form.Get(fieldTag).Text())
	}
	if export := updated.Form.Get(fieldExportTar); export == nil || export.Toggle {
		t.Fatalf("export tar must default false: %#v", export)
	}
	if archive := updated.Form.Get(fieldArchivePath); archive == nil || archive.Text() == "" {
		t.Fatalf("commit archive must have a default path: %#v", archive)
	}
}

func TestOpenContainerFormGuards(t *testing.T) {
	svc := &stubContainerService{}

	// No engine.
	m := formModel(t, svc)
	m.Connection.Engine = nil
	if updated, cmd := openContainerCopyForm(m); cmd != nil || updated.Navigation.Mode == state.ModeContainerForm {
		t.Fatal("copy form must no-op without an engine")
	}

	// Wrong panel.
	m2 := formModel(t, svc)
	m2.Navigation.ActivePanel = state.PanelImages
	if updated, cmd := openContainerUpdateForm(m2); cmd != nil || updated.Navigation.Mode == state.ModeContainerForm {
		t.Fatal("update form must no-op off the containers panel")
	}

	// No selection.
	m3 := formModel(t, svc)
	m3.Resources.Containers.Items = nil
	if updated, cmd := openContainerCommitForm(m3); cmd != nil || updated.Navigation.Mode == state.ModeContainerForm {
		t.Fatal("commit form must no-op without a selected container")
	}
}

func TestHandleContainerFormKeyFocusLoop(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	// copy form slots: field0, field1, Confirm, Cancel. Default focus is Cancel.

	tab := func() *state.AppModel {
		t.Helper()
		updated, _ := handleContainerFormKey(keys.KeyTab, m)
		return updated
	}
	shiftTab := func() *state.AppModel {
		t.Helper()
		updated, _ := handleContainerFormKey(keys.KeyShiftTab, m)
		return updated
	}
	down := func() *state.AppModel {
		t.Helper()
		updated, _ := handleContainerFormKey(keys.KeyDown, m)
		return updated
	}
	up := func() *state.AppModel {
		t.Helper()
		updated, _ := handleContainerFormKey(keys.KeyUp, m)
		return updated
	}

	// Up from default Cancel returns to the Confirm row (BR-043 §3.3 linear model).
	m = up()
	if m.Form.FieldFocus != m.Form.ConfirmSlot() {
		t.Fatalf("Up from Cancel = %d, want Confirm (%d)", m.Form.FieldFocus, m.Form.ConfirmSlot())
	}
	// Up from Confirm returns to the last field (index 1).
	m = up()
	if m.Form.FieldFocus != 1 {
		t.Fatalf("Up from Confirm = %d, want last field 1", m.Form.FieldFocus)
	}
	// Up from field 1 moves to field 0.
	m = up()
	if m.Form.FieldFocus != 0 {
		t.Fatalf("Up from field 1 = %d, want field 0", m.Form.FieldFocus)
	}
	// Up from field 0 wraps to Cancel.
	m = up()
	if m.Form.FieldFocus != m.Form.CancelSlot() {
		t.Fatalf("Up from first field = focus=%d, want Cancel (%d)", m.Form.FieldFocus, m.Form.CancelSlot())
	}
	// Down from Cancel wraps to field 0.
	m = down()
	if m.Form.FieldFocus != 0 {
		t.Fatalf("Down from Cancel wraps = %d, want field 0", m.Form.FieldFocus)
	}
	// Tab never moves focus in non-Update forms.
	m = tab()
	if m.Form.FieldFocus != 0 {
		t.Fatalf("Tab on field in Copy form must not move focus: focus=%d", m.Form.FieldFocus)
	}
	// Shift+Tab same.
	m = shiftTab()
	if m.Form.FieldFocus != 0 {
		t.Fatalf("Shift+Tab on field in Copy form must not move focus: focus=%d", m.Form.FieldFocus)
	}
	// Down from field 0 returns to field 1.
	m = down()
	if m.Form.FieldFocus != 1 {
		t.Fatalf("Down from field 0 = %d, want field 1", m.Form.FieldFocus)
	}
	// Down from last field enters Confirm.
	m = down()
	if m.Form.FieldFocus != m.Form.ConfirmSlot() {
		t.Fatalf("Down from last = focus=%d, want Confirm (%d)", m.Form.FieldFocus, m.Form.ConfirmSlot())
	}
	// Down from Confirm enters Cancel.
	m = down()
	if m.Form.FieldFocus != m.Form.CancelSlot() {
		t.Fatalf("Down from Confirm = focus=%d, want Cancel (%d)", m.Form.FieldFocus, m.Form.CancelSlot())
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
	// Default focus is Cancel, so Enter must simply dismiss the form without
	// issuing any runtime command.
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
	m.Form.MoveField(1) // Cancel -> source field.

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

func TestSubmitContainerCopyFormSuccess(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	m.Form.Get(fieldSourcePath).Input.Set("/etc/app.conf")
	m.Form.Get(fieldDestinationPath).Input.Set("/tmp/app.tar")
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd == nil {
		t.Fatal("valid copy form must return a command")
	}
	if updated.Navigation.Mode != state.ModeNormal || updated.Form.Kind != state.FormNone {
		t.Fatalf("successful submit must close the form: mode=%v kind=%v", updated.Navigation.Mode, updated.Form.Kind)
	}
	msg := cmd()
	if _, ok := msg.(ContainerCopyDone); !ok {
		t.Fatalf("copy cmd returned %T, want ContainerCopyDone", msg)
	}
}

func TestSubmitContainerFormKeepsOriginalTargetAfterListRefresh(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	m.Form.Get(fieldSourcePath).Input.Set("/etc/app.conf")
	m.Form.Get(fieldDestinationPath).Input.Set("/tmp/app.tar")
	m.Form.FieldFocus = m.Form.ConfirmSlot()
	// Simulate an event-driven refresh changing the selected row while the
	// form is open. Submission must retain the container captured at Open.
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c2", Name: "worker"}}

	_, cmd := submitContainerForm(m)
	done, ok := cmd().(ContainerCopyDone)
	if !ok || done.ContainerID != "c1" {
		t.Fatalf("copy target = %#v, want original container c1", done)
	}
}

func TestSubmitContainerExportFormValidation(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	// Export autofills a default destination name; clear it to test validation.
	m.Form.Get(fieldDestinationPath).Input.Set("")
	m.Form.FieldFocus = m.Form.ConfirmSlot()
	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("empty export form returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Feedback.ToastMessage == "" {
		t.Fatal("export validation failure must keep the form open with a toast")
	}
}

func TestSubmitContainerUpdateFormParsesOptions(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerUpdateForm(m)
	m.Form.Get(fieldMemory).Input.Set("256")
	m.Form.Get(fieldMemory).Touched = true
	m.Form.Get(fieldCPUs).Input.Set("1.5")
	m.Form.Get(fieldCPUs).Touched = true
	m.Form.Get(fieldRestartPolicy).Index = 2 // "always"
	m.Form.Get(fieldRestartPolicy).Touched = true
	m.Form.Get(fieldMaxRetries).Input.Set("3")
	m.Form.Get(fieldMaxRetries).Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd == nil {
		t.Fatal("valid update form must return a command")
	}
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatal("successful update must close the form")
	}
	msg := cmd()
	done, ok := msg.(ContainerUpdateDone)
	if !ok {
		t.Fatalf("update cmd returned %T, want ContainerUpdateDone", msg)
	}
	if done.Error != nil {
		t.Fatalf("unexpected update error: %v", done.Error)
	}
	opts := capture.updateOpts
	if opts.Memory == nil || *opts.Memory != 256*1024*1024 {
		t.Fatalf("memory = %v, want 256 MiB in bytes", opts.Memory)
	}
	if opts.NanoCPUs == nil || *opts.NanoCPUs != int64(1.5*1e9) {
		t.Fatalf("nano CPUs = %v, want 1.5 cores", opts.NanoCPUs)
	}
	if opts.RestartPolicy == nil || *opts.RestartPolicy != "always" {
		t.Fatalf("restart policy = %v, want always", opts.RestartPolicy)
	}
	if opts.RestartMaxRetries != nil {
		t.Fatalf("restart max retries must be nil for non-on-failure policy, got %d", *opts.RestartMaxRetries)
	}
}

func TestSubmitContainerUpdateFormOnFailureRetries(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerUpdateForm(m)
	m.Form.Get(fieldRestartPolicy).Index = 4 // "on-failure"
	m.Form.Get(fieldRestartPolicy).Touched = true
	m.Form.Get(fieldMaxRetries).Input.Set("5")
	m.Form.Get(fieldMaxRetries).Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	_, cmd := submitContainerForm(m)
	if cmd == nil {
		t.Fatal("on-failure update must return a command")
	}
	cmd()
	opts := capture.updateOpts
	if opts.RestartPolicy == nil || *opts.RestartPolicy != "on-failure" {
		t.Fatalf("restart policy = %v, want on-failure", opts.RestartPolicy)
	}
	if opts.RestartMaxRetries == nil || *opts.RestartMaxRetries != 5 {
		t.Fatalf("restart max retries = %v, want 5", opts.RestartMaxRetries)
	}
}

func TestSubmitContainerUpdateFormInvalidInput(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	m.Form.Get(fieldMemory).Input.Set("many")
	m.Form.Get(fieldMemory).Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("invalid memory returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Feedback.ToastMessage == "" {
		t.Fatal("invalid update input must keep the form open with a toast")
	}
}

func TestSubmitContainerUpdateFormRequiresAChange(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("empty update returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Feedback.ToastMessage == "" {
		t.Fatal("empty update must keep the form open with a warning")
	}
}

func TestMemoryFieldHasConstraints(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	mem := m.Form.Get(fieldMemory)
	if mem == nil {
		t.Fatal("memory field must exist on the update form")
	}
	if mem.HelperText != "MB" {
		t.Errorf("memory HelperText = %q, want %q", mem.HelperText, "MB")
	}
	if mem.Unit != "MB" {
		t.Errorf("memory Unit = %q, want %q", mem.Unit, "MB")
	}
	if mem.Min == nil || *mem.Min != 0 {
		t.Errorf("memory Min = %v, want pointer to 0", mem.Min)
	}
	if mem.Max == nil {
		t.Fatal("memory Max must be set")
	}
}

func TestCpusFieldHasConstraints(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	cpu := m.Form.Get(fieldCPUs)
	if cpu == nil {
		t.Fatal("cpus field must exist on the update form")
	}
	if cpu.HelperText != "cores" {
		t.Errorf("cpus HelperText = %q, want %q", cpu.HelperText, "cores")
	}
	if cpu.Unit != "cores" {
		t.Errorf("cpus Unit = %q, want %q", cpu.Unit, "cores")
	}
	if cpu.Min == nil || *cpu.Min != 0 {
		t.Errorf("cpus Min = %v, want pointer to 0", cpu.Min)
	}
	if cpu.Max == nil {
		t.Fatal("cpus Max must be set")
	}
}

func TestMemoryFieldSubmitRejectsNegative(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	mem := m.Form.Get(fieldMemory)
	mem.Input.Set("-1")
	mem.Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("negative memory returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm {
		t.Fatal("negative memory must keep the form open")
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("negative memory must surface a toast")
	}
}

func TestMemoryFieldSubmitRejectsExcessive(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	mem := m.Form.Get(fieldMemory)
	mem.Input.Set("9999999999999")
	mem.Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("excessive memory returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm {
		t.Fatal("excessive memory must keep the form open")
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("excessive memory must surface a toast")
	}
}

func TestCpusFieldSubmitRejectsNegative(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	cpu := m.Form.Get(fieldCPUs)
	cpu.Input.Set("-0.5")
	cpu.Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("negative cpus returned a cmd %T, want nil", cmd)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("negative cpus must surface a toast")
	}
}

func TestCpusFieldSubmitRejectsExcessive(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	cpu := m.Form.Get(fieldCPUs)
	cpu.Input.Set("99999999999")
	cpu.Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("excessive cpus returned a cmd %T, want nil", cmd)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("excessive cpus must surface a toast")
	}
}

func TestMemoryFieldSubmitAcceptsValid(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerUpdateForm(m)
	mem := m.Form.Get(fieldMemory)
	mem.Input.Set("512")
	mem.Touched = true
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd == nil {
		t.Fatal("valid memory must return a submit command")
	}
	if updated.Feedback.ToastMessage != "" {
		t.Fatalf("valid memory must not show a toast, got %q", updated.Feedback.ToastMessage)
	}
	if updated.Navigation.Mode == state.ModeContainerForm {
		t.Fatal("valid memory submit must close the form")
	}
}

func TestSubmitContainerCommitFormRequiresRepository(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCommitForm(m)
	m.Form.Get(fieldRepository).Input.Set("")
	m.Form.FieldFocus = m.Form.ConfirmSlot()
	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("empty commit repo returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Feedback.ToastMessage == "" {
		t.Fatal("commit validation failure must keep the form open with a toast")
	}
}

func TestSubmitContainerCommitFormSuccess(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerCommitForm(m)
	m.Form.Get(fieldRepository).Input.Set("registry.example/app")
	m.Form.Get(fieldTag).Input.Set("v2")
	m.Form.Get(fieldComment).Input.Set("pinned build")
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
	m.Form.Get(fieldRepository).Input.Set("registry.example/app")
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

func TestEditFormFieldSpaceTogglesBool(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCommitForm(m)
	pause := m.Form.Get(fieldPause)
	// Focus the pause field (index 4).
	m.Form.MoveField(5)
	if m.Form.Field() == nil || m.Form.Field().Key != fieldPause {
		t.Fatalf("focus = %#v, want pause field", m.Form.Field())
	}
	before := pause.Toggle
	handleContainerFormKey(keys.KeySpace, m)
	if pause.Toggle == before {
		t.Fatal("Space must toggle a bool field")
	}
	// A normal character must be ignored on a bool field.
	handleContainerFormKey("x", m)
	if pause.Toggle == before {
		t.Fatal("typing on a bool field must not re-toggle it")
	}
}

func TestFormBoolIgnoresLeftRight(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCommitForm(m)
	m.Form.FieldFocus = 4
	pause := m.Form.Get(fieldPause)
	before := pause.Toggle
	handleContainerFormKey(keys.KeyLeft, m)
	handleContainerFormKey(keys.KeyRight, m)
	if pause.Toggle != before {
		t.Fatal("Bool must only toggle with Space")
	}
}

func TestExportTarBoolTogglesThroughBubbleTeaSpace(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCommitForm(m)
	m.Form.FieldFocus = 5
	export := m.Form.Get(fieldExportTar)
	if export == nil || export.Toggle {
		t.Fatalf("export setup = %#v", export)
	}
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeySpace})
	updated, cmd := HandleKeyPress(msg, m)
	if cmd != nil {
		t.Fatalf("Space toggle returned command %T", cmd)
	}
	if !updated.Form.Get(fieldExportTar).Toggle {
		t.Fatalf("Bubble Tea key %q did not toggle Export tar", msg.String())
	}
}

func TestNormalizeInputKeyMapsNamedSpace(t *testing.T) {
	if got := normalizeInputKey(keys.KeySpaceName); got != keys.KeySpace {
		t.Fatalf("normalizeInputKey(space) = %q, want a literal space", got)
	}
}

func TestSubmitContainerCommitWithImageExport(t *testing.T) {
	capture := &captureService{}
	m := formModel(t, capture)
	openContainerCommitForm(m)
	destination := filepath.Join(t.TempDir(), "snapshot.tar")
	m.Form.Get(fieldExportTar).Toggle = true
	m.Form.Get(fieldArchivePath).Input.Set(destination)
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd == nil || updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("commit export did not start: mode=%v cmd=%v", updated.Navigation.Mode, cmd)
	}
	done, ok := cmd().(ContainerCommitDone)
	if !ok {
		t.Fatalf("commit export command returned %T", cmd())
	}
	if done.ExportPath != destination || done.ImageID != "img-new" {
		t.Fatalf("commit export result = %#v", done)
	}
}

func TestStepFormSelectLeftRightAreNoOp(t *testing.T) {
	// BR-041 §3.1 / §3.4: Left/Right must NOT cycle FormSelect options; the
	// Select field is changed by opening the popup with Enter and using
	// Up/Down/Enter there.
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	// Focus the restart select (index 2). From Cancel (len+1 = 5) Up returns
	// to the last field, then a Down re-enters Cancel; instead we walk the
	// fields directly.
	m.Form.FieldFocus = 2
	f := m.Form.Field()
	if f == nil || f.Key != fieldRestartPolicy {
		t.Fatalf("focus = %#v, want restart select", f)
	}
	start := f.Index
	handleContainerFormKey(keys.KeyRight, m)
	if f.Index != start {
		t.Fatalf("Right on Select must not change index: %d -> %d", start, f.Index)
	}
	handleContainerFormKey(keys.KeyLeft, m)
	if f.Index != start {
		t.Fatalf("Left on Select must not change index: %d -> %d", start, f.Index)
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

func TestOpenContainerCopyFormPathFields(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	src := m.Form.Get(fieldSourcePath)
	dst := m.Form.Get(fieldDestinationPath)
	if src.Kind != state.FormPath || src.PathSource != state.PathContainer {
		t.Fatalf("copy source must be a container path field: %#v", src)
	}
	if dst.Kind != state.FormPath || dst.PathSource != state.PathLocal || dst.PathMode != state.PathSaveFile {
		t.Fatalf("copy destination must be a local save-file path field: %#v", dst)
	}
	if m.Form.CWD == "" {
		t.Fatal("copy form must capture the working directory")
	}
	// The destination is pre-filled with a default name (BR-041 §4.1).
	if dst.Text() == "" {
		t.Fatal("copy destination must be pre-filled with a default tar name")
	}
}

func TestOpenContainerExportFormAutofillsDestination(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	if dst.Kind != state.FormPath || dst.PathSource != state.PathLocal || dst.PathMode != state.PathSaveFile {
		t.Fatalf("export destination must be a local save-file path field: %#v", dst)
	}
	if dst.Text() == "" {
		t.Fatal("export destination must be pre-filled with a default tar name")
	}
}

func TestFormPathTabCompletionAppliesSingleCandidate(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	// Focus the destination field (index 0).
	m.Form.MoveField(1)
	// Seed a single directory candidate.
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/backup.tar", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst.Input.Set(dir + "/")
	dst.Touched = false
	completeForField(m, dst)
	if len(dst.Suggestions) != 1 {
		t.Fatalf("suggestions = %d, want 1 (temp dir with one file)", len(dst.Suggestions))
	}
	before := dst.Text()
	updated, cmd := handleContainerFormKey(keys.KeyTab, m)
	if cmd != nil {
		t.Fatalf("Tab returned a cmd %T, want nil", cmd)
	}
	if updated.Form.Get(fieldDestinationPath).Text() == before {
		t.Fatal("Tab must fill the single path candidate into the field")
	}
}

func TestFormPathTabMultipleCandidatesOpensPopup(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	m.Form.MoveField(1)
	dir := t.TempDir()
	for _, name := range []string{"alpha.tar", "alpine.tar"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dst.Input.Set(filepath.Join(dir, "al"))

	updated, _ := handleContainerFormKey(keys.KeyTab, m)
	if updated.Form.Popup.Open || !strings.HasSuffix(dst.Text(), "alp") {
		t.Fatalf("first Tab must only extend the common prefix: input=%q popup=%#v", dst.Text(), updated.Form.Popup)
	}
	updated, _ = handleContainerFormKey(keys.KeyTab, updated)
	if !updated.Form.Popup.Open || updated.Form.Popup.Kind != state.PopupPath {
		t.Fatalf("multiple candidates must open path popup: %#v", updated.Form.Popup)
	}
}

func TestCopySourceChangeRegeneratesUntouchedDestination(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	m.Form.MoveField(1) // source
	dst := m.Form.Get(fieldDestinationPath)
	before := dst.Text()
	for _, key := range []string{"/", "e", "t", "c", "/", "a", "p", "p", ".", "c", "o", "n", "f"} {
		handleContainerFormKey(key, m)
	}
	if dst.Text() == before || !strings.Contains(dst.Text(), "app.conf") {
		t.Fatalf("source change did not regenerate destination: before=%q after=%q", before, dst.Text())
	}
	dst.Touched = true
	dst.Input.Set("/custom/output.tar")
	handleContainerFormKey("x", m)
	if dst.Text() != "/custom/output.tar" {
		t.Fatalf("source edit overwrote touched destination: %q", dst.Text())
	}
}

func TestCopySourceBackspaceRegeneratesUntouchedDestination(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	m.Form.MoveField(-1) // Cancel -> destination.
	m.Form.MoveField(-1) // destination -> source.
	src := m.Form.Get(fieldSourcePath)
	dst := m.Form.Get(fieldDestinationPath)
	src.Input.Set("/etc/first.conf")
	prefillDefaultDestination(m, m.Form.CWD, shortContainerID(m.Form.TargetID))
	before := dst.Text()

	handleContainerFormKey(keys.KeyBackspace, m)
	if dst.Text() == before || !strings.Contains(dst.Text(), "first.con") {
		t.Fatalf("Backspace did not refresh default destination: before=%q after=%q", before, dst.Text())
	}
}

func TestCopySourceBlurDoesNotAbsolutizeContainerPath(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	m.Form.FieldFocus = 0 // source, a container path field.
	src := m.Form.Get(fieldSourcePath)
	src.Input.Set("etc/app.conf")

	updated, _ := handleContainerFormKey(keys.KeyDown, m)
	if updated.Form.FieldFocus != 1 {
		t.Fatalf("focus = %d, want destination field", updated.Form.FieldFocus)
	}
	if got := src.Text(); got != "etc/app.conf" {
		t.Fatalf("container source path was rewritten on blur: got %q", got)
	}
}

func TestPathBackspaceRefreshesCompletionState(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	// Export form has one field (destination). Move from default Cancel to the
	// destination field via linear wrap (BR-043 §3.3).
	m.Form.FieldFocus = 0
	dst := m.Form.Get(fieldDestinationPath)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.tar"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst.Input.Set(filepath.Join(dir, "alphx"))
	handleContainerFormKey(keys.KeyBackspace, m)
	if !dst.Touched || len(dst.Suggestions) != 1 || dst.Suggestions[0].Name != "alpha.tar" {
		t.Fatalf("Backspace completion state: touched=%v suggestions=%#v", dst.Touched, dst.Suggestions)
	}
}

func TestNormalizeSaveDestinationUsesCWD(t *testing.T) {
	dir := t.TempDir()
	got, err := normalizeSaveDestination("output.tar", dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(dir, "output.tar") {
		t.Fatalf("destination = %q, want %q", got, filepath.Join(dir, "output.tar"))
	}
}

func TestFormPathCtrlSpaceOpensPopup(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	m.Form.MoveField(1) // focus destination
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/backup.tar", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst.Input.Set(dir + "/")
	completeForField(m, dst)
	if len(dst.Suggestions) == 0 {
		t.Fatal("setup must produce at least one suggestion")
	}

	updated, cmd := handleContainerFormKey(keys.KeyCtrlSpace, m)
	if cmd != nil {
		t.Fatalf("Ctrl+Space returned a cmd %T, want nil", cmd)
	}
	if !updated.Form.Popup.Open || updated.Form.Popup.Kind != state.PopupPath {
		t.Fatalf("Ctrl+Space must open the path popup: %#v", updated.Form.Popup)
	}
	updated, _ = handleContainerFormKey(keys.KeyEsc, m)
	if updated.Form.Popup.Open {
		t.Fatal("Esc must close the popup")
	}
}

func TestContainerPathCompletionFallsBackToManualInput(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	m.Form.MoveField(1)
	source := m.Form.Get(fieldSourcePath)
	source.Input.Set("/etc/ng")

	updated, cmd := handleContainerFormKey(keys.KeyTab, m)
	if cmd != nil || source.Input.Text != "/etc/ng" || updated.Form.Popup.Open {
		t.Fatalf("Tab changed unsupported container path: input=%q popup=%#v cmd=%v", source.Input.Text, updated.Form.Popup, cmd)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("Tab must explain that container path completion is unavailable")
	}
	updated.Feedback.ToastMessage = ""
	updated, _ = handleContainerFormKey(keys.KeyCtrlSpace, updated)
	if updated.Form.Popup.Open || updated.Feedback.ToastMessage == "" {
		t.Fatalf("Ctrl+Space must keep manual input and show a warning: popup=%#v toast=%q", updated.Form.Popup, updated.Feedback.ToastMessage)
	}
}

func TestContainerPathTabCompletesSingleDirectory(t *testing.T) {
	m := formModelWithPathExec(t, "d|etc|drwxr-xr-x|root|root|4096|1700000000|\r\nf|entrypoint.sh|-rw-r--r--|root|root|100|1700000000|\r\n")
	openContainerCopyForm(m)
	m.Form.MoveField(1)
	source := m.Form.Get(fieldSourcePath)
	source.Input.Set("/et")

	updated, cmd := handleContainerFormKey(keys.KeyTab, m)
	if cmd == nil || !source.PathLoading {
		t.Fatalf("Tab must start async completion: loading=%v cmd=%v", source.PathLoading, cmd)
	}
	msg, ok := cmd().(state.ContainerPathCompleted)
	if !ok {
		t.Fatalf("completion command returned %T", cmd())
	}
	updated, _ = HandleContainerPathCompleted(updated, msg)
	if source.Input.Text != "/etc/" || source.PathLoading || updated.Form.Popup.Open {
		t.Fatalf("single directory completion: input=%q loading=%v popup=%#v", source.Input.Text, source.PathLoading, updated.Form.Popup)
	}
}

func TestContainerPathCtrlSpaceOpensDirectoryPopup(t *testing.T) {
	m := formModelWithPathExec(t, "d|nginx|drwxr-xr-x|root|root|4096|1700000000|\nf|hosts|-rw-r--r--|root|root|200|1700000000|\nf|resolv.conf|-rw-r--r--|root|root|100|1700000000|\n")
	openContainerCopyForm(m)
	m.Form.MoveField(1)
	source := m.Form.Get(fieldSourcePath)
	source.Input.Set("/etc/")

	updated, cmd := handleContainerFormKey(keys.KeyCtrlSpace, m)
	if cmd == nil {
		t.Fatal("Ctrl+Space must start container completion")
	}
	msg := cmd().(state.ContainerPathCompleted)
	updated, _ = HandleContainerPathCompleted(updated, msg)
	if !updated.Form.Popup.Open || updated.Form.Popup.Kind != state.PopupPath || len(source.Suggestions) != 3 {
		t.Fatalf("container popup = %#v suggestions=%#v", updated.Form.Popup, source.Suggestions)
	}
	if !source.Suggestions[0].IsDir || source.Suggestions[0].Path != "/etc/nginx" {
		t.Fatalf("directory must sort first: %#v", source.Suggestions)
	}
}

func TestContainerPathCompletionIgnoresStaleInput(t *testing.T) {
	m := formModelWithPathExec(t, "d\tetc\n")
	openContainerCopyForm(m)
	m.Form.MoveField(1)
	source := m.Form.Get(fieldSourcePath)
	source.Input.Set("/et")
	_, cmd := handleContainerFormKey(keys.KeyTab, m)
	msg := cmd().(state.ContainerPathCompleted)
	source.Input.Set("/var")
	source.PathLoading = false
	HandleContainerPathCompleted(m, msg)
	if source.Input.Text != "/var" || len(source.Suggestions) != 0 {
		t.Fatalf("stale completion changed field: %#v", source)
	}
}

func TestCtrlSpaceKeyNameMatchesBubbleTea(t *testing.T) {
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeySpace, Mod: tea.ModCtrl})
	if got := keys.Normalize(msg.String()); got != keys.KeyCtrlSpace {
		t.Fatalf("Bubble Tea Ctrl+Space = %q, configured key = %q", got, keys.KeyCtrlSpace)
	}
}

func TestFormSelectPopupNavigationAndCommit(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	// Focus the restart select (index 2).
	m.Form.MoveField(3)
	if f := m.Form.Field(); f == nil || f.Key != fieldRestartPolicy {
		t.Fatalf("focus = %#v, want restart select", m.Form.Field())
	}
	// Enter opens the popup.
	updated, cmd := handleContainerFormKey(keys.KeyEnter, m)
	if cmd != nil {
		t.Fatalf("Enter returned a cmd %T, want nil", cmd)
	}
	if !updated.Form.Popup.Open || updated.Form.Popup.Kind != state.PopupSelect {
		t.Fatalf("Enter must open the select popup: %#v", updated.Form.Popup)
	}
	// Down moves the cursor.
	updated, _ = handleContainerFormKey(keys.KeyDown, m)
	if updated.Form.Popup.Cursor != 1 {
		t.Fatalf("Down cursor = %d, want 1", updated.Form.Popup.Cursor)
	}
	// Enter commits the highlighted option.
	updated, cmd = handleContainerFormKey(keys.KeyEnter, m)
	if cmd != nil {
		t.Fatalf("popup Enter returned a cmd %T, want nil", cmd)
	}
	if updated.Form.Popup.Open {
		t.Fatal("popup Enter must close the popup")
	}
	if f := updated.Form.Get(fieldRestartPolicy); f.Option() != f.Options[1] {
		t.Fatalf("selected option = %q, want %q", f.Option(), f.Options[1])
	}
}

func TestSubmitContainerCopyOverwriteConfirm(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	// Point the destination at an existing local file.
	dir := t.TempDir()
	target := dir + "/existing.tar"
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.Form.Get(fieldSourcePath).Input.Set("/etc/app.conf")
	m.Form.Get(fieldDestinationPath).Input.Set(target)
	m.Form.FieldFocus = m.Form.ConfirmSlot()

	updated, cmd := submitContainerForm(m)
	if cmd != nil {
		t.Fatalf("overwrite submit must not run a command directly: %T", cmd)
	}
	// A unified confirm box opens with default Cancel focus and a Force option.
	if updated.Navigation.Mode != state.ModeConfirm {
		t.Fatalf("mode = %v, want ModeConfirm", updated.Navigation.Mode)
	}
	if updated.Confirm.Focus != 0 || updated.Confirm.Options[0].ID != keys.ShowOptionCancel {
		t.Fatalf("overwrite confirm must default to Cancel: %#v", updated.Confirm.Options)
	}
	hasForce := false
	for _, opt := range updated.Confirm.Options {
		if opt.ID == keys.ShowOptionForce {
			hasForce = true
		}
	}
	if !hasForce {
		t.Fatal("overwrite confirm must offer a Force option")
	}
}

func TestSubmitContainerCopyOverwriteCancelKeepsForm(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	dir := t.TempDir()
	target := dir + "/existing.tar"
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.Form.Get(fieldSourcePath).Input.Set("/etc/app.conf")
	m.Form.Get(fieldDestinationPath).Input.Set(target)
	m.Form.FieldFocus = m.Form.ConfirmSlot()
	updated, _ := submitContainerForm(m)

	// Cancel returns to the still-open form.
	updated, cmd := handleConfirmKeys(keys.KeyEnter, updated)
	if cmd != nil {
		t.Fatalf("cancel returned a cmd %T, want nil", cmd)
	}
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Form.Kind != state.FormContainerCopy {
		t.Fatalf("cancel must return to the open form: mode=%v kind=%v", updated.Navigation.Mode, updated.Form.Kind)
	}
}

func TestSubmitContainerExportOverwriteForceRuns(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dir := t.TempDir()
	target := dir + "/existing.tar"
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := m.Form.Get(fieldDestinationPath)
	dst.Input.Set(target)
	m.Form.FieldFocus = m.Form.ConfirmSlot()
	updated, _ := submitContainerForm(m)
	if updated.Navigation.Mode != state.ModeConfirm {
		t.Fatalf("expected overwrite confirm, mode=%v", updated.Navigation.Mode)
	}
	// Move focus to the Force option and confirm.
	updated, _ = handleConfirmKeys(keys.KeyTab, updated)
	updated, cmd := handleConfirmKeys(keys.KeyEnter, updated)
	if cmd == nil {
		t.Fatal("force-confirmed export must run the command")
	}
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("after force confirm mode = %v, want ModeNormal", updated.Navigation.Mode)
	}
	if _, ok := cmd().(ContainerExportDone); !ok {
		t.Fatalf("force confirm returned %T, want ContainerExportDone", cmd)
	}
}

func TestFormTabDoesNotMoveFocusOnNonPath(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCommitForm(m)
	// Default focus is Cancel; from there Down reaches the last field (5).
	m.Form.MoveField(1)
	if m.Form.FieldFocus != 0 {
		t.Fatalf("setup focus = %d, want 0", m.Form.FieldFocus)
	}
	start := m.Form.FieldFocus
	handleContainerFormKey(keys.KeyTab, m)
	if m.Form.FieldFocus != start {
		t.Fatalf("Tab on Text must not move focus: %d -> %d", start, m.Form.FieldFocus)
	}
	handleContainerFormKey(keys.KeyShiftTab, m)
	if m.Form.FieldFocus != start {
		t.Fatalf("Shift+Tab on Text must not move focus: %d", m.Form.FieldFocus)
	}
	// Bool field: Tab must also not move focus.
	openContainerCommitForm(m)
	m.Form.MoveField(1) // 0
	m.Form.MoveField(1) // 1
	m.Form.MoveField(1) // 2
	m.Form.MoveField(1) // 3
	m.Form.MoveField(1) // 4 (pause bool)
	start = m.Form.FieldFocus
	if m.Form.Field() == nil || m.Form.Field().Kind != state.FormBool {
		t.Fatalf("setup focus = %#v, want bool", m.Form.Field())
	}
	handleContainerFormKey(keys.KeyTab, m)
	if m.Form.FieldFocus != start {
		t.Fatalf("Tab on Bool must not move focus: %d -> %d", start, m.Form.FieldFocus)
	}
}

func TestFormPathTabNoCandidateKeepsFocusAndShowsToast(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	m.Form.MoveField(1)
	// A path that matches no files in an empty temp dir.
	dst.Input.Set(t.TempDir() + "/definitely-no-such-prefix-xyz/")
	before := m.Form.FieldFocus
	updated, _ := handleContainerFormKey(keys.KeyTab, m)
	if updated.Form.FieldFocus != before {
		t.Fatalf("Tab with no candidate must not move focus: %d -> %d", before, updated.Form.FieldFocus)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("Tab with no candidate must surface a non-blocking toast")
	}
}

func TestPathPopupForwardAndReverseCycle(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	m.Form.MoveField(1)
	dir := t.TempDir()
	for _, name := range []string{"alpha.tar", "alpine.tar", "argon.tar"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dst.Input.Set(filepath.Join(dir, "al"))
	handleContainerFormKey(keys.KeyTab, m)
	handleContainerFormKey(keys.KeyTab, m)
	if !m.Form.Popup.Open {
		t.Fatal("Tab must open popup")
	}
	start := m.Form.Popup.Cursor
	handleContainerFormKey(keys.KeyTab, m)
	if m.Form.Popup.Cursor != (start+1)%3 {
		t.Fatalf("Tab in popup must advance cursor: start=%d got=%d", start, m.Form.Popup.Cursor)
	}
	handleContainerFormKey(keys.KeyShiftTab, m)
	if m.Form.Popup.Cursor != start {
		t.Fatalf("Shift+Tab in popup must reverse cursor: want %d got %d", start, m.Form.Popup.Cursor)
	}
}

func TestPathPopupLeftRightNavigatesDirectories(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	m.Form.MoveField(1)
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "beta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "alpha", "inside.tar"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst.Input.Set(dir + string(os.PathSeparator))
	handleContainerFormKey(keys.KeyTab, m)
	handleContainerFormKey(keys.KeyTab, m)
	if !m.Form.Popup.Open {
		t.Fatal("second Tab must open candidates")
	}
	handleContainerFormKey(keys.KeyRight, m)
	if !m.Form.Popup.Open || !strings.HasSuffix(dst.Text(), "alpha"+string(os.PathSeparator)) {
		t.Fatalf("Right must enter directory: input=%q popup=%#v", dst.Text(), m.Form.Popup)
	}
	handleContainerFormKey(keys.KeyLeft, m)
	if !m.Form.Popup.Open || dst.Text() != dir+string(os.PathSeparator) {
		t.Fatalf("Left must return to parent: input=%q popup=%#v", dst.Text(), m.Form.Popup)
	}
}

func TestFormTextCursorLeftRightHomeEnd(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	m.Form.MoveField(1) // 0 (memory)
	mem := m.Form.Get(fieldMemory)
	mem.Input.Set("256")
	mem.Input.Cursor = 3
	handleContainerFormKey(keys.KeyLeft, m)
	if mem.Input.Cursor != 2 {
		t.Fatalf("Left cursor = %d, want 2", mem.Input.Cursor)
	}
	handleContainerFormKey(keys.KeyRight, m)
	if mem.Input.Cursor != 3 {
		t.Fatalf("Right cursor = %d, want 3", mem.Input.Cursor)
	}
	handleContainerFormKey(keys.KeyHome, m)
	if mem.Input.Cursor != 0 {
		t.Fatalf("Home cursor = %d, want 0", mem.Input.Cursor)
	}
	handleContainerFormKey(keys.KeyEnd, m)
	if mem.Input.Cursor != len([]rune("256")) {
		t.Fatalf("End cursor = %d, want 3", mem.Input.Cursor)
	}
}

func TestFormSelectPopupHomeEndPgUpPgDn(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	m.Form.MoveField(3)
	if f := m.Form.Field(); f == nil || f.Key != fieldRestartPolicy {
		t.Fatalf("focus = %#v, want restart select", m.Form.Field())
	}
	// Open popup; restart has 5 options.
	handleContainerFormKey(keys.KeyEnter, m)
	if !m.Form.Popup.Open {
		t.Fatal("Enter must open select popup")
	}
	handleContainerFormKey(keys.KeyEnd, m)
	if m.Form.Popup.Cursor != 4 {
		t.Fatalf("End cursor = %d, want 4", m.Form.Popup.Cursor)
	}
	handleContainerFormKey(keys.KeyHome, m)
	if m.Form.Popup.Cursor != 0 {
		t.Fatalf("Home cursor = %d, want 0", m.Form.Popup.Cursor)
	}
	handleContainerFormKey(keys.KeyPgDn, m)
	if m.Form.Popup.Cursor != 4 {
		t.Fatalf("PgDn from 0 (visible=8) clamps to last: %d, want 4", m.Form.Popup.Cursor)
	}
	handleContainerFormKey(keys.KeyHome, m)
	handleContainerFormKey(keys.KeyPgUp, m)
	if m.Form.Popup.Cursor != 0 {
		t.Fatalf("PgUp clamps to 0: got %d", m.Form.Popup.Cursor)
	}
	// Space commits select and closes popup.
	handleContainerFormKey(keys.KeyDown, m) // 1
	handleContainerFormKey(keys.KeySpace, m)
	if m.Form.Popup.Open {
		t.Fatal("Space on Select must close popup")
	}
	if f := m.Form.Get(fieldRestartPolicy); f.Option() != f.Options[1] {
		t.Fatalf("selected option = %q, want %q", f.Option(), f.Options[1])
	}
}

func TestTabDoesNotMoveSelectPopup(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	m.Form.FieldFocus = 2
	handleContainerFormKey(keys.KeyEnter, m)
	start := m.Form.Popup.Cursor
	handleContainerFormKey(keys.KeyTab, m)
	handleContainerFormKey(keys.KeyShiftTab, m)
	if m.Form.Popup.Cursor != start {
		t.Fatalf("Tab changed select popup cursor: %d -> %d", start, m.Form.Popup.Cursor)
	}
}

func TestFormSelectEscCancelsPopup(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	m.Form.MoveField(3)
	startIdx := m.Form.Get(fieldRestartPolicy).Index
	handleContainerFormKey(keys.KeyEnter, m)
	handleContainerFormKey(keys.KeyDown, m)
	handleContainerFormKey(keys.KeyDown, m)
	handleContainerFormKey(keys.KeyEsc, m)
	if m.Form.Popup.Open {
		t.Fatal("Esc must close select popup")
	}
	if m.Form.Get(fieldRestartPolicy).Index != startIdx {
		t.Fatalf("Esc must restore original index: %d -> %d", startIdx, m.Form.Get(fieldRestartPolicy).Index)
	}
}

func TestPathPopupEnterOnDirectoryVsFile(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerExportForm(m)
	dst := m.Form.Get(fieldDestinationPath)
	m.Form.MoveField(1)
	dir := t.TempDir()
	// Create a directory with a file inside, plus a file at the top level.
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", "inside.tar"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "backup.tar"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst.Input.Set(dir + "/")
	handleContainerFormKey(keys.KeyTab, m)
	handleContainerFormKey(keys.KeyTab, m)
	if !m.Form.Popup.Open {
		t.Fatal("Tab must open popup")
	}
	// Enter confirms the selected directory instead of navigating into it.
	handleContainerFormKey(keys.KeyEnter, m)
	if m.Form.Popup.Open {
		t.Fatal("Enter on a directory must confirm and close the popup")
	}
	if !strings.HasSuffix(dst.Input.Text, string(os.PathSeparator)) {
		t.Fatalf("Enter on a directory must append separator: %q", dst.Input.Text)
	}
}

func TestFormLeftRightInButtonsToggles(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerCopyForm(m)
	if m.Form.FocusedButton() != keys.ShowOptionCancel {
		t.Fatalf("default focus = %q, want cancel", m.Form.FocusedButton())
	}
	// Up/Left on Cancel → Confirm (BR-043 §3.3 linear model).
	handleContainerFormKey(keys.KeyLeft, m)
	if m.Form.FocusedButton() != keys.ShowOptionConfirm {
		t.Fatalf("Left on Cancel = %q, want confirm", m.Form.FocusedButton())
	}
	// Down/Right on Confirm → Cancel.
	handleContainerFormKey(keys.KeyRight, m)
	if m.Form.FocusedButton() != keys.ShowOptionCancel {
		t.Fatalf("Right on Confirm = %q, want cancel", m.Form.FocusedButton())
	}
	// From a field: Left/Right must NOT toggle button focus.
	m.Form.FieldFocus = 0 // land on first field directly
	handleContainerFormKey(keys.KeyLeft, m)
	if m.Form.FocusedButton() != "" {
		t.Fatalf("Left on a field must not enter button area: %q", m.Form.FocusedButton())
	}
}

// TestRestartPolicyChoicesStructure asserts Phase 3's Key+Label split: the
// choices slice must carry the five canonical Docker/Podman policies, every
// entry must have a non-empty Key and Label, and the "unchanged" lead must
// be preserved (BR-041 §4.2 — no policy chosen = current policy unchanged).
// The compile-time assertion guards the declared type: if a future refactor
// regresses restartPolicyChoices back to []string, this assertion fails to
// build and the test never silently passes against the wrong type.
func TestRestartPolicyChoicesStructure(t *testing.T) {
	var _ []restartChoice = restartPolicyChoices
	if got := len(restartPolicyChoices); got != 5 {
		t.Fatalf("restartPolicyChoices length = %d, want 5", got)
	}
	wantKeys := []string{"unchanged", "no", "always", "unless-stopped", "on-failure"}
	for i, want := range wantKeys {
		c := restartPolicyChoices[i]
		if c.Key != want {
			t.Errorf("restartPolicyChoices[%d].Key = %q, want %q", i, c.Key, want)
		}
		if c.Label == "" {
			t.Errorf("restartPolicyChoices[%d].Label is empty", i)
		}
	}
}

// TestRestartPolicyChoicesLabelsAreTranslated makes sure every label comes
// from i18n.T and is not the raw key value. i18n.T returns the key string
// verbatim when the translation is missing, so this guards against an
// accidentally-untranslated key regressing into a raw key on the UI.
//
// The test pins the language to English so the assertion is deterministic
// regardless of which test ran before (pinned lang is restored on exit).
func TestRestartPolicyChoicesLabelsAreTranslated(t *testing.T) {
	prev := i18n.Current()
	i18n.SetLang(i18n.LanguageEnglish)
	t.Cleanup(func() { i18n.SetLang(prev) })

	for _, c := range restartPolicyChoices {
		if c.Label == "" {
			t.Errorf("label for %q is empty", c.Key)
		}
		if c.Label == c.Key {
			t.Errorf("label for %q is raw key (translation missing)", c.Key)
		}
	}
}

// TestBuildRestartFieldHasDisplayOptions asserts the helper builds a Select
// field with Options (keys) and DisplayOptions (labels) of equal length,
// preserving the canonical order. The label for the third entry
// (Index=2) must be the DisplayOptions label, not the raw key.
func TestBuildRestartFieldHasDisplayOptions(t *testing.T) {
	field := buildRestartField()
	if field.Kind != state.FormSelect {
		t.Fatalf("Kind = %v, want FormSelect", field.Kind)
	}
	if len(field.Options) != 5 {
		t.Fatalf("Options length = %d, want 5", len(field.Options))
	}
	if len(field.DisplayOptions) != len(field.Options) {
		t.Fatalf("DisplayOptions length = %d, Options length = %d (must match)", len(field.DisplayOptions), len(field.Options))
	}
	for i, c := range restartPolicyChoices {
		if field.Options[i] != c.Key {
			t.Errorf("Options[%d] = %q, want %q", i, field.Options[i], c.Key)
		}
		if field.DisplayOptions[i] != c.Label {
			t.Errorf("DisplayOptions[%d] = %q, want %q", i, field.DisplayOptions[i], c.Label)
		}
	}
	if field.Option() != "unchanged" {
		t.Errorf("default Option() = %q, want unchanged", field.Option())
	}
}

// TestRestartPolicyFieldIndexRoundtrip exercises the Wire/Display split
// end-to-end: opening the Update form and selecting the third option
// (Index=2 = "always") must surface the localized label, not the raw key.
func TestRestartPolicyFieldIndexRoundtrip(t *testing.T) {
	m := formModel(t, &stubContainerService{})
	openContainerUpdateForm(m)
	restart := m.Form.Get(fieldRestartPolicy)
	if restart == nil {
		t.Fatal("restart field must be present in Update form")
	}
	restart.Index = 2
	if got := restart.Option(); got != "always" {
		t.Fatalf("Option() = %q, want always (Indexer reads Wire key)", got)
	}
	displayAt := restart.DisplayOptions[restart.Index]
	if displayAt == "" || displayAt == restart.Option() {
		t.Fatalf("DisplayOptions[2] = %q, want a non-empty, non-raw-key label", displayAt)
	}
	if !strings.Contains(displayAt, strings.TrimSpace(restart.DisplayOptions[2])) {
		t.Fatalf("display label %q must match DisplayOptions[2]", displayAt)
	}
}
