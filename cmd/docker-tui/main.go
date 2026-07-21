package main

import (
	"fmt"
	"net/url"
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

	pool := dockerclient.NewPool()
	connections, initialConnection := runtimeConnections(cfg, dockerHost, podmanMode)
	for _, connection := range connections {
		pool.AddHost(connection)
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
	m.RuntimeSelectorDisabled = dockerHost != ""
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
		connectDocker(m.model.Pool, m.initialConnection),
	)
}

func connectDocker(pool *dockerclient.ConnectionPool, name string) tea.Cmd {
	return func() tea.Msg {
		candidates := []string{name}
		if name == "local-docker" {
			candidates = append(candidates, "local-podman")
		}
		var errors []string
		for _, candidate := range candidates {
			if pool.Get(candidate) == nil {
				continue
			}
			if err := pool.Connect(candidate, 2*time.Second); err == nil {
				notice := ""
				if candidate != name {
					notice = fmt.Sprintf("%s unavailable; connected to %s", name, candidate)
				}
				return state.DockerConnected{Client: pool.ActiveClient(), Name: candidate, Notice: notice}
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

func runtimeConnections(cfg *config.Config, hostOverride string, usePodman bool) ([]dockerclient.HostEntry, string) {
	if hostOverride != "" {
		driver := "docker"
		if usePodman {
			driver = "podman"
		}
		return []dockerclient.HostEntry{{Name: "cli", Host: hostOverride, Runtime: driver}}, "cli"
	}

	connections := make([]dockerclient.HostEntry, 0, len(cfg.Runtime.Connections)+2)
	seen := make(map[string]int, len(cfg.Runtime.Connections)+2)
	aliases := make(map[string]string, len(cfg.Runtime.Connections)+2)
	add := func(entry dockerclient.HostEntry, replace bool) {
		key := connectionKey(entry.Runtime, entry.Host)
		if index, exists := seen[key]; exists {
			if replace {
				aliases[connections[index].Name] = entry.Name
				connections[index] = entry
			}
			return
		}
		seen[key] = len(connections)
		connections = append(connections, entry)
	}
	// Local runtimes are always candidates; discovery settings do not hide an
	// installed runtime from the selector. Unavailable sockets fail visibly.
	add(dockerclient.HostEntry{Name: "local-docker", Host: "unix:///var/run/docker.sock", Runtime: "docker"}, false)
	add(dockerclient.HostEntry{Name: "local-podman", Host: localPodmanEndpoint(), Runtime: "podman"}, false)
	for _, connection := range cfg.Runtime.Connections {
		add(dockerclient.HostEntry{
			Name:    connection.Name,
			Host:    connection.Endpoint,
			Runtime: connection.Driver,
			TLS: dockerclient.TLSConfig{
				Enabled:  connection.TLS.Enabled,
				Verify:   connection.TLS.Verify,
				CAFile:   connection.TLS.CAFile,
				CertFile: connection.TLS.CertFile,
				KeyFile:  connection.TLS.KeyFile,
			},
		}, true)
	}
	initial := cfg.Runtime.Default
	if usePodman {
		initial = "local-podman"
	}
	if canonical, exists := aliases[initial]; exists {
		initial = canonical
	}
	return connections, initial
}

func connectionKey(driver, endpoint string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	endpoint = strings.TrimSpace(endpoint)
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" {
		return driver + "|" + endpoint
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	if parsed.Scheme == "unix" {
		parsed.Path = filepath.Clean(parsed.Path)
	}
	return driver + "|" + parsed.String()
}

func localPodmanEndpoint() string {
	if os.Getuid() == 0 {
		return "unix:///run/podman/podman.sock"
	}
	return fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", os.Getuid())
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
