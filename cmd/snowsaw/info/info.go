// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package info provides the info command to print more detailed application information.
package info

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/arcticicestudio/snowsaw/pkg/config"
)

// NewInfoCmd creates and configures a new `info` command.
func NewInfoCmd() *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Prints more detailed application information",
		Run: func(cmd *cobra.Command, args []string) {
			if config.BuildDateTime != "" {
				fmt.Println(fmt.Sprintf("%s %s (build %s)",
					color.CyanString(config.ProjectName),
					color.BlueString(config.Version),
					color.GreenString(config.BuildDateTime)))
			} else {
				fmt.Println(fmt.Sprintf("%s %s",
					color.CyanString(config.ProjectName),
					color.BlueString(config.Version)))
			}
		},
	}

	return infoCmd
}
