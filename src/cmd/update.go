package cmd

import (
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update management commands",
	Long: `Manage updates for the LLMreconing Tool and its components.

This command provides subcommands for:
- Checking for available updates
- Applying updates
- Viewing version information
- Displaying changelogs

Use 'LLMrecon update <subcommand> --help' for more information about each subcommand.`,
	Example: `  # Check for updates
  LLMrecon update check
  
  # Apply all available updates
  LLMrecon update apply
  
  # Check version information
  LLMrecon version --verbose`,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}