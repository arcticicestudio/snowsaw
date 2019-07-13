// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

package snowblock

// TaskRegistry is a registry for available task runner.
type TaskRegistry interface {
	// Add validates and adds the given task runner to the registry.
	Add(TaskRunner) error

	// GetAll returns a list of all currently registered task runners.
	GetAll() map[string]TaskRunner
}
