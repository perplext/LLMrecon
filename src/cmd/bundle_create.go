package cmd

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/spf13/cobra"
)

// OWASP LLM Top 10 categories
var owaspCategories = []string{
	"llm01-prompt-injection",
	"llm02-insecure-output",
	"llm03-training-data-poisoning",
	"llm04-model-denial-of-service",
	"llm05-supply-chain",
	"llm06-sensitive-information",
	"llm07-insecure-plugin",
	"llm08-excessive-agency",
	"llm09-overreliance",
	"llm10-model-theft",
}

// bundleCreateCmd represents the bundle create command
var bundleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new bundle",
	Long: `Create a new offline update bundle containing components and templates.

This command packages selected components into a bundle file that can be:
- Transferred to air-gapped environments
- Verified for integrity and compliance
- Imported to update the system

Templates are organized by OWASP LLM Top 10 categories with proper versioning.`,
	Example: `  # Create a bundle with all components
  LLMrecon bundle create --output=update.bundle

  # Create a bundle with only templates
  LLMrecon bundle create --component=templates --output=templates.bundle

  # Create a bundle for specific OWASP category
  LLMrecon bundle create --category=llm01-prompt-injection --output=prompt-injection.bundle

  # Create a bundle from GitHub source
  LLMrecon bundle create --source=github --output=github-templates.bundle

  # Create a bundle with custom filters
  LLMrecon bundle create --filter="security/*" --output=security-templates.bundle`,
	RunE: runBundleCreate,
}

func init() {
	bundleCmd.AddCommand(bundleCreateCmd)

	// Add flags
	bundleCreateCmd.Flags().StringP("output", "o", "bundle.tar.gz", "Output bundle file path")
	bundleCreateCmd.Flags().StringP("component", "c", "all", "Component to bundle (all, binary, templates, modules)")
	bundleCreateCmd.Flags().StringP("filter", "f", "", "Filter pattern for selecting files")
	bundleCreateCmd.Flags().StringP("source", "s", "local", "Source for templates (local, github, gitlab)")
	bundleCreateCmd.Flags().String("category", "", "OWASP LLM category to bundle")
	bundleCreateCmd.Flags().Bool("include-compliance", true, "Include compliance documentation")
	bundleCreateCmd.Flags().Bool("sign", false, "Sign the bundle with GPG")
	bundleCreateCmd.Flags().String("key", "", "GPG key ID for signing")
	bundleCreateCmd.Flags().Bool("compress", true, "Compress the bundle")
	bundleCreateCmd.Flags().String("compression", "gzip", "Compression algorithm (gzip, zstd)")
	bundleCreateCmd.Flags().Bool("encrypt", false, "Encrypt the bundle")
	bundleCreateCmd.Flags().String("password", "", "Password for encryption")
	bundleCreateCmd.Flags().Duration("timeout", 10*time.Minute, "Timeout for remote operations")
	bundleCreateCmd.Flags().Bool("verbose", false, "Verbose output")
}

func runBundleCreate(cmd *cobra.Command, args []string) error {
	// Get flags
	output, _ := cmd.Flags().GetString("output")
	component, _ := cmd.Flags().GetString("component")
	filter, _ := cmd.Flags().GetString("filter")
	source, _ := cmd.Flags().GetString("source")
	category, _ := cmd.Flags().GetString("category")
	includeCompliance, _ := cmd.Flags().GetBool("include-compliance")
	sign, _ := cmd.Flags().GetBool("sign")
	keyID, _ := cmd.Flags().GetString("key")
	compress, _ := cmd.Flags().GetBool("compress")
	compressionAlg, _ := cmd.Flags().GetString("compression")
	encrypt, _ := cmd.Flags().GetBool("encrypt")
	password, _ := cmd.Flags().GetString("password")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Validate output path
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Display creation parameters
	fmt.Println(bold("Creating Bundle"))
	fmt.Printf("%s: %s\n", cyan("Output"), output)
	fmt.Printf("%s: %s\n", cyan("Component"), component)
	if filter != "" {
		fmt.Printf("%s: %s\n", cyan("Filter"), filter)
	}
	if category != "" {
		fmt.Printf("%s: %s\n", cyan("Category"), category)
	}
	fmt.Printf("%s: %s\n", cyan("Source"), source)
	fmt.Println()

	// Create bundle options
	bundleOpts := createBundleOptions(cfg, component, filter, category, source, includeCompliance, compress, compressionAlg, encrypt, password)

	// Add progress handler
	bundleOpts.ProgressHandler = createProgressHandler(verbose)

	// Create bundle exporter
	exporter, err := createBundleExporter(ctx, cfg, source, bundleOpts)
	if err != nil {
		return fmt.Errorf("creating bundle exporter: %w", err)
	}

	// Export bundle
	fmt.Println(dim("Collecting components..."))
	if err := exporter.Export(); err != nil {
		return fmt.Errorf("exporting bundle: %w", err)
	}

	// Sign bundle if requested
	if sign {
		fmt.Println(dim("Signing bundle..."))
		if err := signBundle(output, keyID); err != nil {
			return fmt.Errorf("signing bundle: %w", err)
		}
	}

	// Display summary
	info, err := os.Stat(output)
	if err != nil {
		return fmt.Errorf("getting bundle info: %w", err)
	}

	fmt.Println()
	fmt.Println(green("✓") + " Bundle created successfully")
	fmt.Printf("%s: %s\n", cyan("Size"), formatSize(info.Size()))
	fmt.Printf("%s: %s\n", cyan("SHA256"), calculateBundleChecksum(output))

	if sign {
		fmt.Printf("%s: %s.sig\n", cyan("Signature"), output)
	}

	return nil
}

