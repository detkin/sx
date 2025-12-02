package main

import (
	"fmt"
	"os"

	"github.com/sleuth-io/skills/internal/commands"
	"github.com/spf13/cobra"
)

var (
	// Version will be set via ldflags during build
	Version = "dev"
	// Commit will be set via ldflags during build
	Commit = "none"
	// Date will be set via ldflags during build
	Date = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "skills",
		Short: "Skills CLI - Provision AI artifacts from remote servers or Git repositories",
		Long: `Skills is a CLI tool that provisions AI artifacts (skills, agents, MCPs, etc.)
from remote Sleuth servers or Git repositories.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default command: run install if lock file exists
			return commands.RunDefaultCommand(cmd, args)
		},
		SilenceUsage: true,
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewInstallCommand())
	rootCmd.AddCommand(commands.NewAddCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
