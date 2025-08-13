// Package cli provides command-line interfaces for bundle operations
package cli

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
	"github.com/spf13/cobra"
)

// OfflineBundleCLI provides a command-line interface for offline bundle operations
type OfflineBundleCLI struct {
	// Creator is the offline bundle creator
	Creator *bundle.OfflineBundleCreator
	// RootCmd is the root command for the CLI
	RootCmd *cobra.Command
	// Output is the output writer
	Output io.Writer
	// KeyPath is the path to the signing key
	KeyPath string
	// AuditTrailManager is the audit trail manager
	AuditTrailManager *trail.AuditTrailManager
}

// NewOfflineBundleCLI creates a new offline bundle CLI
func NewOfflineBundleCLI(output io.Writer, auditTrailManager *trail.AuditTrailManager) *OfflineBundleCLI {
	if output == nil {
		output = os.Stdout
	}

	cli := &OfflineBundleCLI{
		Output:           output,
		AuditTrailManager: auditTrailManager,
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "offline-bundle",
		Short: "Manage offline bundles",
		Long:  "Create, validate, and manage offline bundles for LLM red teaming",
	}

	// Add commands
	rootCmd.AddCommand(cli.createCreateCommand())
	rootCmd.AddCommand(cli.createAddContentCommand())
	rootCmd.AddCommand(cli.createAddComplianceCommand())
	rootCmd.AddCommand(cli.createAddDocumentationCommand())
	rootCmd.AddCommand(cli.createValidateCommand())
	rootCmd.AddCommand(cli.createExportCommand())
	rootCmd.AddCommand(cli.createIncrementalCommand())
	rootCmd.AddCommand(cli.createKeygenCommand())
	rootCmd.AddCommand(cli.createConvertCommand())

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&cli.KeyPath, "key", "k", "", "Path to signing key file")

	cli.RootCmd = rootCmd

	return cli
}

// Execute executes the root command
func (c *OfflineBundleCLI) Execute() error {
	return c.RootCmd.Execute()
}

// createCreateCommand creates the 'create' command
func (c *OfflineBundleCLI) createCreateCommand() *cobra.Command {
	var name, description, version, bundleType, outputPath, authorName, authorEmail, authorOrg string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new offline bundle",
		Long:  "Create a new offline bundle with the specified parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load signing key
			privateKey, err := c.loadSigningKey()
			if err != nil {
				return fmt.Errorf("failed to load signing key: %w", err)
			}

			// Create author
			author := bundle.Author{
				Name:         authorName,
				Email:        authorEmail,
				Organization: authorOrg,
			}

			// Create creator
			creator := bundle.NewOfflineBundleCreator(privateKey, author, c.Output, c.AuditTrailManager)

			// Validate bundle type
			var bundleTypeEnum bundle.BundleType
			switch strings.ToLower(bundleType) {
			case "template":
				bundleTypeEnum = bundle.TemplateBundleType
			case "module":
				bundleTypeEnum = bundle.ModuleBundleType
			case "mixed":
				bundleTypeEnum = bundle.MixedBundleType
			default:
				return fmt.Errorf("invalid bundle type: %s (must be 'template', 'module', or 'mixed')", bundleType)
			}

			// Create bundle
			_, err = creator.CreateOfflineBundle(name, description, version, bundleTypeEnum, outputPath)
			if err != nil {
				return fmt.Errorf("failed to create offline bundle: %w", err)
			}

			fmt.Fprintf(c.Output, "Offline bundle created successfully: %s\n", outputPath)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&name, "name", "n", "", "Bundle name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Bundle description")
	cmd.Flags().StringVarP(&version, "version", "v", "1.0.0", "Bundle version")
	cmd.Flags().StringVarP(&bundleType, "type", "t", "template", "Bundle type (template, module, or mixed)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path (required)")
	cmd.Flags().StringVarP(&authorName, "author-name", "", "", "Author name")
	cmd.Flags().StringVarP(&authorEmail, "author-email", "", "", "Author email")
	cmd.Flags().StringVarP(&authorOrg, "author-org", "", "", "Author organization")

	// Mark required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("output")

	return cmd
}

