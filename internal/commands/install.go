package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInstallCommand creates the install command
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Read lock file, fetch artifacts, and install locally",
		Long: `Read the sleuth.lock file, fetch artifacts from the configured repository,
and install them to ~/.claude/ directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(cmd, args)
		},
	}

	return cmd
}

// runInstall executes the install command
func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Install command - To be implemented")
	return fmt.Errorf("not yet implemented")
}
