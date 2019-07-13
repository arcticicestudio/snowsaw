// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package snowsaw provides the root command of the application and bootstraps the startup.
package snowsaw

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/arcticicestudio/snowsaw/cmd/snowsaw/info"
	"github.com/arcticicestudio/snowsaw/pkg/config"
	"github.com/arcticicestudio/snowsaw/pkg/config/builder"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
)

var (
	// debug indicates if the `debug` flag has been set to enable configure the logging for the debug scope.
	debug bool
	// explicitConfigFilePath stores the path to the application configuration file when the `config` flag is specified.
	// It takes precedence over the default application configuration paths.
	explicitConfigFilePath string
	// snowblockBaseDirs stores the comma-separated paths to the snowblock directories when the `basedirs` flag is
	// specified.
	// It takes precedence over the path of the application configuration file(s).
	snowblockBaseDirs []string
)

// rootCmd is the root command of the application.
var rootCmd = &cobra.Command{
	Use:   config.ProjectName,
	Short: "A lightweight, plugin-driven and dynamic dotfiles bootstrapper.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			prt.Errorf("Failed to run %s: %v", config.ProjectName, err)
			os.Exit(1)
		}
	},
}

// Run is the main application function that adds all child commands to the root command and sets flags appropriately.
// This is called by `main.main()` and only needs to be run once for the root command.
func Run() {
	// Disable cobra's verbose errors and replace with custom formatted CLI logger from the `pkg/prt` package.
	rootCmd.SilenceErrors = true

	// Run the application with the given commands, flags and arguments and exit on any (downstream) error.
	if err := rootCmd.Execute(); err != nil {
		prt.Errorf(err.Error())
		os.Exit(1)
	}
}

func init() {
	// Specify the functions to be run before each command gets executed.
	cobra.OnInitialize(initDebugScope, initConfig, initPrinter, mergeConfigValues)

	// Define global application flags.
	rootCmd.PersistentFlags().StringVar(&explicitConfigFilePath, "config", "", "path to the configuration file")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug information output")
	rootCmd.PersistentFlags().StringSliceVarP(
		&snowblockBaseDirs, "basedirs", "b", config.AppConfig.Snowblocks.Paths,
		"comma-separated paths to snowblock base directories")

	// Set the app version information for the automatically generated `version` flag.
	rootCmd.Version = color.CyanString(config.Version)
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Create and register all subcommands.
	rootCmd.AddCommand(info.NewInfoCmd())
	rootCmd.AddCommand(bootstrap.NewBootstrapCmd())
}

// initConfig searches and loads either the default application configuration file paths or the explicit file at the
// given path specified through the global `config` flag.
func initConfig() {
	if explicitConfigFilePath != "" {
		if err := builder.Load(file.NewFile(explicitConfigFilePath)).Into(&config.AppConfig, false); err != nil {
			prt.Errorf("Failed to load application configuration file: %v", err)
			os.Exit(1)
		}
	} else {
		b := builder.Load(config.AppConfigPaths...)
		if len(b.Files) == 0 {
			prt.Debugf("No configuration files found, using application defaults.")
		}
		if err := b.Into(&config.AppConfig, true); err != nil {
			prt.Errorf("Failed to load application configuration file(s): %v", err)
			os.Exit(1)
		}
	}
}

// initDebugScope configures the application when run with debug scope.
func initDebugScope() {
	if debug {
		prt.SetVerbosityLevel(prt.DebugVerbosity)
	}
}

// setPrinterVerbosityLevel configures the global CLI printer like the verbosity level.
func initPrinter() {
	// The `debug` flag always takes precedence and overrides the application configuration value.
	if !debug {
		lvl, err := prt.ParseVerbosityLevel(strings.ToUpper(config.AppConfig.LogLevel))
		if err != nil {
			prt.Debugf("Error while parsing log level from configuration: %v", err)
			prt.Debugf("Using default INFO level as fallback")
			prt.SetVerbosityLevel(prt.InfoVerbosity)
		} else {
			prt.Debugf("Using configured logger level: %s", strings.ToUpper(config.AppConfig.LogLevel))
			prt.SetVerbosityLevel(lvl)
		}
	}
}

// mergeConfigValues merges the specified (global) flags and merges the corresponding values with the default
// application configuration values.
// Since flags have the highest precedence their value will override application defaults as well as value from
// loaded application configuration files.
func mergeConfigValues() {
	// Use configured or individual snowblock base directories if specified, otherwise fall back to the default directory.
	if len(config.AppConfig.Snowblocks.BaseDirs) == 0 {
		config.AppConfig.Snowblocks.BaseDirs = append(
			config.AppConfig.Snowblocks.BaseDirs, config.DefaultSnowblocksBaseDirectoryName)
	}
	if len(snowblockBaseDirs) > 0 {
		config.AppConfig.Snowblocks.BaseDirs = snowblockBaseDirs
	}

	// If the logging level has not been specified fall back to the default level.
	if config.AppConfig.LogLevel == "" {
		config.AppConfig.LogLevel = config.DefaultLoggingLevel
	}
}
