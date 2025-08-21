package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

// createExportCommand creates the 'export' command
func (c *OfflineBundleCLI) createExportCommand() *cobra.Command {
	var bundlePath, outputPath string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export an offline bundle to a zip file",
		Long:  "Export an offline bundle to a zip file for distribution",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load signing key
			privateKey, err := c.loadSigningKey()
			if err != nil {
				return fmt.Errorf("failed to load signing key: %w", err)
			}
			// Load bundle
			creator := bundle.NewOfflineBundleCreator(privateKey, bundle.Author{}, c.Output, c.AuditTrailManager)
			offlineBundle, err := creator.LoadOfflineBundle(bundlePath)
			if err != nil {
				return fmt.Errorf("failed to load offline bundle: %w", err)
			}

			// Export bundle
			err = creator.ExportOfflineBundle(offlineBundle, outputPath)
			if err != nil {
				return fmt.Errorf("failed to export offline bundle: %w", err)
			}

			fmt.Fprintf(c.Output, "Offline bundle exported successfully: %s\n", outputPath)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the offline bundle directory (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path for the exported bundle (required)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("bundle")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

// createIncrementalCommand creates the 'incremental' command
func (c *OfflineBundleCLI) createIncrementalCommand() *cobra.Command {
	var baseBundlePath, outputPath, newVersion, changesFile string

	cmd := &cobra.Command{
		Use:   "incremental",
		Short: "Create an incremental offline bundle",
		Long:  "Create an incremental offline bundle based on an existing bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load signing key
			privateKey, err := c.loadSigningKey()
			if err != nil {
				return fmt.Errorf("failed to load signing key: %w", err)
			}

			// Load base bundle
			creator := bundle.NewOfflineBundleCreator(privateKey, bundle.Author{}, c.Output, c.AuditTrailManager)
			baseBundle, err := creator.LoadOfflineBundle(baseBundlePath)
			if err != nil {
				return fmt.Errorf("failed to load base bundle: %w", err)
			}

			// Read changes file
			var changes []string
			if changesFile != "" {
				changesData, err := os.ReadFile(filepath.Clean(changesFile))
				if err != nil {
					return fmt.Errorf("failed to read changes file: %w", err)
				}

				// Parse changes (one per line)
				lines := strings.Split(string(changesData), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						changes = append(changes, line)
					}
				}
			} else {
				// Default change
				changes = []string{"Incremental update"}
			}

			// Create incremental bundle
			incrementalBundle, err := creator.CreateIncrementalBundle(baseBundle, newVersion, changes, outputPath)
			if err != nil {
				return fmt.Errorf("failed to create incremental bundle: %w", err)
			}

			fmt.Fprintf(c.Output, "Incremental bundle created successfully: %s\n", outputPath)
			fmt.Fprintf(c.Output, "Bundle version: %s\n", incrementalBundle.EnhancedManifest.Version)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&baseBundlePath, "base", "b", "", "Path to the base offline bundle directory (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path for the incremental bundle (required)")
	cmd.Flags().StringVarP(&newVersion, "version", "v", "", "New version for the incremental bundle (required)")
	cmd.Flags().StringVarP(&changesFile, "changes", "c", "", "Path to a file containing changes (one per line)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("base")
	_ = cmd.MarkFlagRequired("output")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}
