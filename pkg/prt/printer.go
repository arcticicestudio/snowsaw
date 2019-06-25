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
	// FatalVerbosity is the level of the printer used to log any data with fatal scope.
	FatalVerbosity Verbosity = iota
	// ErrorVerbosity is the level of the printer used to log any data with error scope.
	ErrorVerbosity
	// WarnVerbosity is the level of the printer used to log any data with warn scope.
	WarnVerbosity
	// SuccessVerbosity is the level of the printer used to log any data with success scope.
	SuccessVerbosity
	// InfoVerbosity is the level of the printer used to log any data with info scope.
	// If no other level is explicitly set this level is used by default.
	InfoVerbosity
	// DebugVerbosity is the level of the printer used to log any data with debug scope.
	DebugVerbosity
)

var p = newPrinterConfig(InfoVerbosity)

// Debugf prints a debug scope message with a prefix symbol using the given format.
// A new line will be appended if not already included in the given format.
func Debugf(format string, args ...interface{}) { p.debugf(format, args...) }

// Errorf prints a error scope message with a prefix symbol using the given format.
// A new line will be appended if not already included in the given format.
func Errorf(format string, args ...interface{}) { p.errorf(format, args...) }

// Fatalf prints a fatal scope message with a prefix symbol using the given format.
// A new line will be appended if not already included in the given format.
func Fatalf(format string, args ...interface{}) { p.fatalf(format, args...) }

// Infof prints a info scope message with a prefix symbol using the given format.
// A new line will be appended if not already included in the given format.
func Infof(format string, args ...interface{}) { p.infof(format, args...) }

// Successf prints a success scope message with a prefix symbol using the given format.
// A new line will be appended if not already included in the given format.
func Successf(format string, args ...interface{}) { p.successf(format, args...) }

// Warnf prints a warning scope message with a prefix symbol using the given format.
// A new line will be appended if not already included in the given format.
func Warnf(format string, args ...interface{}) { p.warnf(format, args...) }

// SetVerbosityLevel sets the logging level of the printer.
func SetVerbosityLevel(v Verbosity) { p.setVerbosityLevel(v) }

// ParseVerbosityLevel takes a logging level name and returns the Verbosity log level constant.
func ParseVerbosityLevel(lvl string) (Verbosity, error) {
	switch strings.ToLower(lvl) {
	case "fatal":
		return FatalVerbosity, nil
	case "error":
		return ErrorVerbosity, nil
	case "warn":
		return WarnVerbosity, nil
	case "info":
		return InfoVerbosity, nil
	case "debug":
		return DebugVerbosity, nil
	}

	var v Verbosity
	return v, fmt.Errorf("not a valid printer level: %q", lvl)
}

// MarshalText returns the textual representation of itself.
func (v Verbosity) MarshalText() ([]byte, error) {
	switch v {
	case DebugVerbosity:
		return []byte("debug"), nil
	case InfoVerbosity:
		return []byte("info"), nil
	case WarnVerbosity:
		return []byte("warn"), nil
	case ErrorVerbosity:
		return []byte("error"), nil
	case FatalVerbosity:
		return []byte("fatal"), nil
	}

	return nil, fmt.Errorf("not a valid printer level %d", v)
}

// Convert the Verbosity to a string.
func (v Verbosity) String() string {
	if b, err := v.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}

// UnmarshalText implements encoding.TextUnmarshaler to unmarshal a textual representation of itself.
func (v *Verbosity) UnmarshalText(text []byte) error {
	l, err := ParseVerbosityLevel(string(text))
	if err != nil {
		return err
	}

	*v = Verbosity(l)
	return nil
}

func newPrinterConfig(v Verbosity) *printerConfig {
	return &printerConfig{verbosity: v}
}

func (p *printerConfig) debugf(format string, args ...interface{}) {
	prefix := color.New(color.FgMagenta, color.Bold).Sprint("Debug: ")
	p.withNewLine(DebugVerbosity, os.Stdout, prefix, format, args...)
}

func (p *printerConfig) errorf(format string, args ...interface{}) {
	prefix := color.New(color.FgRed, color.Bold).Sprint("✕ ")
	p.withNewLine(ErrorVerbosity, os.Stdout, prefix, format, args...)
}

func (p *printerConfig) fatalf(format string, args ...interface{}) {
	prefix := color.New(color.FgRed).Sprint("⭍ ")
	p.withNewLine(FatalVerbosity, os.Stdout, prefix, format, args...)
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
