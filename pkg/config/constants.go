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
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/arcticicestudio/snowsaw/pkg/config/encoder"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
)

const (
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

	// BuildDateTime is the date and time this application was build.
	BuildDateTime string

	// Version is the application version.
	Version = "0.0.0"
)

func init() {
	AppConfig = Config{
		LogLevel: "info",
	}
	AppConfigPaths = genConfigPaths()
}

func genConfigPaths() []*file.File {
	var files []*file.File

	// Include the user-level dotfile configuration from user's home directory.
	home, err := homedir.Dir()
	if err == nil {
		files = append(files, file.NewFile(filepath.Join(home, fmt.Sprintf(".%s.%s", ProjectName, encoder.ExtensionsYaml))))
	}

	// A file placed in the current working directory takes precedence over the user-level configuration.
	pwd, err := os.Getwd()
	if err == nil {
		files = append(files, file.NewFile(filepath.Join(pwd, fmt.Sprintf("%s.%s", ProjectName, encoder.ExtensionsYaml))))
	}

	return files
}
