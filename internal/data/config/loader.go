package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/elizabevil/docker-tui/internal/utils"
	"gopkg.in/yaml.v3"
)

type AppearanceConfig struct {
	Theme     ThemeName  `json:"theme" yaml:"theme"`
	Overrides ThemePatch `json:"overrides" yaml:"overrides"`
}

type UserConfig struct {
	Version    int              `json:"version" yaml:"version"`
	App        AppPatch         `json:"app" yaml:"app"`
	Appearance AppearanceConfig `json:"appearance" yaml:"appearance"`
}

type LoadOptions struct {
	ConfigPath string
	ThemeName  ThemeName
	Lang       Language
}

type Resolved struct {
	App       *AppConfig
	Theme     *Theme
	ThemeName ThemeName
}

var embeddedDefaultFiles = [...]string{
	defaultGeneralFile, defaultUIFile, defaultDockerFile, defaultRuntimeFile,
	defaultLogsFile, defaultLayoutFile, defaultCommandsFile, defaultKeymapFile,
}

func LoadResolved(options LoadOptions) (*Resolved, error) {
	app := DefaultAppConfig()
	for _, name := range embeddedDefaultFiles {
		data, err := fs.ReadFile(DefaultsFS(), name)
		if err != nil {
			return nil, fmt.Errorf(errReadEmbeddedDefaultFormat, name, err)
		}
		var patch AppPatch
		if err := decodeJSONCStrict(data, &patch); err != nil {
			return nil, fmt.Errorf(errParseEmbeddedDefaultFormat, name, err)
		}
		if err := validateEmbeddedDefaultScope(name, patch); err != nil {
			return nil, err
		}
		patch.Apply(app)
	}

	user, configPath, err := loadUserConfig(options.ConfigPath)
	if err != nil {
		return nil, err
	}
	if user != nil {
		if user.Version != CurrentConfigVersion {
			return nil, fmt.Errorf(errConfigVersionFormat, configPath, CurrentConfigVersion)
		}
		user.App.Apply(app)
	}
	if options.Lang != Language(emptyValue) {
		app.General.Lang = options.Lang
	}
	if err := ValidateApp(app); err != nil {
		return nil, fmt.Errorf(errValidateAppFormat, err)
	}

	theme := DefaultTheme()
	base, err := loadNamedTheme(ThemeName(themeDefaultName), emptyValue, theme)
	if err != nil {
		return nil, err
	}
	theme = base.Theme
	themeName := ThemeName(themeDefaultName)
	if user != nil && user.Appearance.Theme != ThemeName(emptyValue) {
		themeName = user.Appearance.Theme
	}
	if options.ThemeName != ThemeName(emptyValue) {
		themeName = options.ThemeName
	}
	themeDir := emptyValue
	if configPath != emptyValue {
		themeDir = filepath.Join(filepath.Dir(configPath), userThemesDirName)
	}
	if themeName != ThemeName(themeDefaultName) {
		selected, loadErr := loadNamedTheme(themeName, themeDir, theme)
		if loadErr != nil {
			return nil, loadErr
		}
		theme = selected.Theme
	} else if themeDir != emptyValue {
		path := filepath.Join(themeDir, themeDefaultName+jsoncExtension)
		if data, readErr := os.ReadFile(path); readErr == nil {
			selected, parseErr := loadThemeDocument(data, path, theme)
			if parseErr != nil {
				return nil, parseErr
			}
			theme = selected.Theme
		} else if !os.IsNotExist(readErr) {
			return nil, fmt.Errorf(errReadThemeFormat, path, readErr)
		}
	}
	if user != nil {
		user.Appearance.Overrides.Apply(theme)
	}
	if err := ValidateTheme(theme); err != nil {
		return nil, fmt.Errorf(errValidateResolvedThemeFormat, err)
	}
	return &Resolved{App: app, Theme: theme, ThemeName: themeName}, nil
}

func validateEmbeddedDefaultScope(name string, patch AppPatch) error {
	matches := false
	switch name {
	case defaultGeneralFile:
		matches = patch.General != nil
		patch.General = nil
	case defaultUIFile:
		matches = patch.UI != nil
		patch.UI = nil
	case defaultDockerFile:
		matches = patch.Docker != nil
		patch.Docker = nil
	case defaultRuntimeFile:
		matches = patch.Runtime != nil
		patch.Runtime = nil
	case defaultLogsFile:
		matches = patch.Logs != nil
		patch.Logs = nil
	case defaultLayoutFile:
		matches = patch.Layout != nil
		patch.Layout = nil
	case defaultCommandsFile:
		matches = patch.Commands != nil
		patch.Commands = nil
	case defaultKeymapFile:
		matches = patch.Keymap != nil
		patch.Keymap = nil
	default:
		return fmt.Errorf(errEmbeddedScopeUnknownFormat, name)
	}
	if !matches {
		return fmt.Errorf(errEmbeddedScopeMissingFormat, name)
	}
	if patch.General != nil || patch.UI != nil || patch.Docker != nil || patch.Runtime != nil ||
		patch.Keymap != nil || patch.Logs != nil || patch.Layout != nil || patch.Commands != nil {
		return fmt.Errorf(errEmbeddedScopeExtraFormat, name)
	}
	return nil
}

func loadUserConfig(path string) (*UserConfig, string, error) {
	if path == emptyValue {
		var err error
		path, err = ConfigFile()
		if err != nil {
			return nil, emptyValue, nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, path, nil
		}
		return nil, path, fmt.Errorf(errReadConfigFormat, path, err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, path, fmt.Errorf(errParseConfigFormat, path, err)
	}
	if err := rejectYAMLNulls(&document); err != nil {
		return nil, path, fmt.Errorf(errParseConfigFormat, path, err)
	}
	var user UserConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&user); err != nil {
		return nil, path, fmt.Errorf(errParseConfigFormat, path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = errors.New(errMultipleYAMLDocuments)
		}
		return nil, path, fmt.Errorf(errParseConfigFormat, path, err)
	}
	return &user, path, nil
}

func decodeJSONCStrict(data []byte, target any) error {
	clean := utils.StripJSONCComments(data)
	if err := rejectJSONNulls(clean); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(clean))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New(errMultipleJSONDocuments)
		}
		return err
	}
	return nil
}

func rejectJSONNulls(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if token == nil {
			return errors.New(errNullNotAllowed)
		}
	}
}

func rejectYAMLNulls(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Tag == yamlNullTag {
		return fmt.Errorf(errNullAtLineFormat, node.Line)
	}
	for _, child := range node.Content {
		if err := rejectYAMLNulls(child); err != nil {
			return err
		}
	}
	return nil
}
