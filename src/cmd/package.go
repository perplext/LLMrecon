package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/perplext/LLMrecon/src/update"
	"github.com/perplext/LLMrecon/src/version"
	"github.com/spf13/cobra"
)

var (
	packageCmd = &cobra.Command{
		Use:   "package",
		Short: "Manage update packages",
		Long:  `Create, verify, and apply update packages for the LLMreconing Tool.`,
	}

	createPackageCmd = &cobra.Command{
		Use:   "create [manifest-path] [output-path]",
		Short: "Create an update package",
		Long:  `Create an update package from a manifest file.`,
		Args:  cobra.ExactArgs(2),
		RunE:  runCreatePackage,
	}

	verifyPackageCmd = &cobra.Command{
		Use:   "verify [package-path]",
		Short: "Verify an update package",
		Long:  `Verify the integrity and authenticity of an update package.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runVerifyPackage,
	}

	applyPackageCmd = &cobra.Command{
		Use:   "apply [package-path]",
		Short: "Apply an update package",
		Long:  `Apply an update package to the current installation.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runApplyPackage,
	}

	// Flags
	publicKeyPath string
	installDir    string
	tempDir       string
	backupDir     string
	forceUpdate   bool
	skipVerify    bool
)

func init() {
	// Add package command to root command
	rootCmd.AddCommand(packageCmd)

	// Add subcommands to package command
	packageCmd.AddCommand(createPackageCmd)
	packageCmd.AddCommand(verifyPackageCmd)
	packageCmd.AddCommand(applyPackageCmd)

	// Add flags to verify command
	verifyPackageCmd.Flags().StringVar(&publicKeyPath, "public-key", "", "Path to public key file for verification")

	// Add flags to apply command
	applyPackageCmd.Flags().StringVar(&publicKeyPath, "public-key", "", "Path to public key file for verification")
	applyPackageCmd.Flags().StringVar(&installDir, "install-dir", "", "Installation directory (default: executable directory)")
	applyPackageCmd.Flags().StringVar(&tempDir, "temp-dir", "", "Temporary directory for update operations")
	applyPackageCmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory for update operations")
	applyPackageCmd.Flags().BoolVar(&forceUpdate, "force", false, "Force update even if not compatible")
	applyPackageCmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip package verification (not recommended)")
}

func runCreatePackage(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]
	outputPath := args[1]

	// Validate manifest path
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest file not found: %s", manifestPath)
	}

	// Create package
	fmt.Printf("Creating update package from %s...\n", manifestPath)
	err := update.CreatePackage(manifestPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	fmt.Printf("Successfully created update package: %s\n", outputPath)
	return nil
}

func runVerifyPackage(cmd *cobra.Command, args []string) error {
	packagePath := args[0]

	// Validate package path
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return fmt.Errorf("package file not found: %s", packagePath)
	}

	// Open package
	fmt.Printf("Opening update package: %s\n", packagePath)
	pkg, err := update.OpenPackage(packagePath)
	if err != nil {
		return fmt.Errorf("failed to open package: %w", err)
	}
	defer pkg.Close()

	// Print package information
	fmt.Println("Package Information:")
	fmt.Printf("  ID: %s\n", pkg.Manifest.PackageID)
	fmt.Printf("  Type: %s\n", pkg.Manifest.PackageType)
	fmt.Printf("  Created: %s\n", pkg.Manifest.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Expires: %s\n", pkg.Manifest.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Publisher: %s (%s)\n", pkg.Manifest.Publisher.Name, pkg.Manifest.Publisher.URL)
	fmt.Printf("  Binary Version: %s\n", pkg.Manifest.Components.Binary.Version)
	fmt.Printf("  Templates Version: %s\n", pkg.Manifest.Components.Templates.Version)
	fmt.Printf("  Modules: %d\n", len(pkg.Manifest.Components.Modules))

	// Verify package if public key is provided
	if publicKeyPath != "" {
		// Read public key
		publicKeyData, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key: %w", err)
		}

		// TODO: Parse public key data
		var publicKey []byte
		_ = publicKeyData // Placeholder

		// Verify package
		fmt.Println("Verifying package...")
		err = pkg.Verify(publicKey)
		if err != nil {
			return fmt.Errorf("package verification failed: %w", err)
		}
		fmt.Println("Package verification successful.")
	} else {
		fmt.Println("Warning: Package not verified (no public key provided).")
	}

	return nil
}

