// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package filesystem provides utility functions related to filesystem actions.
package filesystem

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
)

// DirExists checks if the directory at the path exists and is not a file.
// Returns true if the given path exists and is a directory, false otherwise.
// If an error occurs, false is returned along with the error.
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s is a file", path)
	}

	return false, err
}

// ExpandPath expands environment variables and special elements like the tilde character for the given path.
// If an error occurs the original passed path is returned along with the corresponding error.
func ExpandPath(path string) (string, error) {
	// Handle special case for Unix tilde character expansion.
	expandedUserHomePath, err := homedir.Expand(path)
	if err != nil {
		return path, err
	}

	return os.ExpandEnv(expandedUserHomePath), nil
}

// FileExists checks if the file at the given path exists and is not a directory.
// If an error occurs, false is returned along with the corresponding error.
func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	if info.IsDir() {
		return false, fmt.Errorf("%s is a directory", path)
	}

	return false, err
}

// IsFileWritable checks if the file at the given path is writable.
// If an error occurs, false is returned along with the corresponding error.
func IsFileWritable(path string) (bool, error) {
	_, err := os.OpenFile(path, os.O_WRONLY, 0660)
	if err != nil {
		return false, err
	}

	return true, nil
}

// IsSymlink checks if the specified path is a symbolic link.
// If an error occurs, false is returned along with the corresponding error.
func IsSymlink(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	return fi.Mode()&os.ModeSymlink == os.ModeSymlink, nil
}

// NodeExists checks if the node at the given path exists.
// If an error occurs, false is returned along with the corresponding error.
func NodeExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
