// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package version provides the version command to print more detailed application version information.
package version

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/arcticicestudio/snowsaw/pkg/config"
)

// NewVersionCmd creates and configures a new `version` command.
func NewVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints more detailed application version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fmt.Sprintf("%s %s build at %s with %s",
				color.CyanString(config.ProjectName),
				color.BlueString(config.AppVersion),
				color.GreenString(config.AppVersionBuildDateTime),
				color.BlueString(config.AppVersionGoRuntime)))
		},
	}

	return versionCmd
}
