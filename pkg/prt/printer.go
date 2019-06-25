// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package prt provides functions to print any data to os.Stdout with a specific level for different purposes.
package prt

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Verbosity defines the detail level of the printer logging behavior.
type Verbosity uint32

type printerConfig struct {
	verbosity Verbosity
}

const (
	// ErrorVerbosity is the level of the printer used to log any data with error scope.
	ErrorVerbosity Verbosity = iota
	// WarnVerbosity is the level of the printer used to log any data with error scope.
	WarnVerbosity
	// InfoVerbosity is the level of the printer used to log any data with error scope.
	InfoVerbosity
	// SuccessVerbosity is the level of the printer used to log any data with error scope.
	SuccessVerbosity
	// DebugVerbosity is the level of the printer used to log any data with debug scope.
	DebugVerbosity
)

var p = newPrinterConfig(SuccessVerbosity)

// Debugf prints a prefixed debug message using the given format.
// A new line will be appended if not already included in the given format.
func Debugf(format string, args ...interface{}) { p.debugf(format, args...) }

// Errorf prints a prefixed error message using the given format.
// A new line will be appended if not already included in the given format.
func Errorf(format string, args ...interface{}) { p.errorf(format, args...) }

// Infof prints a prefixed info message using the given format.
// A new line will be appended if not already included in the given format.
func Infof(format string, args ...interface{}) { p.infof(format, args...) }

// Successf prints a prefixed success message using the given format.
// A new line will be appended if not already included in the given format.
func Successf(format string, args ...interface{}) { p.successf(format, args...) }

// Warnf prints a prefixed warning message using the given format.
// A new line will be appended if not already included in the given format.
func Warnf(format string, args ...interface{}) { p.warnf(format, args...) }

// SetVerbosityLevel sets the logging level of the printer.
func SetVerbosityLevel(v Verbosity) { p.setVerbosityLevel(v) }

func newPrinterConfig(v Verbosity) *printerConfig {
	return &printerConfig{verbosity: v}
}

func (p *printerConfig) debugf(format string, args ...interface{}) {
	prefix := color.New(color.FgHiMagenta, color.Bold).Sprint("Debug: ")
	p.withNewLine(DebugVerbosity, os.Stdout, prefix, format, args...)
}

func (p *printerConfig) errorf(format string, args ...interface{}) {
	prefix := color.New(color.FgRed, color.Bold).Sprint("Error: ")
	p.withNewLine(ErrorVerbosity, os.Stdout, prefix, format, args...)
}

func (p *printerConfig) infof(format string, args ...interface{}) {
	prefix := color.New(color.FgBlue).Sprint("➜ ")
	p.withNewLine(InfoVerbosity, os.Stdout, prefix, format, args...)
}

func (p *printerConfig) successf(format string, args ...interface{}) {
	prefix := color.New(color.FgGreen).Sprint("✓ ")
	p.withNewLine(SuccessVerbosity, os.Stdout, prefix, format, args...)
}

func (p *printerConfig) warnf(format string, args ...interface{}) {
	prefix := color.New(color.FgYellow, color.Bold).Sprint("! ")
	p.withNewLine(WarnVerbosity, os.Stdout, prefix, format, args...)
}

// isPrinterEnabled checks if the verbosity of the printer is greater than the verbosity param.
func (p *printerConfig) isPrinterEnabled(v Verbosity) bool { return p.verbosity >= v }

func (p *printerConfig) setVerbosityLevel(v Verbosity) { p.verbosity = v }

// withNewLine writes to the specified writer and appends a new line to the given format if not already.
func (p *printerConfig) withNewLine(v Verbosity, w io.Writer, prefix string, format string, args ...interface{}) {
	if p.isPrinterEnabled(v) {
		if !strings.HasSuffix(format, "\n") {
			format = format + "\n"
		}
		fmt.Fprintf(w, prefix+format, args...)
	}
}
