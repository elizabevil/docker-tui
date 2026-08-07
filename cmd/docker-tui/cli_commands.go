package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agilira/orpheus/pkg/orpheus"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

// newInfoCommand builds the `dtui info` leaf command.
func newInfoCommand() *orpheus.Command {
	cmd := orpheus.NewCommand("info", "Show config, log and theme paths")
	cmd.SetHandler(func(ctx *orpheus.Context) error {
		return runInfo(ctx)
	})
	return cmd
}

// newConfigCommandGroup builds the `dtui config` command group with
// init and validate subcommands.
func newConfigCommandGroup() *orpheus.Command {
	cmd := orpheus.NewCommand("config", "Manage the config file")
	cmd.AddSubcommand(orpheus.NewCommand("init", "Create a config file from a full template").
		SetHandler(func(ctx *orpheus.Context) error {
			return runConfigInit(ctx)
		}))
	cmd.AddSubcommand(orpheus.NewCommand("validate", "Validate the config file").
		SetHandler(func(ctx *orpheus.Context) error {
			return runConfigValidate(ctx)
		}))
	return cmd
}

func runInfo(ctx *orpheus.Context) error {
	report, err := infoReport()
	if err != nil {
		return err
	}
	fmt.Print(report)
	return nil
}

func runConfigInit(ctx *orpheus.Context) error {
	message, err := initConfigFile()
	if err != nil {
		return err
	}
	fmt.Println(message)
	return nil
}

// initConfigFile writes the full annotated config template to
// config.ConfigFile(), creating the parent directory if needed.
// It refuses to overwrite an existing file and returns the written
// path on success.
func initConfigFile() (string, error) {
	path, err := config.ConfigFile()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("config file already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	template, err := config.Template()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, template, 0o644); err != nil {
		return "", err
	}
	return "Wrote " + path + " (full annotated template)", nil
}

func runConfigValidate(ctx *orpheus.Context) error {
	fmt.Println("config validate: not yet implemented")
	return nil
}

func infoReport() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	cfgFile, err := config.ConfigFile()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Config file: %s\n", cfgFile)
	fmt.Fprintf(&b, "Config dir: %s\n", dir)
	fmt.Fprintf(&b, "Log dir: %s\n", filepath.Join(dir, "logs"))
	fmt.Fprintf(&b, "Theme dir: %s\n", filepath.Join(dir, "themes"))
	fmt.Fprintf(&b, "Themes: %s\n", strings.Join(config.ListThemes(), ", "))
	return b.String(), nil
}


