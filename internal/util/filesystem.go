// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// AbsPath converts the given path to an absolute path.
func AbsPath(p string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}

	p, err := filepath.Abs(p)
	if err == nil {
		return filepath.Clean(p), nil
	}

	return "", fmt.Errorf("failed to convert to absolute path: %s", p)
}

// FileExists checks if the file at the given path exists and is not a directory.
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

// IsFileWritable checks if the given file is writable.
func IsFileWritable(path string) bool {
	_, err := os.OpenFile(path, os.O_WRONLY, 0660)
	if err != nil {
		return false
	}

	return true
}
