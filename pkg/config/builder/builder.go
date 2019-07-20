// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package builder provides methods to load and merge configuration files using the builder design pattern.
package builder

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/imdario/mergo"

	"github.com/arcticicestudio/snowsaw/pkg/config"
	"github.com/arcticicestudio/snowsaw/pkg/config/encoder"
	"github.com/arcticicestudio/snowsaw/pkg/config/encoder/yaml"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
	"github.com/arcticicestudio/snowsaw/pkg/util/filesystem"
)

// Builder contains the current configuration building state.
type Builder struct {
	Files []*file.File
}

// Load tries to load all given configuration files.
// It checks if the path is valid and exists, tries to assign a matching encoder based on the file extension and returns
// a pointer to a builder to chain the merge function.
func Load(files ...*file.File) *Builder {
	b := &Builder{Files: []*file.File{}}

	for _, f := range files {
		// Convert to an absolute path and check if the file exists, otherwise ignore and check next.
		absPath, absPathErr := filepath.Abs(f.Path)
		if absPathErr != nil {
			prt.Debugf("Could not convert to absolute configuration file path: %v", absPathErr)
			continue
		}
		if exists, _ := filesystem.FileExists(absPath); !exists {
			prt.Debugf("Ignoring non-existent configuration file: %s", color.CyanString(f.Path))
			continue
		}
		f.Path = absPath

		fileExt := filepath.Ext(f.Path)
		// Check if the file matches the supported YAML extension...
		if len(fileExt) <= 1 || fileExt[1:] != encoder.ExtensionsYaml {
			prt.Debugf("Ignoring configuration file without supported extension: %s", color.CyanString(f.Path))
			continue
		}
		// ...when trimming the dot character that separates the file name and extension.
		if fileExt[1:] == encoder.ExtensionsYaml {
			f.Encoder = yaml.NewYamlEncoder()
			b.Files = append(b.Files, f)
		}
	}

	return b
}

// Into accepts a configuration struct pointer and populates it with the current config state.
// The order of the files are maintained when merging of the configuration states is enabled.
func (b *Builder) Into(c *config.Config, merge bool) error {
	for _, f := range b.Files {
		content, err := ioutil.ReadFile(f.Path)
		if err != nil {
			return err
		}

		if !merge {
			// Decode the file content into the given configuration state using the assigned encoder...
			if encErr := f.Encoder.Decode(content, &c); encErr != nil {
				return fmt.Errorf("%s: %v", f.Path, encErr)
			}
			continue
		}

		newState := &config.Config{}
		// ...or merge into the given encoded configuration state.
		if encErr := f.Encoder.Decode(content, &newState); encErr != nil {
			return fmt.Errorf("%s: %v", f.Path, encErr)
		}
		if encErr := mergo.Merge(c, newState, mergo.WithAppendSlice, mergo.WithOverride); encErr != nil {
			return fmt.Errorf("%s: %v", f.Path, encErr)
		}
	}

	return nil
}
