// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package snowblock provides the implementation of the snowblock API.
package snowblock

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/fatih/color"

	api "github.com/arcticicestudio/snowsaw/pkg/api/snowblock"
	"github.com/arcticicestudio/snowsaw/pkg/config/encoder"
	"github.com/arcticicestudio/snowsaw/pkg/config/encoder/json"
	"github.com/arcticicestudio/snowsaw/pkg/config/source/file"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
	"github.com/arcticicestudio/snowsaw/pkg/util/filesystem"
)

// Configuration i
type Configuration []map[string]interface{}

// Snowblock represents the state and actions of a snowblock.
// It implements the snowblock API interface.
type Snowblock struct {
	// IsValid indicates if this snowblock is valid.
	IsValid bool

	// Path is the path of this snowblock.
	Path string

	// TaskObjects is a list of task objects this snowblock configuration consists of.
	TaskObjects []api.Task

	// TaskRunnerMapping contains the assignments from task objects to a matching task runner.
	TaskRunnerMapping map[api.TaskRunner]api.TaskConfiguration

	// UnsupportedTasks is a list of task names that are not supported by an registered task runner.
	UnsupportedTasks []api.TaskConfiguration
}

// NewSnowblock returns a new snowblock.
func NewSnowblock(path string) *Snowblock {
	return &Snowblock{
		Path:              path,
		TaskObjects:       make([]api.Task, 0),
		TaskRunnerMapping: make(map[api.TaskRunner]api.TaskConfiguration),
	}
}

// Dispatch handles the processing of the snowblock by dispatching the configured tasks to a registered runner that can
// handle it.
func (s *Snowblock) Dispatch() error {
	for runner, instructions := range s.TaskRunnerMapping {
		if err := runner.Run(instructions, s.Path); err != nil {
			return err
		}
	}

	return nil
}

// Validate ensures the snowblock and the configuration file exist.
// It returns false along with an corresponding error if the snowblock path doesn't exist or is not a directory.
// The snowblock is also not valid either when the given directory does not contain a configuration file or the file
// is not parsable.
// Defined tasks that are not supported by any of the given task runners are filtered into the separate array.
func (s *Snowblock) Validate(taskRunner map[string]api.TaskRunner) error {
	// Expand the given snowblock path and convert it into an absolute path.
	expandedPath, expPathErr := filesystem.ExpandPath(s.Path)
	if expPathErr != nil {
		return fmt.Errorf("could not dissolve path: %v", expPathErr)
	}
	expandedAbsPath, expAbsPathErr := filepath.Abs(expandedPath)
	if expAbsPathErr != nil {
		return fmt.Errorf("could not convert into absolute path: %v", expAbsPathErr)
	}

	// Ensure the path is a directory and exists.
	dirExists, dirExistsErr := filesystem.DirExists(expandedAbsPath)
	if dirExistsErr != nil {
		return dirExistsErr
	}
	if !dirExists {
		return fmt.Errorf("no such directory: %s", expandedAbsPath)
	}
	s.Path = expandedAbsPath

	// Try to read and encode the task objects when the directory contains a configuration file.
	configFilePath := filepath.Join(s.Path, fmt.Sprintf("%s.%s", api.ConfigurationFileName, encoder.ExtensionsJson))
	if configLoadErr := loadConfigFile(configFilePath, &s.TaskObjects); configLoadErr != nil {
		prt.Debugf("Ignoring snowblock directory %s without valid configuration file: %s: %v",
			color.CyanString(filepath.Base(s.Path)), color.BlueString(configFilePath), configLoadErr)
		return nil
	}

	// Assign each task object to a registered task runner when matching, otherwise add to list of unsupported tasks.
	for _, taskObject := range s.TaskObjects {
		for taskName, taskConfigMap := range taskObject {
			runner, exists := taskRunner[taskName]
			if exists {
				s.TaskRunnerMapping[runner] = taskConfigMap
				continue
			}
			s.UnsupportedTasks = append(s.UnsupportedTasks, taskName)
			prt.Debugf("Ignoring task without registered runner: %s", color.RedString(taskName))
		}
	}

	s.IsValid = true
	return nil
}

func loadConfigFile(absPath string, tasks *[]api.Task) error {
	f := file.NewFile(absPath).WithEncoder(json.NewJsonEncoder())
	// Check if the file exists...
	if exists, _ := filesystem.FileExists(f.Path); !exists {
		return fmt.Errorf("no such snowblock configuration file: %s", color.RedString(f.Path))
	}
	// ...and decode the file content with the given tasks using the assigned encoder.
	content, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return err
	}
	if encErr := f.Encoder.Decode(content, tasks); encErr != nil {
		return fmt.Errorf("could not load snowblock configuration file: %s: %v", f.Path, encErr)
	}

	return nil
}
