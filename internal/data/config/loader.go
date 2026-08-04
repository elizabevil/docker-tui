package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// LoadSplit produces a Config by merging the embedded split defaults, then
// any user override files in userDir (defaults to ~/.config/docker-tui/styles
// when empty). The merged map is unmarshaled into the Config struct.
//
// Merge order (later overrides earlier):
//  1. embedded split files in styles/*.jsonc
//  2. user override files in userDir/*.jsonc (alphabetical)
//  3. CLI YAML config (applied by callers via Load)
//
// User overrides use the same schema as the split files (each file is
// `{"section_name": {...}}`); missing sections fall back to defaults.
// Returns an error only on parse failure (missing dir is not an error).
func LoadSplit(userDir string) (*Config, error) {
	merged, err := loadEmbeddedStyles()
	if err != nil {
		return nil, fmt.Errorf("load embedded defaults: %w", err)
	}

	if userDir == "" {
		userDir = defaultStylesDir()
	}
	if entries, err := os.ReadDir(userDir); err == nil {
		paths := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(name, ".jsonc") {
				continue
			}
			paths = append(paths, filepath.Join(userDir, name))
		}
		sort.Strings(paths) // deterministic order for tests + predictable layering
		for _, p := range paths {
			data, rerr := os.ReadFile(p)
			if rerr != nil {
				return nil, fmt.Errorf("read %s: %w", p, rerr)
			}
			if merr := mergeJSONCSection(merged, data); merr != nil {
				return nil, fmt.Errorf("parse %s: %w", p, merr)
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read user styles dir %s: %w", userDir, err)
	}

	cfg, err := mergedToConfig(merged)
	if err != nil {
		return nil, fmt.Errorf("build config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate merged config: %w", err)
	}
	return cfg, nil
}

// loadEmbeddedStyles reads every *.jsonc under the embedded styles dir and
// merges its top-level keys into a single map[string]json.RawMessage.
func loadEmbeddedStyles() (map[string]json.RawMessage, error) {
	merged := map[string]json.RawMessage{}
	entries, err := fs.ReadDir(StylesFS(), ".")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonc") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // stable order regardless of embed.FS traversal
	for _, name := range names {
		data, err := fs.ReadFile(StylesFS(), name)
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", name, err)
		}
		if err := mergeJSONCSection(merged, data); err != nil {
			return nil, fmt.Errorf("parse embedded %s: %w", name, err)
		}
	}
	return merged, nil
}

// mergeJSONCSection unmarshals data as a top-level JSON object and merges each
// key into dst. Later writes win (same behavior as `defaults | user`).
func mergeJSONCSection(dst map[string]json.RawMessage, data []byte) error {
	var section map[string]json.RawMessage
	if err := utils.UnmarshalJSONCSonic(data, &section); err != nil {
		return err
	}
	for k, v := range section {
		dst[k] = v
	}
	return nil
}

// mergedToConfig marshals the merged map and unmarshals into Config so the
// YAML/JSON tags drive field assignment (deep nested merge without reflection).
func mergedToConfig(merged map[string]json.RawMessage) (*Config, error) {
	if len(merged) == 0 {
		return DefaultConfig(), nil
	}
	// Ensure configVersion is set before unmarshaling (defaults assume it).
	if _, ok := merged["configVersion"]; !ok {
		merged["configVersion"] = json.RawMessage(fmt.Sprintf("%d", CurrentConfigVersion))
	}
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, _ := json.Marshal(k)
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(merged[k])
	}
	buf.WriteByte('}')
	cfg := DefaultConfig() // seed with theme defaults; merged JSON overrides
	if err := utils.UnmarshalJSONCSonic([]byte(buf.String()), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// defaultStylesDir returns ~/.config/docker-tui/styles for user overrides.
func defaultStylesDir() string {
	if d, err := ConfigDir(); err == nil {
		return filepath.Join(d, "styles")
	}
	return ""
}