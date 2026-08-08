package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/agilira/orpheus/pkg/orpheus"
	"github.com/bytedance/sonic"
	"github.com/charmbracelet/x/term"
	"github.com/elizabevil/docker-tui/internal/buildinfo"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/runtimeinit"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
	"github.com/elizabevil/docker-tui/internal/tui/update"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var dtuiInfo = buildinfo.Read("dtui")

// firstRunHint returns the localized hint message when no config file
// exists yet, or an empty string when one does. It never creates files
// or directories.
func firstRunHint() string {
	path, err := config.ConfigFile()
	if err != nil {
		return ""
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return i18n.T("toast.firstRunHint")
	}
	return ""
}

func main() {
	app := buildApp()

	if err := app.Run(os.Args[1:]); err != nil {
		var exitErr *cliExitError
		if errors.As(err, &exitErr) {
			fmt.Fprintln(os.Stderr, exitErr.err)
			os.Exit(exitErr.code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// buildApp constructs the orpheus CLI application with all commands
// registered. Kept separate from main for testability.
func buildApp() *orpheus.App {
	marshal, _ := sonic.MarshalIndent(dtuiInfo, " ", " ") //nolint:errcheck // version info struct; marshal cannot fail in practice.
	app := orpheus.New("dtui").
		SetDescription("Docker & Podman TUI Manager").
		SetVersion(string(marshal))

	app.AddGlobalFlag("config", "f", "", "Config file path")
	app.AddGlobalFlag("host", "H", "", "Daemon socket or host")
	app.AddGlobalFlag("theme", "t", "", "Theme name (default, dark, light, nord, dracula, solarized)")
	app.AddGlobalBoolFlag("podman", "p", false, "Use Podman socket")
	app.AddGlobalFlag("lang", "L", "", "Language (zh/en)")

	cmd := orpheus.NewCommand("run", "Run the TUI")
	cmd.SetHandler(func(ctx *orpheus.Context) error {
		return runTUI(ctx)
	})
	app.AddCommand(cmd)
	app.SetDefaultCommand("run")

	app.AddCommand(newInfoCommand())
	app.AddCommand(newConfigCommandGroup())

	return app
}

func runTUI(ctx *orpheus.Context) error {
	cfgFile := ctx.GetGlobalFlagString("config")
	dockerHost := ctx.GetGlobalFlagString("host")
	themeName := ctx.GetGlobalFlagString("theme")
	podmanMode := ctx.GetGlobalFlagBool("podman")
	langFlag := ctx.GetGlobalFlagString("lang")

	resolved, err := config.LoadResolved(config.LoadOptions{
		ConfigPath: cfgFile,
		ThemeName:  config.ThemeName(themeName),
		Lang:       config.Language(langFlag),
	})
	if err != nil {
		return err
	}
	cfg := resolved.App
	theme := resolved.Theme
	i18n.SetLang(string(cfg.General.Lang))
	utils.SetSizeFormat(string(cfg.General.SizeFormat))
	tui.ApplyTheme(theme)
	tui.ApplyUIConfig(&cfg.UI)
	tui.ApplyLayoutConfig(&cfg.Layout)

	pool := runtimeapi.NewPool(runtimeinit.NewEngineFactory())
	connections, initialConnection := runtimeapi.BuildConnections(&cfg.Runtime,
		runtimeapi.WithHostOverride(dockerHost, podmanMode),
	)
	for _, connection := range connections {
		pool.AddHost(connection)
	}
	dtuInfo := buildinfo.Read("dtui")

	m := state.NewAppModel(cfg, nil, dtuInfo.Version)
	if hint := firstRunHint(); hint != "" {
		m.Feedback.ShowToast(hint, state.NotificationInfo, 30)
	}
	if dir, configErr := config.ConfigDir(); configErr == nil {
		logsDir := filepath.Join(dir, "logs")
		if sink, err := audit.NewFileSink(logsDir); err == nil {
			m.Dependencies.Audit = audit.NewService(sink)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: audit log init failed: %v\n", err)
			m.Dependencies.Audit = audit.NewService(nil)
		}
	} else {
		m.Dependencies.Audit = audit.NewService(nil)
	}
	if dir, configErr := config.ConfigDir(); configErr == nil {
		loadFilters(m, filepath.Join(dir, "filters.json"))
	}
	m.Connection.Pool = pool
	m.Dependencies.Theme = theme
	m.Viewport.HeaderVisible = true
	m.Connection.Connecting = true
	m.Connection.RuntimeSelectorDisabled = dockerHost != ""
	if m.Dependencies.Audit != nil {
		runtimeType := "docker"
		if podmanMode {
			runtimeType = "podman"
		}
		m.Dependencies.Audit.LogSession(
			"session.start",
			audit.RuntimeContext{Type: runtimeType, Name: initialConnection, Host: dockerHost},
			audit.SessionTarget{
				ID:   "session-" + dtuiInfo.Version,
				Name: "dtui " + dtuiInfo.Version,
				Meta: audit.SessionMeta{Version: dtuiInfo.Version, OS: runtime.GOOS, Arch: runtime.GOARCH},
			},
			"dtui session started",
		)
	}
	parking := newCursorParkingWriter(os.Stdout)
	p := tea.NewProgram(&mainModel{model: m, initialConnection: initialConnection, parking: parking},
		tea.WithOutput(parking))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	pool.Close()
	return nil
}

// cursorParkTarget is the terminal position (0-based) where the cursor
// should be parked after each render, plus whether parking is active.
// When active the cursor's *position* is moved to the target while its
// *visibility* stays hidden: View.Cursor is always nil, so the renderer
// emits \x1b[?25l once and never shows the cursor again. OS-level IME
// candidate windows still follow the cursor position, so parking on the
// footer keeps them at the bottom of the screen instead of over the
// table.
type cursorParkTarget struct {
	x, y   int
	active bool
}

// cursorParkingWriter wraps the TUI output writer and re-asserts the
// parked cursor position after every Write. It implements term.File
// (Read/Close/Fd forwarded to the wrapped *os.File) so bubbletea still
// detects a TTY: ttyOutput drives window size, resize events, terminal
// restore and the color profile.
type cursorParkingWriter struct {
	f      *os.File
	target atomic.Value // cursorParkTarget
}

func newCursorParkingWriter(f *os.File) *cursorParkingWriter {
	c := &cursorParkingWriter{f: f}
	c.target.Store(cursorParkTarget{})
	return c
}

// Compile-time contract: bubbletea type-asserts the output to term.File
// in initInput to detect a TTY. ttyOutput drives window size, resize
// events, terminal restore and the color profile, so breaking this
// assertion silently disables rendering at 0x0.
var _ term.File = (*cursorParkingWriter)(nil)

func (c *cursorParkingWriter) Write(p []byte) (int, error) {
	n, err := c.f.Write(p)
	if err != nil {
		return n, err
	}
	if t, ok := c.target.Load().(cursorParkTarget); ok && t.active {
		// CUP + trailing hide. The renderer may emit \x1b[?25h on
		// altscreen transitions, so the hide is asserted here too.
		seq := fmt.Sprintf("\x1b[%d;%dH\x1b[?25l", t.y+1, t.x+1)
		if _, werr := io.WriteString(c.f, seq); werr != nil {
			_ = werr // best effort: the frame already made it out
		}
	}
	return n, err
}

func (c *cursorParkingWriter) Read(p []byte) (int, error) { return c.f.Read(p) }

func (c *cursorParkingWriter) Close() error { return c.f.Close() }

func (c *cursorParkingWriter) Fd() uintptr { return c.f.Fd() }

// SetTarget updates the parked cursor target. Call it from View so the
// target reflects the current layout before the next render flush.
func (c *cursorParkingWriter) SetTarget(x, y int, active bool) {
	c.target.Store(cursorParkTarget{x: x, y: y, active: active})
}

type mainModel struct {
	model             *state.AppModel
	initialConnection string
	parking           *cursorParkingWriter
}

const initialResizeSettleDelay = 150 * time.Millisecond

func (m *mainModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.RequestWindowSize,
		func() tea.Msg { return state.HostStatsTick{} },
		func() tea.Msg { return state.RuntimeHealthTick{} },
		cursorBlinkCmd(),
		connectDocker(m.model.Connection.Pool, m.initialConnection),
		probeAllOnStart(m.model.Connection.Pool, m.initialConnection),
	}
	if m.model.Feedback.ToastTimer > 0 {
		generation := m.model.Feedback.ToastGeneration
		cmds = append(cmds, func() tea.Msg { return state.ToastTick{Generation: generation} })
	}
	return tea.Batch(cmds...)
}

func cursorBlinkCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return state.CursorBlinkTick{} })
}

