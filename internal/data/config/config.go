package config

import (
	"bytes"
	"fmt"
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
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		var err error
		path, err = ConfigFile()
		if err != nil {
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
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config %s: %w", path, err)
	}

	return cfg, nil
}

// Save writes the configuration to the given path.
func Save(cfg *Config, path string) error {
	if err := Validate(cfg); err != nil {
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
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.ConfigVersion != CurrentConfigVersion {
		return fmt.Errorf("configVersion must be %d", CurrentConfigVersion)
	}
	if cfg.Runtime.Health.IntervalSec <= 0 || cfg.Runtime.Health.TimeoutSec <= 0 || cfg.Runtime.Health.FailureThreshold <= 0 {
		return fmt.Errorf("runtime.health values must be greater than zero")
	}
	names := make(map[string]struct{}, len(cfg.Runtime.Connections))
	for i, connection := range cfg.Runtime.Connections {
		prefix := fmt.Sprintf("runtime.connections[%d]", i)
		if strings.TrimSpace(connection.Name) == "" {
			return fmt.Errorf("%s.name is required", prefix)
		}
		if _, exists := names[connection.Name]; exists {
			return fmt.Errorf("%s.name %q is duplicated", prefix, connection.Name)
		}
		names[connection.Name] = struct{}{}
		if connection.Driver != "docker" && connection.Driver != "podman" {
			return fmt.Errorf("%s.driver must be docker or podman", prefix)
		}
		if strings.TrimSpace(connection.Endpoint) == "" {
			return fmt.Errorf("%s.endpoint is required", prefix)
		}
		if connection.TLS.Enabled {
			if !connection.TLS.Verify {
				return fmt.Errorf("%s.tls.verify must be true", prefix)
			}
			if connection.TLS.CAFile == "" {
				return fmt.Errorf("%s.tls.caFile is required", prefix)
			}
			if (connection.TLS.CertFile == "") != (connection.TLS.KeyFile == "") {
				return fmt.Errorf("%s.tls.certFile and keyFile must be configured together", prefix)
			}
		}
	}
	if cfg.Runtime.Default == "local-docker" && !cfg.Runtime.Discovery.LocalDocker {
		return fmt.Errorf("runtime.default local-docker requires runtime.discovery.localDocker")
	}
	if cfg.Runtime.Default == "local-podman" && !cfg.Runtime.Discovery.LocalPodman {
		return fmt.Errorf("runtime.default local-podman requires runtime.discovery.localPodman")
	}
	if cfg.Runtime.Default != "local-docker" && cfg.Runtime.Default != "local-podman" {
		if _, exists := names[cfg.Runtime.Default]; !exists {
			return fmt.Errorf("runtime.default %q does not name a connection", cfg.Runtime.Default)
		}
	}
	return nil
}
