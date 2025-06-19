package cmd

import (
	"github.com/spf13/cobra"
)

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Bundle management commands",
	Long: `Manage offline update bundles for the LLMreconing Tool.

This command provides subcommands for:
- Creating bundles with components and templates
- Verifying bundle integrity and compliance
- Importing bundles into the system
- Viewing bundle information

Bundles support OWASP LLM Top 10 categorization and compliance documentation.`,
	Example: `  # Create a bundle with all components
  LLMrecon bundle create --output=update.bundle
  
  # Verify a bundle
  LLMrecon bundle verify update.bundle
  
  # Import a bundle
  LLMrecon bundle import update.bundle --backup
  
  # View bundle information
  LLMrecon bundle info update.bundle`,
}

func init() {
	rootCmd.AddCommand(bundleCmd)
}