func requestWindowSizeAfter(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(time.Time) tea.Msg { return tea.RequestWindowSize() })
}

func probeAllOnStart(pool *runtimeapi.ConnectionPool, active string) tea.Cmd {
	return func() tea.Msg {
		results := pool.RefreshAll(2 * time.Second)
		cmds := make([]tea.Cmd, 0, len(results)+1)
		var activeResult *runtimeapi.RefreshResult
		for i := range results {
			r := &results[i]
			if r.Name == active {
				activeResult = r
				continue
			}
			probe := r
			cmds = append(cmds, func() tea.Msg {
				return state.RuntimeProbeResult{Name: probe.Name, Error: probe.Error}
			})
		}
		if activeResult != nil {
			ar := activeResult
			cmds = append(cmds, func() tea.Msg {
				return state.RuntimeProbeResult{Name: ar.Name, Error: ar.Error}
			})
		}
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return state.ConnectionRefreshTick{}
		}))
		return tea.Batch(cmds...)
	}
}

func connectDocker(pool *runtimeapi.ConnectionPool, name string) tea.Cmd {
	return func() tea.Msg {
		names := pool.KnownHostNames()
		if len(names) == 0 {
			return state.DockerConnected{Name: name, Error: fmt.Errorf("no runtime connections configured")}
		}

		var errors []string
		for _, candidate := range names {
			if pool.Get(candidate) == nil {
				continue
			}
			if err := pool.Connect(candidate, 2*time.Second); err == nil {
				notice := ""
				if candidate != name {
					notice = fmt.Sprintf("%s unavailable; connected to %s", name, candidate)
				}
				return state.DockerConnected{Engine: pool.ActiveEngine(), Name: candidate, Notice: notice}
			} else {
				errors = append(errors, fmt.Sprintf("%s: %v", candidate, err))
			}
		}
		if len(errors) == 0 {
			return state.DockerConnected{Name: name, Error: fmt.Errorf("unknown runtime connection %q", name)}
		}
		return state.DockerConnected{Name: name, Error: fmt.Errorf("no available runtime (%s)", strings.Join(errors, "; "))}
	}
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.QuitMsg); ok {
		saveFilters(m.model)
	}
	toastGeneration := m.model.Feedback.ToastGeneration
	updatedModel, cmd := update.Update(msg, m.model)
	m.model = updatedModel
	if updatedModel.Feedback.ToastTimer > 0 && updatedModel.Feedback.ToastGeneration != toastGeneration {
		generation := updatedModel.Feedback.ToastGeneration
		toastCmd := tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
			return state.ToastTick{Generation: generation}
		})
		if cmd == nil {
			return m, toastCmd
		}
		return m, tea.Batch(cmd, toastCmd)
	}
	return m, cmd
}

