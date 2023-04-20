// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package shell provides a task runner implementation to run arbitrary shell commands.
package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"

	"github.com/arcticicestudio/snowsaw/pkg/api/snowblock"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
)

const (
	// CommandConfigArrayMaxArgs is the maximum amount of values that are allowed when using an a array of strings
	// as shell configuration type.
	CommandConfigArrayMaxArgs = 2
)

// Shell is a task runner to run arbitrary shell commands.
type Shell struct {
	cmd              string
	cmdArgs          []string
	config           *config
	snowblockAbsPath string
}

type config struct {
	Command     string `json:"command" yaml:"command"`
	Description string `json:"description" yaml:"description"`
	Stderr      bool   `json:"stderr" yaml:"stderr"`
	Stdin       bool   `json:"stdin" yaml:"stdin"`
	Stdout      bool   `json:"stdout" yaml:"stdout"`
}

// GetTaskName returns the name of the task this runner can process.
func (s Shell) GetTaskName() string {
	return "shell"
}

// Run processes a task using the given task instructions.
// The snowblockAbsPath parameter is the absolute path of the snowblock used as contextual information.
func (s *Shell) Run(configuration snowblock.TaskConfiguration, snowblockAbsPath string) error {
	s.snowblockAbsPath = snowblockAbsPath

	// Try to convert given task configurations...
	configMap, ok := configuration.([]interface{})
	if !ok {
		prt.Debugf("invalid shell configuration type: %s", color.RedString("%T", configuration))
		return errors.New("malformed shell configuration")
	}

	// ...and handle the possible types.
	for idxConfigMap, configData := range configMap {
		s.config = &config{}
		s.cmd = ""
		s.cmdArgs = []string{}

		switch configType := configData.(type) {
		// Handle JSON `object` configurations used to define a command with a description and additional options.
		case map[string]interface{}:
			if err := mapstructure.Decode(configType, &s.config); err != nil {
				return err
			}
			if execErr := s.execute(); execErr != nil {
				return execErr
			}

		// Handle JSON `string` configurations used to only specify a single command.
		case string:
			s.config.Command = configType
			if execErr := s.execute(); execErr != nil {
				return execErr
			}

		// Handle JSON `array` configurations storing `string` values used to specify a command with a description.
		case []interface{}:
			var configStringValues []string
			for idxConfigArray, value := range configType {
				configString, isStringValue := value.(string)
				if !isStringValue {
					prt.Debugf("Unsupported value in %s shell command configuration of type %s at index %s",
						color.CyanString("%d", idxConfigMap),
						color.RedString("%T", value),
						color.BlueString("%d", idxConfigArray))
					return fmt.Errorf("unsupported value in %d shell configuration at index %d: %v",
						idxConfigMap, idxConfigArray, value)
				}
				configStringValues = append(configStringValues, configString)
			}
			if len(configStringValues) > CommandConfigArrayMaxArgs || len(configStringValues) < CommandConfigArrayMaxArgs {
				return fmt.Errorf("invalid amount of shell command arguments, expected %d but got %d",
					CommandConfigArrayMaxArgs, len(configStringValues))
			}
			s.config.Command = configStringValues[0]
			s.config.Description = configStringValues[1]
			if execErr := s.execute(); execErr != nil {
				return execErr
			}

		// Reject invalid or unsupported JSON data structures.
		default:
			prt.Debugf("unsupported shell command configuration type: %s", color.RedString("%T", configType))
			return fmt.Errorf("unsupported shell command configuration at index %d", idxConfigMap)
		}
	}

	return nil
}

func (s *Shell) execute() error {
	s.cmd = os.Getenv("SHELL")
	s.cmdArgs = append(s.cmdArgs, "-c", s.config.Command)
	cmd := exec.Command(s.cmd, s.cmdArgs...)
	cmd.Dir = s.snowblockAbsPath
	cmd.Env = os.Environ()

	if s.config.Description != "" {
		prt.Infof(s.config.Description)
	}
	if s.config.Stderr {
		cmd.Stderr = os.Stderr
	}
	if s.config.Stdin {
		cmd.Stdin = os.Stdin
	}
	if s.config.Stdout {
		cmd.Stdout = os.Stdout
	}

	runErr := cmd.Run()
	if runErr != nil {
		return fmt.Errorf("failed to execute shell command: %s",
			color.CyanString("%s %s", s.cmd, strings.Join(s.cmdArgs, " ")))
	}

	return nil
}