// createAddContentCommand creates the 'add-content' command
func (c *OfflineBundleCLI) createAddContentCommand() *cobra.Command {
	var bundlePath, sourcePath, targetPath, contentType, id, version, description string

	cmd := &cobra.Command{
		Use:   "add-content",
		Short: "Add content to an offline bundle",
		Long:  "Add content to an offline bundle with the specified parameters",
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

			// Validate content type
			var contentTypeEnum bundle.ContentType
			switch strings.ToLower(contentType) {
			case "template":
				contentTypeEnum = bundle.TemplateContentType
			case "module":
				contentTypeEnum = bundle.ModuleContentType
			case "config":
				contentTypeEnum = bundle.ConfigContentType
			case "resource":
				contentTypeEnum = bundle.ResourceContentType
			default:
				return fmt.Errorf("invalid content type: %s (must be 'template', 'module', 'config', or 'resource')", contentType)
			}

			// Add content
			err = creator.AddContentToOfflineBundle(offlineBundle, sourcePath, targetPath, contentTypeEnum, id, version, description)
			if err != nil {
				return fmt.Errorf("failed to add content to offline bundle: %w", err)
			}

			fmt.Fprintf(c.Output, "Content added successfully to offline bundle\n")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the offline bundle directory (required)")
	cmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Path to the source file (required)")
	cmd.Flags().StringVarP(&targetPath, "target", "t", "", "Target path within the bundle (required)")
	cmd.Flags().StringVarP(&contentType, "type", "c", "", "Content type (template, module, config, or resource) (required)")
	cmd.Flags().StringVarP(&id, "id", "i", "", "Content ID (optional, will be generated if not provided)")
	cmd.Flags().StringVarP(&version, "version", "v", "1.0.0", "Content version")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Content description")

	// Mark required flags
	cmd.MarkFlagRequired("bundle")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("type")

	return cmd
}

// createAddComplianceCommand creates the 'add-compliance' command
func (c *OfflineBundleCLI) createAddComplianceCommand() *cobra.Command {
	var bundlePath, contentID, owaspCategories, isoControls string

	cmd := &cobra.Command{
		Use:   "add-compliance",
		Short: "Add compliance mappings to an offline bundle",
		Long:  "Add compliance mappings to an offline bundle with the specified parameters",
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

			// Parse categories and controls
			owaspCategoriesList := strings.Split(owaspCategories, ",")
			isoControlsList := strings.Split(isoControls, ",")

			// Clean empty entries
			cleanOwaspCategories := []string{}
			for _, category := range owaspCategoriesList {
				category = strings.TrimSpace(category)
				if category != "" {
					cleanOwaspCategories = append(cleanOwaspCategories, category)
				}
			}

			cleanISOControls := []string{}
			for _, control := range isoControlsList {
				control = strings.TrimSpace(control)
				if control != "" {
					cleanISOControls = append(cleanISOControls, control)
				}
			}

			// Add compliance mappings
			err = creator.AddComplianceMappingToOfflineBundle(offlineBundle, contentID, cleanOwaspCategories, cleanISOControls)
			if err != nil {
				return fmt.Errorf("failed to add compliance mappings to offline bundle: %w", err)
			}

			fmt.Fprintf(c.Output, "Compliance mappings added successfully to offline bundle\n")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the offline bundle directory (required)")
	cmd.Flags().StringVarP(&contentID, "content-id", "i", "", "Content ID to map (required)")
	cmd.Flags().StringVarP(&owaspCategories, "owasp", "o", "", "OWASP LLM Top 10 categories (comma-separated)")
	cmd.Flags().StringVarP(&isoControls, "iso", "s", "", "ISO/IEC 42001 controls (comma-separated)")

	// Mark required flags
	cmd.MarkFlagRequired("bundle")
	cmd.MarkFlagRequired("content-id")

	return cmd
}

// createAddDocumentationCommand creates the 'add-documentation' command
func (c *OfflineBundleCLI) createAddDocumentationCommand() *cobra.Command {
	var bundlePath, docType, sourcePath string

	cmd := &cobra.Command{
		Use:   "add-documentation",
		Short: "Add documentation to an offline bundle",
		Long:  "Add documentation to an offline bundle with the specified parameters",
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

			// Add documentation
			err = creator.AddDocumentationToOfflineBundle(offlineBundle, docType, sourcePath)
			if err != nil {
				return fmt.Errorf("failed to add documentation to offline bundle: %w", err)
			}

			fmt.Fprintf(c.Output, "Documentation added successfully to offline bundle\n")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the offline bundle directory (required)")
	cmd.Flags().StringVarP(&docType, "type", "t", "", "Documentation type (e.g., 'usage', 'installation') (required)")
	cmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Path to the documentation file (required)")

	// Mark required flags
	cmd.MarkFlagRequired("bundle")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("source")

	return cmd
}

// loadSigningKey loads the signing key from the specified path
func (c *OfflineBundleCLI) loadSigningKey() (ed25519.PrivateKey, error) {
	if c.KeyPath == "" {
		return nil, fmt.Errorf("signing key path not specified")
	}

	// Read the key file
	keyData, err := os.ReadFile(c.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signing key: %w", err)
	}

	// Try to parse as PEM
	block, _ := pem.Decode(keyData)
	if block != nil && block.Type == "PRIVATE KEY" {
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		// Check if the key is an Ed25519 key
		edKey, ok := privKey.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not an Ed25519 key")
		}

		return edKey, nil
	}

	// If not PEM, try to parse as raw Ed25519 key
	if len(keyData) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key size")
	}

	return ed25519.PrivateKey(keyData), nil
}

// createValidateCommand creates the 'validate' command for validating bundles
func (c *OfflineBundleCLI) createValidateCommand() *cobra.Command {
	var bundlePath, publicKeyPath string
	var level string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an offline bundle",
		Long:  "Validate an offline bundle's integrity and authenticity",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine validation level
			var validationLevel bundle.ValidationLevel
			switch strings.ToLower(level) {
			case "basic":
				validationLevel = bundle.BasicValidation
			case "standard":
				validationLevel = bundle.StandardValidation
			case "strict":
				validationLevel = bundle.StrictValidation
			case "manifest":
				validationLevel = bundle.ManifestValidationLevel
			case "checksum":
				validationLevel = bundle.ChecksumValidationLevel
			case "signature":
				validationLevel = bundle.SignatureValidationLevel
			case "compatibility":
				validationLevel = bundle.CompatibilityValidationLevel
			default:
				return fmt.Errorf("invalid validation level: %s", level)
			}

			// Load bundle
			fmt.Fprintf(c.Output, "Loading bundle from %s...\n", bundlePath)
			offlineBundle, err := bundle.LoadBundle(bundlePath)
			if err != nil {
				return fmt.Errorf("failed to load bundle: %w", err)
			}

			// Load public key if provided
			var publicKey []byte
			if publicKeyPath != "" {
				fmt.Fprintf(c.Output, "Loading public key from %s...\n", publicKeyPath)
				publicKey, err = os.ReadFile(publicKeyPath)
				if err != nil {
					return fmt.Errorf("failed to read public key: %w", err)
				}
			}

			// Validate bundle
			fmt.Fprintf(c.Output, "Validating bundle with level: %s...\n", level)
			var result *bundle.ValidationResult

			if validationLevel == bundle.SignatureValidationLevel && publicKey != nil {
				// Validate signature
				result, err = bundle.VerifyBundle(offlineBundle, publicKey)
				if err != nil {
					return fmt.Errorf("failed to verify bundle: %w", err)
				}
			} else if validationLevel == bundle.ChecksumValidationLevel {
				// Validate checksums
				result, err = bundle.VerifyBundleChecksums(offlineBundle)
				if err != nil {
					return fmt.Errorf("failed to verify bundle checksums: %w", err)
				}
			} else {
				return fmt.Errorf("validation level %s not implemented yet", level)
			}

			// Print validation results
			if result.Valid {
				fmt.Fprintf(c.Output, "✅ Validation successful: %s\n", result.Message)
			} else {
				fmt.Fprintf(c.Output, "❌ Validation failed: %s\n", result.Message)
				for _, err := range result.Errors {
					fmt.Fprintf(c.Output, "  - %s\n", err)
				}
				return fmt.Errorf("bundle validation failed")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Path to the offline bundle directory (required)")
	cmd.Flags().StringVarP(&publicKeyPath, "key", "k", "", "Path to the public key file for signature verification")
	cmd.Flags().StringVarP(&level, "level", "l", "standard", "Validation level (basic, standard, strict, manifest, checksum, signature, compatibility)")

	// Mark required flags
	cmd.MarkFlagRequired("bundle")

	return cmd
}

// createKeygenCommand creates the 'keygen' command for generating signing keys
func (c *OfflineBundleCLI) createKeygenCommand() *cobra.Command {
	var outputPath string
	var keyType string
	var force bool

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate signing keys for offline bundles",
		Long:  "Generate public and private keys for signing and verifying offline bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Check if key files already exist
			privateKeyPath := filepath.Join(outputPath, "private.key")
			publicKeyPath := filepath.Join(outputPath, "public.key")

			if !force {
				if _, err := os.Stat(privateKeyPath); err == nil {
					return fmt.Errorf("private key file already exists: %s (use --force to overwrite)", privateKeyPath)
				}
				if _, err := os.Stat(publicKeyPath); err == nil {
					return fmt.Errorf("public key file already exists: %s (use --force to overwrite)", publicKeyPath)
				}
			}

			// Generate key pair based on type
			var privateKey, publicKey []byte
			var err error

			switch strings.ToLower(keyType) {
			case "ed25519":
				// Generate Ed25519 key pair
				publicKey, privateKey, err = bundle.GenerateKeyPair()
				if err != nil {
					return fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
				}
			default:
				return fmt.Errorf("unsupported key type: %s (only ed25519 is currently supported)", keyType)
			}

			// Write private key to file
			if err := os.WriteFile(privateKeyPath, privateKey, 0600); err != nil {
				return fmt.Errorf("failed to write private key file: %w", err)
			}

			// Write public key to file
			if err := os.WriteFile(publicKeyPath, publicKey, 0644); err != nil {
				return fmt.Errorf("failed to write public key file: %w", err)
			}

			// Audit logging would go here, but we'll keep it simple for now
			// TODO: Add proper audit logging when AuditTrailManager interface is finalized

			fmt.Fprintf(c.Output, "Generated signing keys successfully:\n")
			fmt.Fprintf(c.Output, "  Private key: %s\n", privateKeyPath)
			fmt.Fprintf(c.Output, "  Public key: %s\n", publicKeyPath)
			fmt.Fprintf(c.Output, "\nIMPORTANT: Keep your private key secure and back it up safely.\n")
			fmt.Fprintf(c.Output, "The public key can be shared for bundle verification.\n")

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Output directory for the key files")
	cmd.Flags().StringVarP(&keyType, "type", "t", "ed25519", "Key type (currently only ed25519 is supported)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing key files")

	return cmd
}
