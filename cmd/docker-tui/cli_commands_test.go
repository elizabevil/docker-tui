package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestConfigInitWritesTemplate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	message, err := initConfigFile()
	if err != nil {
		t.Fatalf("initConfigFile: %v", err)
	}

	cfgFile, err := config.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}
	if !strings.Contains(message, cfgFile) {
		t.Fatalf("message %q missing path %q", message, cfgFile)
	}

	got, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("read %s: %v", cfgFile, err)
	}
	want, err := config.Template()
	if err != nil {
		t.Fatalf("Template: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("written file != template:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestConfigInitRefusesOverwrite(t *testing.T) {
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
	original := []byte("# existing user config\n")
	if err := os.WriteFile(cfgFile, original, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := initConfigFile(); err == nil {
		t.Fatalf("initConfigFile should refuse to overwrite existing file")
	}

	got, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("read %s: %v", cfgFile, err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("existing config was overwritten:\n--- got ---\n%s\n--- want ---\n%s", got, original)
	}
}

func TestConfigValidateOK(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfgFile, err := config.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}
	template, err := config.Template()
	if err != nil {
		t.Fatalf("Template: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfgFile), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(cfgFile, template, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	message, err := validateConfigFile(cfgFile)
	if err != nil {
		t.Fatalf("validateConfigFile: %v", err)
	}
	if want := "OK: " + cfgFile; message != want {
		t.Fatalf("message = %q, want %q", message, want)
	}
}

func TestConfigValidateRejectsBadConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfgFile, err := config.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}
	bad := []byte("version: 1\napp:\n  general:\n    lang: xx\n")
	if err := os.MkdirAll(filepath.Dir(cfgFile), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(cfgFile, bad, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	message, err := validateConfigFile(cfgFile)
	if err == nil {
		t.Fatalf("validateConfigFile should reject bad config, got %q", message)
	}
	if !strings.Contains(err.Error(), "general.lang") {
		t.Fatalf("error %q missing field path", err)
	}
	var exitErr *cliExitError
	if errors.As(err, &exitErr) {
		t.Fatalf("validation error should not carry exit code, got %d", exitErr.code)
	}
}

func TestConfigValidateMissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfgFile, err := config.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	message, err := validateConfigFile(cfgFile)
	if err == nil {
		t.Fatalf("validateConfigFile should fail for missing file, got %q", message)
	}
	var exitErr *cliExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("missing file error should carry exit code, got %T: %v", err, err)
	}
	if exitErr.code != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.code)
	}
	if !strings.Contains(err.Error(), "config init") {
		t.Fatalf("error %q missing init hint", err)
	}
}