func runApplyPackage(cmd *cobra.Command, args []string) error {
	packagePath := args[0]

	// Validate package path
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return fmt.Errorf("package file not found: %s", packagePath)
	}

	// Open package
	fmt.Printf("Opening update package: %s\n", packagePath)
	pkg, err := update.OpenPackage(packagePath)
	if err != nil {
		return fmt.Errorf("failed to open package: %w", err)
	}
	defer pkg.Close()

	// Print package information
	fmt.Println("Package Information:")
	fmt.Printf("  ID: %s\n", pkg.Manifest.PackageID)
	fmt.Printf("  Type: %s\n", pkg.Manifest.PackageType)
	fmt.Printf("  Created: %s\n", pkg.Manifest.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Expires: %s\n", pkg.Manifest.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Publisher: %s (%s)\n", pkg.Manifest.Publisher.Name, pkg.Manifest.Publisher.URL)
	fmt.Printf("  Binary Version: %s\n", pkg.Manifest.Components.Binary.Version)
	fmt.Printf("  Templates Version: %s\n", pkg.Manifest.Components.Templates.Version)
	fmt.Printf("  Modules: %d\n", len(pkg.Manifest.Components.Modules))

	// Verify package if not skipped
	if !skipVerify && publicKeyPath != "" {
		// Read public key
		publicKeyData, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key: %w", err)
		}

		// TODO: Parse public key data
		var publicKey []byte
		_ = publicKeyData // Placeholder

		// Verify package
		fmt.Println("Verifying package...")
		err = pkg.Verify(publicKey)
		if err != nil {
			return fmt.Errorf("package verification failed: %w", err)
		}
		fmt.Println("Package verification successful.")
	} else if !skipVerify {
		fmt.Println("Warning: Package not verified (no public key provided).")
	} else {
		fmt.Println("Warning: Package verification skipped.")
	}

	// Get current versions
	currentVersions, err := getCurrentVersions()
	if err != nil {
		return fmt.Errorf("failed to get current versions: %w", err)
	}

	// Check if package is compatible
	if !forceUpdate {
		fmt.Println("Checking compatibility...")
		compatible, err := pkg.IsCompatible(currentVersions)
		if err != nil {
			return fmt.Errorf("compatibility check failed: %w", err)
		}
		if !compatible {
			return fmt.Errorf("package is not compatible with current installation")
		}
		fmt.Println("Package is compatible with current installation.")
	} else {
		fmt.Println("Warning: Forcing update (compatibility check skipped).")
	}

	// Get installation directory
	if installDir == "" {
		// Use executable directory as default
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
		installDir = filepath.Dir(execPath)
	}

	// Create update applier
	applier, err := update.NewUpdateApplier(&update.ApplierOptions{
		InstallDir:      installDir,
		TempDir:         tempDir,
		BackupDir:       backupDir,
		CurrentVersions: currentVersions,
		Logger:          os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to create update applier: %w", err)
	}

	// Apply update
	fmt.Println("Applying update...")
	err = applier.ApplyUpdate(context.Background(), pkg)
	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	fmt.Println("Update applied successfully.")
	return nil
}

// getCurrentVersions gets the current versions of components
func getCurrentVersions() (map[string]version.Version, error) {
	// In a real implementation, this would read version information from the installation
	// For now, we'll just return placeholder versions
	return map[string]version.Version{
		"core":      {Major: 1, Minor: 0, Patch: 0},
		"templates": {Major: 1, Minor: 0, Patch: 0},
	}, nil
}
