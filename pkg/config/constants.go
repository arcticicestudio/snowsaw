// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

package config

import (
	"github.com/arcticicestudio/snowsaw/pkg/api/snowblock"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
	"github.com/arcticicestudio/snowsaw/pkg/snowblock/task"
	"github.com/arcticicestudio/snowsaw/pkg/snowblock/task/link"
)

const (
	// DefaultLoggingLevel is the default application-wide level of logging the verbosity.
	DefaultLoggingLevel = "info"

	// DefaultSnowblocksBaseDirectoryName is the default name of the snowblocks base directory.
	DefaultSnowblocksBaseDirectoryName = "snowblocks"

	// PackageName is the name of this Go module
	PackageName = "github.com/arcticicestudio/" + ProjectName

	// ProjectName is the name of the project.
	ProjectName = "snowsaw"
)

var (
	// AppConfig is the main application configuration with initial default values.
	AppConfig = Config{}

	// AppConfigPaths is the default paths the application will search for configuration files.
	AppConfigPaths []*file.File

	availableTaskRunner = []snowblock.TaskRunner{
		&link.Link{},
	}

	// BuildDateTime is the date and time this application was build.
	BuildDateTime string

	// SnowblockTaskRunnerRegistry is the application-wide registry for snowblock task runner.
	SnowblockTaskRunnerRegistry = task.NewRegistry()

	// Version is the application version.
	Version = "0.0.0"
)
