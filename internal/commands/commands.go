package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RunDefaultCommand runs the default command (install if lock file exists)
func RunDefaultCommand(cmd *cobra.Command, args []string) error {
	// Check if sleuth.lock exists in current directory
	if _, err := os.Stat("sleuth.lock"); err == nil {
		// Lock file exists, run install
		return runInstall(cmd, args)
	}

	// No lock file, show help
	fmt.Fprintln(os.Stderr, "No sleuth.lock file found in current directory.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "To get started:")
	fmt.Fprintln(os.Stderr, "  1. Run 'skills init' to configure a repository")
	fmt.Fprintln(os.Stderr, "  2. Run 'skills install' to install artifacts from the lock file")
	fmt.Fprintln(os.Stderr, "")
	return cmd.Help()
}
