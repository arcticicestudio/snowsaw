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

// TaskRunner processes a specific task.
type TaskRunner interface {
	// GetTaskName returns the name of the task this runner can process.
	GetTaskName() string

	// Run processes a task using the given task instructions.
	// The snowblockPath parameter is the path of the snowblock used as contextual information.
	Run(instructions TaskConfiguration, snowblockPath string) error
}
