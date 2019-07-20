// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package bootstrap provides the command to run the main task of processing all configured snowblocks.
package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/arcticicestudio/snowsaw/pkg/config"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
	"github.com/arcticicestudio/snowsaw/pkg/snowblock"
	"github.com/arcticicestudio/snowsaw/pkg/util/filesystem"
)

type cmdOptions struct {
	SnowblockPaths []string
}

// NewBootstrapCmd creates and configures a new `bootstrap` command.
func NewBootstrapCmd() *cobra.Command {
	o := cmdOptions{}
	bootstrapCmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstraps all configured snowblocks",
		Long: `Bootstraps all configured snowblocks
To process individual snowblocks a list of space-separated paths can be passed as arguments.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			o.prepare(args)
			o.run()
		},
	}
	return bootstrapCmd
}

func (o *cmdOptions) prepare(args []string) {
	// Use explicit snowblocks if specified, otherwise find all snowblocks within the base directories.
	if len(args) > 0 {
		prt.Debugf("Using individual snowblocks instead of configured base directories(s): %s",
			color.CyanString("%v", args))
		config.AppConfig.Snowblocks.Paths = args
	} else if err := o.readSnowblockDirectories(); err != nil {
		prt.Errorf("Failed to read snowblocks from base directories: %v", err)
		os.Exit(1)
	}
	o.SnowblockPaths = config.AppConfig.Snowblocks.Paths
}

func (o *cmdOptions) readSnowblockDirectories() error {
	var validBaseDirs []string
	for _, baseDir := range config.AppConfig.Snowblocks.BaseDirs {
		expBaseDir, expBaseDirErr := filesystem.ExpandPath(baseDir)
		if expBaseDirErr != nil {
			return fmt.Errorf("could not expand base snowblock directory path: %v", expBaseDirErr)
		}
		baseDirExists, baseDirExistsChkErr := filesystem.DirExists(expBaseDir)
		if baseDirExistsChkErr != nil {
			return fmt.Errorf("could not read snowblock base directory: %v", baseDirExistsChkErr)
		}
		if baseDirExists {
			sbDirs, sbDirListErr := ioutil.ReadDir(expBaseDir)
			if sbDirListErr != nil {
				return fmt.Errorf("could not read snowblock base directory: %s", color.RedString("%v", sbDirListErr))
			}
			for _, sbFileInfo := range sbDirs {
				if sbFileInfo.IsDir() {
					config.AppConfig.Snowblocks.Paths = append(
						config.AppConfig.Snowblocks.Paths, filepath.Join(expBaseDir, sbFileInfo.Name()))
				}
			}
			validBaseDirs = append(validBaseDirs, baseDir)
			continue
		}
		prt.Warnf("Ignoring non-existent snowblock base directory: %s", color.CyanString(baseDir))
	}

	if len(validBaseDirs) > 0 {
		prt.Debugf("Processing configured snowblock base directories: %s", color.CyanString("%v", validBaseDirs))
	}
	return nil
}

func (o *cmdOptions) run() {
	for _, path := range o.SnowblockPaths {
		sb := snowblock.NewSnowblock(path)
		err := sb.Validate(config.SnowblockTaskRunnerRegistry.GetAll())
		if err != nil {
			prt.Errorf("Failed to validate snowblock %s: %v",
				color.CyanString(filepath.Base(path)), color.RedString(err.Error()))
			os.Exit(1)
		}
		if !sb.IsValid {
			prt.Warnf("Skipped processing of invalid snowblock %s", color.CyanString(filepath.Base(sb.Path)))
			continue
		}

		err = sb.Dispatch()
		if err != nil {
			prt.Errorf("Failed to process snowblock %s: %v",
				color.CyanString(filepath.Base(path)), color.RedString(err.Error()))
			os.Exit(1)
		}
		if sb.IsValid {
			prt.Successf("Successfully bootstrapped snowblock %s", color.CyanString(filepath.Base(path)))
		}
	}

	if len(o.SnowblockPaths) > 0 {
		prt.Successf("Bootstrapped all configured snowblocks")
	} else {
		prt.Warnf("No valid snowblocks found")
	}
}
