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

	// Include user-level dot-file configurations from the user's home directory.
	home, err := homedir.Dir()
	if err == nil {
		for _, ext := range encoder.ExtensionsJson {
			files = append(files, file.NewFile(filepath.Join(home, fmt.Sprintf(".%s.%s", ProjectName, ext))))
		}
		// Since YAML is a superset of JSON, YAML files take precedence over pure JSON based configurations.
		for _, ext := range encoder.ExtensionsYaml {
			files = append(files, file.NewFile(filepath.Join(home, fmt.Sprintf(".%s.%s", ProjectName, ext))))
		}
	}

	// Files placed in the current working directory take precedence over user-level configurations.
	pwd, err := os.Getwd()
	if err == nil {
		for _, ext := range encoder.ExtensionsJson {
			files = append(files, file.NewFile(filepath.Join(pwd, fmt.Sprintf("%s.%s", ProjectName, ext))))
		}
		// Since YAML is a superset of JSON, YAML files take precedence over pure JSON based configurations.
		for _, ext := range encoder.ExtensionsYaml {
			files = append(files, file.NewFile(filepath.Join(pwd, fmt.Sprintf("%s.%s", ProjectName, ext))))
		}
	}

	return files
}
