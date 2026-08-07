package main

import (
	"fmt"
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
	fmt.Println("config init: not yet implemented")
	return nil
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


