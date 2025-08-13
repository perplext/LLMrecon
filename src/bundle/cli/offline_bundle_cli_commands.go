package cli

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

// createValidateCommand creates the 'validate' command
func (c *OfflineBundleCLI) createValidateCommand() *cobra.Command {
	var bundlePath, level string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an offline bundle",
		Long:  "Validate an offline bundle with the specified validation level",
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

			// Determine validation level
			var validationLevel bundle.ValidationLevel
			switch strings.ToLower(level) {
			case "basic":
				validationLevel = bundle.BasicValidation
			case "standard":
				validationLevel = bundle.StandardValidation
			case "strict":
				validationLevel = bundle.StrictValidation
			default:
				return fmt.Errorf("invalid validation level: %s (must be 'basic', 'standard', or 'strict')", level)
			}

			// Validate bundle
			result, err := creator.ValidateOfflineBundle(offlineBundle, validationLevel)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			// Print validation result
			fmt.Fprintf(c.Output, "Validation result: %s\n", result.Message)
			fmt.Fprintf(c.Output, "Valid: %t\n", result.Valid)
			fmt.Fprintf(c.Output, "Level: %s\n", result.Level)

			if len(result.Errors) > 0 {
				fmt.Fprintf(c.Output, "\nErrors:\n")
				for _, err := range result.Errors {
					fmt.Fprintf(c.Output, "- %s\n", err)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Fprintf(c.Output, "\nWarnings:\n")
				for _, warning := range result.Warnings {
					fmt.Fprintf(c.Output, "- %s\n", warning)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the offline bundle directory (required)")
	cmd.Flags().StringVarP(&level, "level", "l", "standard", "Validation level (basic, standard, or strict)")

	// Mark required flags
	cmd.MarkFlagRequired("bundle")

	return cmd
}

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
	cmd.MarkFlagRequired("bundle")
	cmd.MarkFlagRequired("output")

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
				changesData, err := os.ReadFile(changesFile)
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
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&baseBundlePath, "base", "b", "", "Path to the base offline bundle directory (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path for the incremental bundle (required)")
	cmd.Flags().StringVarP(&newVersion, "version", "v", "", "New version for the incremental bundle (required)")
	cmd.Flags().StringVarP(&changesFile, "changes", "c", "", "Path to a file containing changes (one per line)")

	// Mark required flags
	cmd.MarkFlagRequired("base")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagRequired("version")

	return cmd
}

// createKeygenCommand creates the 'keygen' command
func (c *OfflineBundleCLI) createKeygenCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate a signing key pair",
		Long:  "Generate a new Ed25519 signing key pair for offline bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate key pair
			publicKey, privateKey, err := bundle.GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair: %w", err)
			}

			// Encode keys
			privateKeyEncoded := base64.StdEncoding.EncodeToString(privateKey)
			publicKeyEncoded := base64.StdEncoding.EncodeToString(publicKey)

			// Write keys to files
			privateKeyPath := filepath.Join(outputPath, "offline_bundle_private.key")
			publicKeyPath := filepath.Join(outputPath, "offline_bundle_public.key")

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Write private key
			if err := os.WriteFile(privateKeyPath, []byte(privateKeyEncoded), 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}

			// Write public key
			if err := os.WriteFile(publicKeyPath, []byte(publicKeyEncoded), 0644); err != nil {
				return fmt.Errorf("failed to write public key: %w", err)
			}

			fmt.Fprintf(c.Output, "Key pair generated successfully:\n")
			fmt.Fprintf(c.Output, "Private key: %s\n", privateKeyPath)
			fmt.Fprintf(c.Output, "Public key: %s\n", publicKeyPath)
			fmt.Fprintf(c.Output, "\nIMPORTANT: Keep the private key secure and back it up safely.\n")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Output directory for the key files")

	return cmd
}
