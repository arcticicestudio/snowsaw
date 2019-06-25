// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package file provides a struct and methods to handle configuration files.
package file

import (
	"io/ioutil"
	"os"

	"github.com/arcticicestudio/snowsaw/pkg/config/encoder"
)

// File represents a configuration file.
type File struct {
	Data    []byte
	Encoder encoder.Encoder
	Path    string
}

// NewFile returns a new File.
func NewFile(path string) *File {
	return &File{Path: path}
}

// WithEncoder sets the encoder for the File.
func (f *File) WithEncoder(e encoder.Encoder) *File {
	f.Encoder = e
	return f
}

// Read tries to read in the File data.
// If any error occurs while trying to read the file at the specified path,
// nil is returned for the File along with the error.
func (f *File) Read() (*File, error) {
	fh, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	b, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}
	_, err = fh.Stat()
	if err != nil {
		return nil, err
	}

	return &File{Data: b}, nil
}
