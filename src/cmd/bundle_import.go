// v0.10.0 #177 — fresh `bundle import` implementation against the
// live src/bundle API.
//
// Replaces src/cmd/bundle_import.go.disabled (atticked at
// attic/v0-7-0-bundle-disabled/), which referenced APIs that no
// longer exist (bundle.LoadBundleManifest, bundle.ImportEvent, etc.).
//
// Scope: this command takes an EXTRACTED bundle directory (one
// containing manifest.json) and runs the bundle package's
// DefaultBundleImporter end-to-end. Archive extraction is left to a
// separate `bundle extract` invocation or a manual `tar xf` /
// `unzip`; combining the two would duplicate the v0.10.0 #174 Tier 2
// atomic-replace path and create two import surfaces operators have
// to remember.
//
// The honesty invariant: import errors return non-zero. No fake
// success on a failed validation or apply.
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

var bundleImportCmd = &cobra.Command{
	Use:   "import PATH",
	Short: "Import an extracted bundle directory",
	Long: `Import an extracted bundle into the current installation.

PATH must point to a directory containing a manifest.json (extract
your .tar.gz / .zip archive first; this command does not extract).

The import:
  1. Validates the bundle at the requested level (--level).
  2. Optionally creates a backup of the target directory (--backup).
  3. Copies bundle content (templates / modules) into the target dir.

Validation levels (--level):
  basic       — manifest schema + signature check
  standard    — basic + content checksums (default)
  strict      — standard + version-compatibility check

Exit code is 0 only when the import completes successfully.`,
	Example: `  # Default standard validation
  llmrecon bundle import ./extracted-bundle --target=./templates

  # Strict validation with backup
  llmrecon bundle import ./extracted-bundle --target=./templates \
      --level=strict --backup=./templates.bak`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]
		targetDir, _ := cmd.Flags().GetString("target")
		backupDir, _ := cmd.Flags().GetString("backup")
		level, _ := cmd.Flags().GetString("level")
		force, _ := cmd.Flags().GetBool("force")

		if targetDir == "" {
			return fmt.Errorf("--target is required")
		}

		validationLevel, err := parseValidationLevel(level)
		if err != nil {
			return err
		}

		// Pre-flight: load the bundle so we can fail fast on a bad
		// manifest before instantiating the heavyweight importer.
		bndl, err := bundle.LoadBundle(bundlePath)
		if err != nil {
			return fmt.Errorf("load bundle: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Bundle %s @ %s — importing %d items at level=%s\n",
			bndl.Manifest.BundleID,
			bndl.Manifest.Version,
			len(bndl.Manifest.Content),
			validationLevel)

		// Hand off to the bundle package's importer. nil validator
		// and report-manager mean it constructs reasonable defaults.
		importer := bundle.NewBundleImporter(nil, nil, cmd.OutOrStdout())
		result, err := importer.Import(context.Background(), bundlePath, bundle.ImportOptions{
			ValidationLevel: validationLevel,
			TargetDir:       targetDir,
			BackupDir:       backupDir,
			Force:           force,
			Logger:          cmd.OutOrStdout(),
		})
		if err != nil {
			// The importer's Import returns a result even on error so
			// the report includes structured failure info — surface
			// both the error AND the result's collected errors.
			if result != nil {
				for _, e := range result.Errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
				}
			}
			return fmt.Errorf("import failed: %w", err)
		}
		if !result.Success {
			for _, e := range result.Errors {
				fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
			}
			return fmt.Errorf("import did not complete successfully (%d error(s))", len(result.Errors))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✓ Imported %d files into %s\n", len(result.ImportedFiles), targetDir)
		if len(result.SkippedFiles) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  (%d files skipped)\n", len(result.SkippedFiles))
		}
		for _, w := range result.Warnings {
			fmt.Fprintf(cmd.OutOrStdout(), "  warning: %s\n", w)
		}
		return nil
	},
}

// parseValidationLevel maps the operator-friendly --level string to
// the bundle package's enum. Centralized so verify and import share
// the same mapping; the issue-#177 callout that the existing
// validation-level switch only handled 2 of 7 documented levels was
// fixed by routing through this helper.
func parseValidationLevel(level string) (bundle.ValidationLevel, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "standard":
		return bundle.StandardValidation, nil
	case "basic":
		return bundle.BasicValidation, nil
	case "strict":
		return bundle.StrictValidation, nil
	case "manifest":
		return bundle.ManifestValidationLevel, nil
	case "checksum":
		return bundle.ChecksumValidationLevel, nil
	case "signature":
		return bundle.SignatureValidationLevel, nil
	case "compatibility":
		return bundle.CompatibilityValidationLevel, nil
	default:
		return "", fmt.Errorf("unknown validation level %q (valid: basic, standard, strict, manifest, checksum, signature, compatibility)", level)
	}
}

func init() {
	bundleCmd.AddCommand(bundleImportCmd)
	bundleImportCmd.Flags().String("target", "", "Target directory to import bundle content into (required)")
	bundleImportCmd.Flags().String("backup", "", "Optional backup directory; if set, target dir is backed up before import")
	bundleImportCmd.Flags().String("level", "standard", "Validation level: basic, standard, strict, manifest, checksum, signature, compatibility")
	bundleImportCmd.Flags().Bool("force", false, "Force import even if there are conflicts")
}
