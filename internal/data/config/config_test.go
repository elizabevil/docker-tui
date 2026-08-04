package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultAppConfig(t *testing.T) {
	cfg := DefaultAppConfig()
	if err := ValidateApp(cfg); err != nil {
		t.Fatalf("compiled fallback is invalid: %v", err)
	}
	if cfg.Version != CurrentConfigVersion {
		t.Fatalf("version = %d", cfg.Version)
	}
	if cfg.Keymap.Navigation.Up.Primary != "up" || cfg.Keymap.Navigation.Up.Secondary != "k" {
		t.Fatalf("unexpected navigation binding: %#v", cfg.Keymap.Navigation.Up)
	}
	if cfg.Commands.DockerCompose.Executable != "docker" {
		t.Fatalf("unexpected compose command: %#v", cfg.Commands.DockerCompose)
	}
}

func TestAppPatchPreservesOmittedFieldsAndExplicitZeroValues(t *testing.T) {
	cfg := DefaultAppConfig()
	falseValue := false
	zero := 0
	lang := LanguageChinese
	patch := AppPatch{
		General: &GeneralPatch{Lang: &lang},
		UI:      &UIPatch{EnableMouse: &falseValue, HintTimeout: &zero},
	}
	patch.Apply(cfg)

	if cfg.General.Lang != "zh" || cfg.General.SizeFormat != "binary" {
		t.Fatalf("general patch result: %#v", cfg.General)
	}
	if cfg.UI.EnableMouse || cfg.UI.HintTimeout != 0 {
		t.Fatalf("explicit zero values were not preserved: %#v", cfg.UI)
	}
}

func TestAppPatchReplacesConnectionCollection(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Runtime.Connections = []RuntimeConnection{{Name: "old"}}
	connections := []RuntimeConnection{{Name: "new", Driver: "docker", Endpoint: "unix:///tmp/docker.sock"}}
	AppPatch{Runtime: &RuntimePatch{Connections: &connections}}.Apply(cfg)
	if len(cfg.Runtime.Connections) != 1 || cfg.Runtime.Connections[0].Name != "new" {
		t.Fatalf("connections were not replaced: %#v", cfg.Runtime.Connections)
	}
}

func TestLoadResolvedLayersEmbeddedUserAndCLI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	data := []byte(`version: 1
app:
  general:
    lang: zh
  ui:
    enableMouse: false
    window:
      contentWidthPercent: 88
    table:
      rowPrefixSelected: "> "
  runtime:
    health:
      intervalSec: 5
appearance:
  theme: nord
  overrides:
    dialog:
      border:
        token: primary
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	resolved, err := LoadResolved(LoadOptions{ConfigPath: path, Lang: "en"})
	if err != nil {
		t.Fatalf("LoadResolved: %v", err)
	}
	if resolved.App.General.Lang != "en" {
		t.Fatalf("CLI lang did not win: %q", resolved.App.General.Lang)
	}
	if resolved.App.UI.EnableMouse {
		t.Fatal("explicit false was not preserved")
	}
	if resolved.App.UI.Window.ContentWidthPercent != 88 || resolved.App.UI.Table.RowPrefixSelected != "> " {
		t.Fatalf("typed UI scopes were not applied: %#v", resolved.App.UI)
	}
	if resolved.App.Runtime.Health.IntervalSec != 5 || resolved.App.Runtime.Health.TimeoutSec != 2 {
		t.Fatalf("nested patch lost defaults: %#v", resolved.App.Runtime.Health)
	}
	if resolved.ThemeName != ThemeName("nord") || resolved.Theme.Dialog.Border != TokenRef(ColorTokenPrimary) {
		t.Fatalf("theme cascade failed: name=%q dialog=%#v", resolved.ThemeName, resolved.Theme.Dialog)
	}
	if resolved.Theme.Palette.Primary != "#81a1c1" {
		t.Fatalf("selected theme palette not applied: %q", resolved.Theme.Palette.Primary)
	}
}

func TestLoadResolvedCLIThemeChangesBaseButKeepsUserOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	data := []byte(`version: 1
appearance:
  theme: light
  overrides:
    dialog:
      border:
        token: accentSecondary
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := LoadResolved(LoadOptions{ConfigPath: path, ThemeName: "nord"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ThemeName != "nord" || resolved.Theme.Palette.Background != "#242933" {
		t.Fatalf("CLI theme did not select nord: %#v", resolved.Theme.Palette)
	}
	if resolved.Theme.Dialog.Border != TokenRef(ColorTokenAccentSecondary) {
		t.Fatalf("user property override did not remain highest: %q", resolved.Theme.Dialog.Border)
	}
}

func TestLoadResolvedRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("version: 1\napp:\n  runtime:\n    typo: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadResolved(LoadOptions{ConfigPath: path}); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestLoadResolvedRejectsNull(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("version: 1\napp:\n  general:\n    lang: null\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadResolved(LoadOptions{ConfigPath: path}); err == nil {
		t.Fatal("expected null to be rejected")
	}
}

func TestLoadResolvedRejectsUnknownThemeScope(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("version: 1\nappearance:\n  theme: invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "themes"), 0o700); err != nil {
		t.Fatal(err)
	}
	theme := []byte(`{"meta":{"name":"Invalid"},"theme":{"runtime":{"default":"bad"}}}`)
	if err := os.WriteFile(filepath.Join(dir, "themes", "invalid.jsonc"), theme, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadResolved(LoadOptions{ConfigPath: path}); err == nil {
		t.Fatal("expected forbidden theme scope to fail strict decoding")
	}
}

func TestThemePatchRejectsStringColorReference(t *testing.T) {
	var patch ThemePatch
	err := decodeJSONCStrict([]byte(`{"dialog":{"border":"primary"}}`), &patch)
	if err == nil {
		t.Fatal("expected string color reference to be rejected")
	}
}

func TestLoadResolvedRejectsThemePathTraversal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("version: 1\nappearance:\n  theme: ../outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadResolved(LoadOptions{ConfigPath: path}); err == nil {
		t.Fatal("expected unsafe theme name to fail")
	}
}

func TestValidateEmbeddedDefaultScope(t *testing.T) {
	patch := AppPatch{General: &GeneralPatch{}, Runtime: &RuntimePatch{}}
	if err := validateEmbeddedDefaultScope("general.jsonc", patch); err == nil {
		t.Fatal("expected cross-scope embedded default to fail")
	}
}

func TestValidateAppRejectsInvalidDialogPosition(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.UI.Dialog.Position.Horizontal = "floating"
	if err := ValidateApp(cfg); err == nil {
		t.Fatal("expected invalid dialog position to fail")
	}
}

func TestValidateAppRejectsInvalidKeyBinding(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Keymap.Global.Help = KeyBinding{Secondary: "f1"}
	if err := ValidateApp(cfg); err == nil {
		t.Fatal("expected secondary key without primary to fail")
	}
}

func TestRuntimeConnectionValidation(t *testing.T) {
	valid := RuntimeConnection{Name: "remote", Driver: "docker", Endpoint: "tcp://example:2375"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid connection rejected: %v", err)
	}
	invalid := valid
	invalid.Endpoint = "missing-scheme"
	if err := invalid.Validate(); err == nil {
		t.Fatal("invalid endpoint accepted")
	}
}

func TestRuntimeHealthDurations(t *testing.T) {
	health := RuntimeHealthConfig{IntervalSec: 7, TimeoutSec: 4, FailureThreshold: 3}
	if health.Interval() != 7*time.Second || health.Timeout() != 4*time.Second {
		t.Fatalf("durations = %s, %s", health.Interval(), health.Timeout())
	}
}
