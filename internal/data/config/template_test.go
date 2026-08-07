package config

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestTemplateGolden(t *testing.T) {
	got, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}
	golden := filepath.Join("testdata", "config-template.golden.yml")
	if *updateGolden {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to regenerate)", golden, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("template mismatch with golden:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestTemplateSections(t *testing.T) {
	got, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}
	out := string(got)
	for _, section := range []string{"version:", "general:", "ui:", "docker:", "runtime:", "keymap:", "logs:", "layout:", "commands:", "appearance:", "theme:"} {
		if !strings.Contains(out, section) {
			t.Errorf("template missing section %q", section)
		}
	}
	if strings.Contains(out, "operations") {
		t.Errorf("template must not contain an operations section")
	}
}

func TestTemplateEverySectionHasComment(t *testing.T) {
	got, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}
	lines := strings.Split(string(got), "\n")
	for _, section := range []string{"  general:", "  ui:", "  docker:", "  runtime:", "  keymap:", "  logs:", "  layout:", "  commands:", "appearance:"} {
		found := false
		for i, line := range lines {
			if line == section {
				found = true
				if i == 0 || !strings.HasPrefix(strings.TrimSpace(lines[i-1]), "#") {
					t.Errorf("section %q has no comment line above it (line %d: %q)", section, i, lines[i-1])
				}
			}
		}
		if !found {
			t.Errorf("section %q not found in template", section)
		}
	}
}

func TestTemplateStrictDecode(t *testing.T) {
	got, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}
	var user UserConfig
	dec := yaml.NewDecoder(bytes.NewReader(got))
	dec.KnownFields(true)
	if err := dec.Decode(&user); err != nil {
		t.Fatalf("template does not strictly decode: %v", err)
	}
	if user.Version != CurrentConfigVersion {
		t.Errorf("version = %d, want %d", user.Version, CurrentConfigVersion)
	}
	if user.Appearance.Theme != ThemeName(themeDefaultName) {
		t.Errorf("appearance.theme = %q, want %q", user.Appearance.Theme, themeDefaultName)
	}
	if user.App.General == nil {
		t.Fatal("app.general missing")
	}
	if *user.App.General.ScrollHeight != 2 {
		t.Errorf("general.scrollHeight = %d, want 2", *user.App.General.ScrollHeight)
	}
	if *user.App.General.Lang != LanguageEnglish {
		t.Errorf("general.lang = %q, want %q", *user.App.General.Lang, LanguageEnglish)
	}
	if user.App.Docker == nil {
		t.Fatal("app.docker missing")
	}
	if *user.App.Docker.Timeout != DefaultDockerTimeout {
		t.Errorf("docker.timeout = %v, want %v", *user.App.Docker.Timeout, DefaultDockerTimeout)
	}
	if user.App.Logs == nil || *user.App.Logs.Tail != DefaultLogsTail {
		t.Errorf("logs.tail = %v, want %q", user.App.Logs, DefaultLogsTail)
	}
	if user.App.Runtime == nil || len(*user.App.Runtime.Connections) != 0 {
		t.Errorf("runtime.connections = %v, want empty", user.App.Runtime)
	}
}
