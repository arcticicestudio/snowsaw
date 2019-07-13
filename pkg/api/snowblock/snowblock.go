// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package snowblock provides the snowblock API.
//
// Note: API v0 is a legacy implementation to keep the compatibility with the original Python implementation and will be
// superseded with API v1 with snowsaw version 1.0.0!
package snowblock

// Snowblock represents a snowblock.
type Snowblock interface {
	// Dispatch handles the processing of the snowblock by dispatching the configured TaskConfiguration to a registered
	// TaskRunner that can handle it.
	Dispatch() error

	// Validate ensures the snowblock exists and the configuration is valid.
	// The taskRunner parameter is a map of registered task runners that are able to process the task instructions.
	Validate(taskRunner map[string]TaskRunner) error
}
