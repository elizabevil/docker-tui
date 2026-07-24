package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/runtimeinit"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/app"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var (
	cfgFile     string
	dockerHost  string
	themeName   string
	listThemes  bool
	langFlag    string
	showVersion bool
	podmanMode  bool
	version     = "0.2.0"
)

var rootCmd = &cobra.Command{
	Use:   "dtui",
	Short: "dtui - Docker & Podman TUI Manager",
	Long: `dtui is a terminal user interface for managing Docker and Podman
containers, images, volumes, networks, and Compose projects.

Inspired by k9s for Kubernetes, it provides a keyboard-driven,
real-time view of your container environment.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if showVersion {
			fmt.Printf("dtui version %s\n", version)
			return nil
		}
		if listThemes {
			for _, t := range config.ListThemes() {
				fmt.Println(t)
			}
			return nil
		}
		return runTUI()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "f", "", "config file path")
	rootCmd.Flags().StringVarP(&dockerHost, "host", "H", "", "daemon socket or host")
	rootCmd.Flags().StringVarP(&themeName, "theme", "t", "", "theme name (default, dark, light, nord, dracula, solarized)")
	rootCmd.Flags().BoolVarP(&podmanMode, "podman", "p", false, "use Podman socket")
	rootCmd.Flags().BoolVar(&listThemes, "list-themes", false, "list available themes")
	rootCmd.Flags().StringVarP(&langFlag, "lang", "L", "", "language (zh/en)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}
	lang := langFlag
	if lang == "" {
		lang = cfg.General.Lang
	}
	i18n.Init(lang)

	// Configure size formatting (default 1024-base binary, optional 1000-base SI).
	utils.SetSizeFormat(cfg.General.SizeFormat == "si")

	// Load theme.
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
		m.Dependencies.Audit = audit.NewService(audit.NewFileSink(filepath.Join(dir, "logs")))
	} else {
		m.Dependencies.Audit = audit.NewService(nil)
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

// probeAllOnStart runs an initial probe of every known host so the runtime
// selector displays latency, version and status on first launch. The
// selected host is probed first so the active connection is measured
// immediately.
func probeAllOnStart(pool *runtimeapi.ConnectionPool, active string) tea.Cmd {
	return func() tea.Msg {
		results := pool.RefreshAll(2 * time.Second)
		cmds := make([]tea.Cmd, 0, len(results)+1)
		// Order non-active results first, then the active one so the UI
		// measures the selected connection immediately on first paint.
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

		// Try all connections sequentially; first success wins.
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
	updatedModel, cmd := tui.Update(msg, m.model)
	m.model = updatedModel
	return m, cmd
}

func (m *mainModel) View() tea.View {
	v := tea.NewView(view.RenderApp(m.model))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