func createBundleOptions(cfg *config.Config, component, filter, category, source string, includeCompliance, compress bool, compressionAlg string, encrypt bool, password string) *bundle.ExportOptions {
	opts := &bundle.ExportOptions{
		Format:           bundle.FormatTarGz,
		IncludeBinary:    component == "all" || component == "binary",
		IncludeTemplates: component == "all" || component == "templates",
		IncludeModules:   component == "all" || component == "modules",
		IncludeDocs:      includeCompliance,
		Metadata: map[string]interface{}{
			"created_at":   time.Now().Format(time.RFC3339),
			"tool_version": currentVersion,
			"source":       source,
		},
	}

	// Set compression
	if compress {
		switch compressionAlg {
		case "zstd":
			opts.Compression = bundle.CompressionZstd
		default:
			opts.Compression = bundle.CompressionGzip
		}
	} else {
		opts.Compression = bundle.CompressionNone
	}

	// Set encryption
	if encrypt && password != "" {
		opts.Encryption = &bundle.EncryptionOptions{
			Algorithm: "aes-256-gcm",
			Password:  password,
		}
	}

	// Set filters
	if filter != "" || category != "" {
		opts.Filters = &bundle.ExportFilters{}

		if filter != "" {
			opts.Filters.IncludeList = []string{filter}
		}

		if category != "" {
			// Add OWASP category filter
			opts.Filters.TemplateCategories = []string{category}
			opts.Metadata["owasp_category"] = category
		}
	}

	// Add compliance metadata
	if includeCompliance {
		opts.Metadata["compliance"] = map[string]interface{}{
			"iso42001": true,
			"owasp":    true,
		}
	}

	return opts
}

func createBundleExporter(ctx context.Context, cfg *config.Config, source string, opts *bundle.ExportOptions) (*bundle.BundleExporter, error) {
	switch source {
	case "github":
		// Clone/fetch from GitHub
		fmt.Println(dim("Fetching templates from GitHub..."))
		_, err := fetchFromGitHub(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("fetching from GitHub: %w", err)
		}

	case "gitlab":
		// Clone/fetch from GitLab
		fmt.Println(dim("Fetching templates from GitLab..."))
		_, err := fetchFromGitLab(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("fetching from GitLab: %w", err)
		}

	default:
		// Use local source
	}

	return bundle.NewBundleExporter(*opts), nil
}

func createProgressHandler(verbose bool) bundle.ExportProgressHandler {
	startTime := time.Now()

	return func(progress bundle.ProgressInfo) {
		// Show progress
		if progress.Message != "" && verbose {
			fmt.Printf("%s: %s\n", progress.Stage, progress.Message)
		}

		// Show progress bar for file operations
		if progress.Total > 0 {
			fmt.Printf("\r  Progress: %3.0f%% [%d/%d files]",
				progress.Percentage, progress.Current, progress.Total)
		}

		// Show current file in verbose mode
		if verbose && progress.CurrentFile != "" {
			fmt.Printf("\n  %s %s", dim("→"), progress.CurrentFile)
		}

		// Show completion
		if progress.Stage == "completed" {
			elapsed := time.Since(startTime)
			fmt.Printf("\n\n%s Completed in %s\n", green("✓"), elapsed.Round(time.Millisecond))
		}

	}
}

func fetchFromGitHub(ctx context.Context, cfg *config.Config) (string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use git or GitHub API
	return "", fmt.Errorf("GitHub integration not implemented yet")
}

func fetchFromGitLab(ctx context.Context, cfg *config.Config) (string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use git or GitLab API
	return "", fmt.Errorf("GitLab integration not implemented yet")
}

func signBundle(bundlePath, keyID string) error {
	// This is a placeholder implementation
	// In a real implementation, this would use GPG or similar
	return fmt.Errorf("bundle signing not implemented yet")
}

func calculateBundleChecksum(bundlePath string) string {
	// Calculate SHA256 checksum
	checksum, err := update.CalculateChecksum(bundlePath)
	if err != nil {
		return "error"
	}

	// Return first 12 characters for display
	if len(checksum) > 12 {
		return checksum[:12] + "..."
	}
	return checksum
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ValidateOWASPCategory validates if a category is valid
func ValidateOWASPCategory(category string) bool {
	for _, valid := range owaspCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// GetOWASPCategories returns all OWASP LLM Top 10 categories
func GetOWASPCategories() []string {
	return owaspCategories
}
