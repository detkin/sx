package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewAddCommand creates the add command
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [zip-file]",
		Short: "Add a local zip file artifact to the repository",
		Long: `Take a local zip file, detect metadata from its contents, prompt for
confirmation/edits, install it to the repository, and update the lock file.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var zipFile string
			if len(args) > 0 {
				zipFile = args[0]
			}
			return runAdd(cmd, zipFile)
		},
	}

	return cmd
}

// runAdd executes the add command
func runAdd(cmd *cobra.Command, zipFile string) error {
	fmt.Println("Add command - To be implemented")
	if zipFile != "" {
		fmt.Printf("Zip file: %s\n", zipFile)
	}
	return fmt.Errorf("not yet implemented")
}
