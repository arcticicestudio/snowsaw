// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

package task

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/arcticicestudio/snowsaw/pkg/api/snowblock"
)

// Registry is a registry for available task runner.
type Registry struct {
	runner map[string]snowblock.TaskRunner
}

// NewRegistry returns a new task runner registry instance.
func NewRegistry() *Registry {
	return &Registry{runner: make(map[string]snowblock.TaskRunner)}
}

// Add validates and adds the given task runner to the registry.
// If the name of the given task runner has already been registered and error is returned.
func (reg *Registry) Add(r snowblock.TaskRunner) error {
	_, exists := reg.runner[r.GetTaskName()]
	if exists {
		return fmt.Errorf("runner for task name already exists: %s", color.CyanString(r.GetTaskName()))
	}
	reg.runner[r.GetTaskName()] = r
	return nil
}

// GetAll returns a list of all currently registered task runner.
func (reg *Registry) GetAll() map[string]snowblock.TaskRunner {
	return reg.runner
}
