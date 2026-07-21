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
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/app"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
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
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		cfg = config.DefaultConfig()
	}
	if dockerHost != "" {
		cfg.Docker.Host = dockerHost
	}
	if podmanMode {
		cfg.Docker.Host = "unix:///run/podman/podman.sock"
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

	pool := dockerclient.NewPool()
	rt := strings.ToLower(cfg.General.Runtime)

	// Collect runtime connections
	type connPair struct{ name, addr string }
	var hosts []connPair
	for _, c := range cfg.Runtime.Connections {
		hosts = append(hosts, connPair{c.Name, c.Addr})
	}
	if len(hosts) == 0 {
		hosts = []connPair{
			{"local-docker", "unix:///var/run/docker.sock"},
			{"local-podman", "unix:///run/user/1000/podman/podman.sock"},
		}
	}
	for _, h := range hosts {
		pool.AddHost(dockerclient.HostEntry{Name: h.name, Host: h.addr, Runtime: rt})
	}

	m := state.NewAppModel(cfg, nil, version)
	if dir, configErr := config.ConfigDir(); configErr == nil {
		m.Audit = audit.NewService(audit.NewFileSink(filepath.Join(dir, "logs")))
	} else {
		m.Audit = audit.NewService(nil)
	}
	m.Pool = pool
	m.Theme = theme
	m.HeaderVisible = true
	m.Connecting = true
	p := tea.NewProgram(&mainModel{model: m})
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	pool.Close()
	return nil
}

type mainModel struct {
	model *state.AppModel
}

func (m *mainModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return state.HostStatsTick{} },
		func() tea.Msg { return state.ToastTick{} },
		connectDocker(m.model.Pool),
	)
}

func connectDocker(pool *dockerclient.ConnectionPool) tea.Cmd {
	return func() tea.Msg {
		for _, name := range []string{"local-podman", "local-docker"} {
			entry := pool.Get(name)
			if entry == nil {
				continue
			}
			if err := pool.Connect(name, 2*time.Second); err == nil {
				return state.DockerConnected{Client: pool.ActiveClient(), Name: name}
			}
		}
		return state.DockerConnected{Error: fmt.Errorf("no container engine available")}
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
