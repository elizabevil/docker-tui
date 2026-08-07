package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
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
