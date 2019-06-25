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

// Config represents the application-wide configurations.
type Config struct {
	// LogLevel is the verbosity level of the application-wide logging behavior.
	LogLevel string
}
