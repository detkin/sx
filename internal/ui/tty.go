// Package ui provides terminal UI components for the Skills CLI.
package ui

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IsTTY returns true if the given writer is a terminal.
func IsTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

// IsStdoutTTY returns true if stdout is a terminal.
func IsStdoutTTY() bool {
	return IsTTY(os.Stdout)
}

// IsStdinTTY returns true if stdin is a terminal.
func IsStdinTTY() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// NoColor returns true if color output should be disabled.
// Respects the NO_COLOR environment variable standard.
func NoColor() bool {
	_, exists := os.LookupEnv("NO_COLOR")
	return exists
}
