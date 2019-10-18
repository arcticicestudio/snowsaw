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
	"github.com/arcticicestudio/snowsaw/pkg/snowblock/task/clean"
	"github.com/arcticicestudio/snowsaw/pkg/snowblock/task/link"
	"github.com/arcticicestudio/snowsaw/pkg/snowblock/task/shell"
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

	// AppVersion is the application version.
	AppVersion = "0.0.0"

	// AppVersionBuildDateTime is the date and time when this application version was built.
	AppVersionBuildDateTime string

	// AppVersionGoRuntime is the Go runtime version with which this application was built.
	AppVersionGoRuntime string

	availableTaskRunner = []snowblock.TaskRunner{
		&clean.Clean{},
		&link.Link{},
		&shell.Shell{},
	}

	// SnowblockTaskRunnerRegistry is the application-wide registry for snowblock task runner.
	SnowblockTaskRunnerRegistry = task.NewRegistry()
)
