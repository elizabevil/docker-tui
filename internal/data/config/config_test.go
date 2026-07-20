package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if cfg.Docker.Host != "" {
		t.Errorf("expected empty default host, got %q", cfg.Docker.Host)
	}
	if cfg.Docker.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cfg.Docker.Timeout)
	}
	if cfg.General.ScrollHeight != 2 {
		t.Errorf("expected scrollHeight 2, got %d", cfg.General.ScrollHeight)
	}
	if cfg.Keymap.Quit == nil || len(cfg.Keymap.Quit) == 0 {
		t.Error("expected Quit keybindings to be set")
	}
	if cfg.UI.Theme.ActiveBorderColor == nil {
		t.Error("expected ActiveBorderColor to be set")
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir failed: %v", err)
	}
	if dir == "" {
		t.Fatal("ConfigDir returned empty string")
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "docker-tui")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestConfigFile(t *testing.T) {
	path, err := ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile failed: %v", err)
	}
	if filepath.Base(path) != "config.yml" {
		t.Errorf("expected config.yml, got %s", filepath.Base(path))
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yml")
	if err != nil {
		t.Fatalf("Load should not error for missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load returned nil config")
	}
	if cfg.Docker.Host != "" {
		t.Errorf("expected empty host from defaults, got %q", cfg.Docker.Host)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test-config.yml")

	cfg := DefaultConfig()
	cfg.Docker.Host = "tcp://192.168.1.1:2375"
	cfg.UI.Theme.ActiveBorderColor = []string{"red"}

	if err := Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Docker.Host != "tcp://192.168.1.1:2375" {
		t.Errorf("host mismatch: got %q", loaded.Docker.Host)
	}
	if loaded.UI.Theme.ActiveBorderColor[0] != "red" {
		t.Errorf("theme mismatch: got %v", loaded.UI.Theme.ActiveBorderColor)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yml")
	if err := os.WriteFile(cfgPath, []byte("invalid: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConfigKeymapDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Keymap.TabNext) == 0 || cfg.Keymap.TabNext[0] != "tab" {
		t.Error("expected TabNext to be 'tab'")
	}
	if len(cfg.Keymap.Up) < 2 || cfg.Keymap.Up[0] != "up" || cfg.Keymap.Up[1] != "k" {
		t.Error("expected Up to be 'up' and 'k'")
	}
}

func TestLoadKeymapOverridePreservesOtherDefaults(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(cfgPath, []byte("keymap:\n  help: [f3]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.Keymap.Help) != 1 || cfg.Keymap.Help[0] != "f3" {
		t.Fatalf("help override = %v", cfg.Keymap.Help)
	}
	if len(cfg.Keymap.Quit) == 0 || cfg.Keymap.Quit[0] != "q" {
		t.Fatalf("quit default was not preserved: %v", cfg.Keymap.Quit)
	}
}
