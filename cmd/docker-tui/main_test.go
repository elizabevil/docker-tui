package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
)

func TestRuntimeConnectionsIncludeLocalAndConfiguredEntries(t *testing.T) {
	cfg := config.DefaultAppConfig()
	cfg.Runtime.Default = "remote"
	cfg.Runtime.Connections = []config.RuntimeConnection{{Name: "remote", Driver: "docker", Endpoint: "tcp://example:2376"}}

	entries, initial := runtime.BuildConnections(&cfg.Runtime)
	if initial != "remote" || len(entries) != 3 {
		t.Fatalf("initial=%q entries=%#v", initial, entries)
	}
	if entries[0].Name != "local-docker" || entries[1].Name != "local-podman" || entries[2].Name != "remote" {
		t.Fatalf("entry order=%#v", entries)
	}
}

func TestRuntimeConnectionsDeduplicateConfiguredLocalEndpoint(t *testing.T) {
	cfg := config.DefaultAppConfig()
	cfg.Runtime.Default = "configured-docker"
	cfg.Runtime.Connections = []config.RuntimeConnection{{Name: "configured-docker", Driver: "docker", Endpoint: "unix:///var/run/docker.sock"}}

	entries, initial := runtime.BuildConnections(&cfg.Runtime)
	if len(entries) != 2 || entries[0].Name != "configured-docker" || initial != "configured-docker" {
		t.Fatalf("initial=%q entries=%#v", initial, entries)
	}
}

func TestRuntimeConnectionsHostOverrideIsExclusive(t *testing.T) {
	cfg := config.DefaultAppConfig()
	entries, initial := runtime.BuildConnections(&cfg.Runtime, runtime.WithHostOverride("tcp://example:2375", true))
	if initial != "cli" || len(entries) != 1 || entries[0].Runtime != "podman" {
		t.Fatalf("initial=%q entries=%#v", initial, entries)
	}
}

func TestPodmanOverrideAddsLocalEntryWhenDiscoveryDisabled(t *testing.T) {
	cfg := config.DefaultAppConfig()
	cfg.Runtime.Discovery.LocalPodman = false
	entries, initial := runtime.BuildConnections(&cfg.Runtime, runtime.WithHostOverride("", true))
	if initial != "local-podman" {
		t.Fatalf("initial=%q", initial)
	}
	found := false
	for _, entry := range entries {
		found = found || entry.Name == "local-podman"
	}
	if !found {
		t.Fatalf("entries=%#v", entries)
	}
}

func TestRuntimeConnectionsAlwaysIncludeLocalRuntimes(t *testing.T) {
	cfg := config.DefaultAppConfig()
	cfg.Runtime.Discovery.LocalDocker = false
	cfg.Runtime.Discovery.LocalPodman = false
	entries, _ := runtime.BuildConnections(&cfg.Runtime)
	if len(entries) < 2 || entries[0].Name != "local-docker" || entries[1].Name != "local-podman" {
		t.Fatalf("entries=%#v", entries)
	}
}

func TestBuildAppRegistersInfoCommand(t *testing.T) {
	app := buildApp()
	cmds := app.GetCommands()
	info, ok := cmds["info"]
	if !ok {
		t.Fatalf("info command not registered; got %#v", cmds)
	}
	if info.HasSubcommands() {
		t.Fatalf("info command must be leaf; got subcommands")
	}
}

func TestBuildAppRegistersConfigCommandGroup(t *testing.T) {
	app := buildApp()
	cmds := app.GetCommands()
	cfgCmd, ok := cmds["config"]
	if !ok {
		t.Fatalf("config command not registered; got %#v", cmds)
	}
	subs := cfgCmd.GetSubcommands()
	if subs["init"] == nil {
		t.Fatalf("config init not registered; got %#v", subs)
	}
	if subs["validate"] == nil {
		t.Fatalf("config validate not registered; got %#v", subs)
	}
}

func TestBuildAppRemovesListThemesFlag(t *testing.T) {
	app := buildApp()
	help := app.GenerateHelp()
	if strings.Contains(help, "list-themes") {
		t.Fatalf("--list-themes still present in help:\n%s", help)
	}
}

func TestFirstRunHintShownWhenConfigMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	hint := firstRunHint()
	if hint == "" {
		t.Fatalf("firstRunHint should be non-empty when config file is missing")
	}
	if !strings.Contains(hint, "config init") {
		t.Fatalf("hint %q missing config init reference", hint)
	}
}

func TestFirstRunHintSuppressedWhenConfigExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfgFile, err := config.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfgFile), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(cfgFile, []byte("version: 1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if hint := firstRunHint(); hint != "" {
		t.Fatalf("firstRunHint = %q, want empty when config file exists", hint)
	}
}

func TestCursorParkingWriterAppendsParkSequence(t *testing.T) {
	w, read := newTestParkingWriter(t)
	w.SetTarget(79, 21, true)

	if _, err := w.Write([]byte("frame")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	want := "frame\x1b[22;80H\x1b[?25l"
	if got := read(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestCursorParkingWriterZeroBasedToOneBased(t *testing.T) {
	w, read := newTestParkingWriter(t)
	w.SetTarget(0, 0, true)

	if _, err := w.Write([]byte("x")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	want := "x\x1b[1;1H\x1b[?25l"
	if got := read(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestCursorParkingWriterIdempotentAcrossWrites(t *testing.T) {
	w, read := newTestParkingWriter(t)
	w.SetTarget(10, 5, true)

	if _, err := w.Write([]byte("a")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := w.Write([]byte("b")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	want := "a\x1b[6;11H\x1b[?25lb\x1b[6;11H\x1b[?25l"
	if got := read(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestCursorParkingWriterInactiveNoSequence(t *testing.T) {
	w, read := newTestParkingWriter(t)
	w.SetTarget(10, 5, false)

	if _, err := w.Write([]byte("frame")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := read(); got != "frame" {
		t.Fatalf("output = %q, want %q", got, "frame")
	}
}

func TestCursorParkingWriterDefaultInactive(t *testing.T) {
	w, read := newTestParkingWriter(t)

	if _, err := w.Write([]byte("frame")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := read(); got != "frame" {
		t.Fatalf("output = %q, want %q", got, "frame")
	}
}

func TestCursorParkingWriterForwardsWriteError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "parking-*.out")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	f.Close() // write on a closed file must surface an error
	w := newCursorParkingWriter(f)
	if _, err := w.Write([]byte("frame")); err == nil {
		t.Fatalf("Write: want error, got nil")
	}
}

func newTestParkingWriter(t *testing.T) (*cursorParkingWriter, func() string) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "parking-*.out")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	read := func() string {
		t.Helper()
		b, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		return string(b)
	}
	return newCursorParkingWriter(f), read
}

func TestViewParksCursorOnFooterWhenIdle(t *testing.T) {
	m := newMainModelForView(t, 120, 32)
	v := m.View()
	if v.Cursor != nil {
		t.Fatalf("View.Cursor = %#v, want nil (visibility must stay hidden)", v.Cursor)
	}
	rep := view.ResolveLayout(m.model)
	target := m.parking.target.Load().(cursorParkTarget)
	if !target.active || target.x != 119 || target.y != rep.FooterTop {
		t.Fatalf("target = %#v, want active (119, %d)", target, rep.FooterTop)
	}
}

func TestViewDoesNotParkWhileTextInputActive(t *testing.T) {
	m := newMainModelForView(t, 120, 32)
	m.model.Navigation.Mode = state.ModeFilter
	v := m.View()
	if v.Cursor != nil {
		t.Fatalf("View.Cursor = %#v, want nil", v.Cursor)
	}
	target := m.parking.target.Load().(cursorParkTarget)
	if target.active {
		t.Fatalf("target = %#v, want inactive while text input owns the keyboard", target)
	}
}

func newMainModelForView(t *testing.T, width, height int) *mainModel {
	t.Helper()
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = width
	app.Viewport.Height = height
	f, err := os.CreateTemp(t.TempDir(), "parking-*.out")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	return &mainModel{model: app, parking: newCursorParkingWriter(f)}
}
