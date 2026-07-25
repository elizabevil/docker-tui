package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/agilira/orpheus/pkg/orpheus"
	"github.com/bytedance/sonic"
	"github.com/elizabevil/docker-tui/internal/buildinfo"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/runtimeinit"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
	"github.com/elizabevil/docker-tui/internal/tui/update"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var version = "0.2.0"

func main() {
	read := buildinfo.Read("dtui", version)
	marshal, _ := sonic.MarshalIndent(read, " ", " ") //nolint:errcheck // version info struct; marshal cannot fail in practice.
	app := orpheus.New("dtui").
		SetDescription("Docker & Podman TUI Manager").
		SetVersion(string(marshal))

	app.AddGlobalFlag("config", "f", "", "Config file path")
	app.AddGlobalFlag("host", "H", "", "Daemon socket or host")
	app.AddGlobalFlag("theme", "t", "", "Theme name (default, dark, light, nord, dracula, solarized)")
	app.AddGlobalBoolFlag("podman", "p", false, "Use Podman socket")
	app.AddGlobalBoolFlag("list-themes", "", false, "List available themes")
	app.AddGlobalFlag("lang", "L", "", "Language (zh/en)")

	cmd := orpheus.NewCommand("run", "Run the TUI")
	cmd.SetHandler(func(ctx *orpheus.Context) error {
		if ctx.GetGlobalFlagBool("list-themes") {
			for _, t := range config.ListThemes() {
				fmt.Println(t)
			}
			return nil
		}
		return runTUI(ctx)
	})
	app.AddCommand(cmd)
	app.SetDefaultCommand("run")

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(ctx *orpheus.Context) error {
	cfgFile := ctx.GetGlobalFlagString("config")
	dockerHost := ctx.GetGlobalFlagString("host")
	themeName := ctx.GetGlobalFlagString("theme")
	podmanMode := ctx.GetGlobalFlagBool("podman")
	langFlag := ctx.GetGlobalFlagString("lang")

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}
	lang := langFlag
	if lang == "" {
		lang = cfg.General.Lang
	}
	i18n.SetLang(lang)
	utils.SetSizeFormat(cfg.General.SizeFormat)

	if themeName == "" {
		themeName = cfg.Theme
	}
	theme, err := config.LoadTheme(themeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v (using default)\n", err)
		theme = config.DefaultTheme()
	}
	tui.ApplyTheme(theme)
	tui.ApplyLayoutConfig(&cfg.Layout)

	pool := runtimeapi.NewPool(runtimeinit.NewEngineFactory())
	connections, initialConnection := runtimeapi.BuildConnections(&cfg.Runtime,
		runtimeapi.WithHostOverride(dockerHost, podmanMode),
	)
	for _, connection := range connections {
		pool.AddHost(connection)
	}

	m := state.NewAppModel(cfg, nil, version)
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
				ID:   "session-" + version,
				Name: "dtui " + version,
				Meta: audit.SessionMeta{Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH},
			},
			"dtui session started",
		)
	}
	m.Connection.Pool = pool
	m.Dependencies.Theme = theme
	m.Viewport.HeaderVisible = true
	m.Connection.Connecting = true
	m.Connection.RuntimeSelectorDisabled = dockerHost != ""
	p := tea.NewProgram(&mainModel{model: m, initialConnection: initialConnection})
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	pool.Close()
	return nil
}

type mainModel struct {
	model             *state.AppModel
	initialConnection string
}

func (m *mainModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return state.HostStatsTick{} },
		func() tea.Msg { return state.ToastTick{} },
		func() tea.Msg { return state.RuntimeHealthTick{} },
		connectDocker(m.model.Connection.Pool, m.initialConnection),
		probeAllOnStart(m.model.Connection.Pool, m.initialConnection),
	)
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
	updatedModel, cmd := update.Update(msg, m.model)
	m.model = updatedModel
	return m, cmd
}

func (m *mainModel) View() tea.View {
	v := tea.NewView(view.RenderApp(m.model))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
