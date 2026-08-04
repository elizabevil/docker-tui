package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// AppName is the application name used for config dirs.
	AppName  = "github.com/elizabevil/docker-tui"
	localDir = "docker-tui"
)

// ConfigDir returns the configuration directory path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".config", localDir), nil
}

// ConfigFile returns the path to the config file.
func ConfigFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

// Load reads the configuration from the default path, merging with defaults.
// Uses LoadSplit for embedded defaults + user style overrides, then layers
// the CLI --config YAML on top (BR-043 §3.4 split config).
func Load(path string) (*Config, error) {
	cfg, err := LoadSplit("")
	if err != nil {
		return nil, fmt.Errorf("load embedded+user styles: %w", err)
	}

	if path == "" {
		var perr error
		path, perr = ConfigFile()
		if perr != nil {
			return cfg, nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config %s: %w", path, err)
	}

	return cfg, nil
}

// Save writes the configuration to the given path.
func Save(cfg *Config, path string) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	if path == "" {
		var err error
		path, err = ConfigFile()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}

	return nil
}

func Validate(cfg *Config) error {
	return cfg.Validate()
}

// Validate checks the complete configuration after decoding. Callers may use
// values directly after Load succeeds without repeating defensive fallbacks.
func (cfg *Config) Validate() error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.ConfigVersion != CurrentConfigVersion {
		return fmt.Errorf("configVersion must be %d", CurrentConfigVersion)
	}
	return cfg.Runtime.Validate()
}

func (r RuntimeConfig) Validate() error {
	if err := r.Health.Validate(); err != nil {
		return err
	}
	names := make(map[string]struct{}, len(r.Connections))
	for i, connection := range r.Connections {
		if err := connection.Validate(); err != nil {
			return fmt.Errorf("runtime.connections[%d]: %w", i, err)
		}
		if _, exists := names[connection.Name]; exists {
			return fmt.Errorf("runtime.connections[%d].name %q is duplicated", i, connection.Name)
		}
		names[connection.Name] = struct{}{}
	}
	if r.Default != "local-docker" && r.Default != "local-podman" {
		if _, exists := names[r.Default]; !exists {
			return fmt.Errorf("runtime.default %q does not name a connection", r.Default)
		}
	}
	return nil
}

func (h RuntimeHealthConfig) Validate() error {
	if h.IntervalSec <= 0 || h.TimeoutSec <= 0 || h.FailureThreshold <= 0 {
		return fmt.Errorf("runtime.health values must be greater than zero")
	}
	return nil
}

func (c RuntimeConn) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if c.Driver != "docker" && c.Driver != "podman" {
		return fmt.Errorf("driver must be docker or podman")
	}
	if strings.TrimSpace(c.Endpoint) == "" {
		return fmt.Errorf("endpoint is required")
	}
	endpoint, err := url.Parse(c.Endpoint)
	if err != nil || endpoint.Scheme == "" {
		return fmt.Errorf("endpoint must be a valid URI")
	}
	if c.TLS.Enabled && endpoint.Scheme == "unix" {
		return fmt.Errorf("tls cannot be enabled for a unix socket")
	}
	return c.TLS.Validate()
}

func (tls RuntimeTLSConfig) Validate() error {
	if !tls.Enabled {
		return nil
	}
	if tls.Verify == tls.InsecureSkipVerify {
		return fmt.Errorf("exactly one of tls.verify or tls.insecureSkipVerify must be true")
	}
	if tls.Verify && strings.TrimSpace(tls.CAFile) == "" {
		return fmt.Errorf("tls.caFile is required")
	}
	if (strings.TrimSpace(tls.CertFile) == "") != (strings.TrimSpace(tls.KeyFile) == "") {
		return fmt.Errorf("tls.certFile and keyFile must be configured together")
	}
	return nil
}
