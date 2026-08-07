package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestInfoReportContainsPathsAndThemes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	report, err := infoReport()
	if err != nil {
		t.Fatalf("infoReport: %v", err)
	}

	dir, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir: %v", err)
	}
	cfgFile, err := config.ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile: %v", err)
	}

	for _, want := range []string{
		"Config file: " + cfgFile,
		"Config dir: " + dir,
		"Log dir: " + filepath.Join(dir, "logs"),
		"Theme dir: " + filepath.Join(dir, "themes"),
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("report missing %q:\n%s", want, report)
		}
	}

	themes := config.ListThemes()
	if len(themes) == 0 {
		t.Fatalf("ListThemes returned empty")
	}
	for _, theme := range themes {
		if !strings.Contains(report, theme) {
			t.Fatalf("report missing theme %q:\n%s", theme, report)
		}
	}
}

func TestInfoReportErrorsWhenNoHome(t *testing.T) {
	oldHome, hadHome := os.LookupEnv("HOME")
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	if !hadHome {
		t.Cleanup(func() { os.Unsetenv("HOME") })
	} else {
		t.Cleanup(func() { t.Setenv("HOME", oldHome) })
	}

	if _, err := infoReport(); err == nil {
		t.Fatalf("infoReport should error without HOME")
	}
}