// loadFilters reads <path> and applies it to m. Missing or corrupt
// files are silently ignored — the user just starts with empty
// filters in that case.
func loadFilters(m *state.AppModel, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	_ = filter.Load(m, f)
}

// saveFilters writes m's per-panel filter text to <path> in JSON.
// Errors are best-effort: a failed save just means the filters
// won't survive the next launch, but the current session is fine.
func saveFilters(m *state.AppModel) {
	dir, err := config.ConfigDir()
	if err != nil {
		return
	}
	path := filepath.Join(dir, "filters.json")
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_ = filter.New(m).Save(f)
}

func (m *mainModel) View() tea.View {
	v := tea.NewView(view.RenderApp(m.model))
	v.AltScreen = true
	// Mouse support is opt-out via config.ui.enableMouse; keyboard
	// remains the primary input path even when mouse is on.
	if m.model.Dependencies.Config != nil && m.model.Dependencies.Config.UI.EnableMouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	// When no text input owns the keyboard, park the cursor on the
	// footer rail so IME candidate windows render at the bottom of the
	// screen instead of over the table. The cursor stays hidden (View.Cursor
	// is never set); only its position is moved, by cursorParkingWriter.
	if m.parking != nil {
		if x, y, ok := view.IdleCursorPosition(m.model); ok {
			m.parking.SetTarget(x, y, true)
		} else {
			m.parking.SetTarget(0, 0, false)
		}
	}
	return v
}
