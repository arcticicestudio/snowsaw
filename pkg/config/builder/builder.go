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
	"reflect"

	"github.com/fatih/color"
	"github.com/imdario/mergo"

	"github.com/arcticicestudio/snowsaw/pkg/config/encoder"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
	"github.com/arcticicestudio/snowsaw/pkg/util/filesystem"
)

// builder contains the current configuration building state.
type builder struct {
	Files []*file.File
}

// Load tries to load all given configuration files.
// It checks if the path is valid and exists, tries to assign a matching encoder.Encoder based on the file extension and
// returns a pointer to a builder to chain and pass the loaded files to the Merge function.
func Load(files ...*file.File) *builder {
	s := &builder{Files: []*file.File{}}

	for _, f := range files {
		// Convert to absolute path and check if file exists, otherwise ignore and check next.
		f.Path, _ = filepath.Abs(f.Path)
		if exists, _ := filesystem.FileExists(f.Path); !exists {
			prt.Debugf("Ignoring non-existent configuration file: %s", color.CyanString(f.Path))
			continue
		}

		// Find matching encoder by file extension if not already set.
		if f.Encoder == nil {
			fileExt := filepath.Ext(f.Path)
			if len(fileExt) <= 1 {
				prt.Debugf("Ignoring configuration file without supported extension: %s", color.CyanString(f.Path))
				continue
			}

			// Strip dot character separating the file name and extension.
			fileExt = fileExt[1:]

			// Only add files with supported encoders.
			for ext, enc := range encoder.ExtensionMapping {
				if ext == fileExt {
					f.Encoder = enc
					s.Files = append(s.Files, f)
					break
				}
			}
		} else {
			s.Files = append(s.Files, f)
		}
	}

	return s
}

// Into accepts a configuration struct pointer and populates it with the current config state.
// The order of the files array is maintained when merging the configuration states into the struct is enabled.
func (s *builder) Into(c interface{}, merge bool) error {
	base := reflect.New(reflect.TypeOf(c).Elem()).Interface()

	for _, f := range s.Files {
		content, err := ioutil.ReadFile(f.Path)
		if err != nil {
			return err
		}

		if !merge {
			// Decode the file content into the given base configuration state using the assigned encoder...
			if encErr := f.Encoder.Decode(content, &c); encErr != nil {
				return fmt.Errorf("%s: %v", f.Path, encErr)
			}
			continue
		}

		// ...or merge into the given base configuration state.
		raw := base
		if encErr := f.Encoder.Decode(content, &raw); encErr != nil {
			return fmt.Errorf("%s: %v", f.Path, encErr)
		}
		if encErr := mergo.Merge(c, raw, mergo.WithAppendSlice, mergo.WithOverride); encErr != nil {
			return fmt.Errorf("%s: %v", f.Path, encErr)
		}
	}

	return nil
}
