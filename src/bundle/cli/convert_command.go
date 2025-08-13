package cli

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

// createConvertCommand creates the 'convert' command
func (c *OfflineBundleCLI) createConvertCommand() *cobra.Command {
	var bundlePath, outputPath, autoDetectCompliance string

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert a standard bundle to an offline bundle",
		Long:  "Convert a standard bundle to an offline bundle with enhanced features",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load signing key
			privateKey, err := c.loadSigningKey()
			if err != nil {
				return fmt.Errorf("failed to load signing key: %w", err)
			}

			// Load standard bundle
			standardBundle, err := bundle.LoadBundle(bundlePath)
			if err != nil {
				return fmt.Errorf("failed to load standard bundle: %w", err)
			}

			// Create offline bundle creator
			creator := bundle.NewOfflineBundleCreator(privateKey, standardBundle.Manifest.Author, c.Output, c.AuditTrailManager)

			// Create converter
			converter := bundle.NewBundleConverter(creator, c.Output, c.AuditTrailManager)

			// Convert bundle
			offlineBundle, err := converter.ConvertToOfflineBundle(standardBundle, outputPath)
			if err != nil {
				return fmt.Errorf("failed to convert bundle: %w", err)
			}

			// Auto-detect compliance if requested
			if autoDetectCompliance == "true" || autoDetectCompliance == "yes" {
				fmt.Fprintf(c.Output, "Auto-detecting compliance mappings for templates...\n")
				err = converter.AutoDetectComplianceForTemplates(offlineBundle)
				if err != nil {
					return fmt.Errorf("failed to auto-detect compliance mappings: %w", err)
				}
			}

			fmt.Fprintf(c.Output, "Bundle converted successfully to offline format: %s\n", outputPath)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the standard bundle directory (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path for the offline bundle (required)")
	cmd.Flags().StringVarP(&autoDetectCompliance, "auto-detect-compliance", "a", "false", "Auto-detect compliance mappings for templates (true/false)")

	// Mark required flags
	cmd.MarkFlagRequired("bundle")
	cmd.MarkFlagRequired("output")

	return cmd
}
