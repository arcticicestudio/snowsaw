// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package config contains application-wide configurations and constants.
// It provides a file abstraction to de/encode, load and validate YAML and JSON data using the builder design pattern.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/arcticicestudio/snowsaw/pkg/config/encoder"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
)

// Config represents the application-wide configurations.
type Config struct {
	// LogLevel is the application-wide logging verbosity level.
	LogLevel string `yaml:"logLevel"`

	// Snowblocks are general snowblocks configurations.
	Snowblocks Snowblocks `yaml:"snowblocks,flow"`
}

// Snowblocks represents the general snowblocks configurations.
type Snowblocks struct {
	// BaseDirs are the paths of the snowblock base directories.
	BaseDirs []string `yaml:"baseDirs,flow"`
	// Paths are the paths of the snowblocks directories.
	Paths []string `yaml:"paths,flow"`
}

func init() {
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
