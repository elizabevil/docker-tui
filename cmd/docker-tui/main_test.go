package main

import (
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
