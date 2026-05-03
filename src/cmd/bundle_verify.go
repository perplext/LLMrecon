// v0.10.0 #177 — fresh `bundle verify` implementation against the
// live src/bundle API.
//
// Replaces src/cmd/bundle_verify.go.disabled (atticked at
// attic/v0-7-0-bundle-disabled/), which referenced APIs that no
// longer exist (bundle.LoadBundleManifest, etc.). This file is
// minimal and routes to the existing signature.go and validator.go
// entry points.
//
// Usage:
//
//	llmrecon bundle verify <path-to-extracted-bundle>
//	llmrecon bundle verify <path> --public-key=key.pub
//	llmrecon bundle verify <path> --level=checksum
//
// The path must be an EXTRACTED bundle directory (containing
// manifest.json), not a .tar.gz / .zip archive. Extraction is the
// caller's responsibility — `bundle import` handles the full
// extract+verify+install loop, `bundle verify` is the verification-
// only primitive operators use to check bundle integrity before
// committing to import.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

var bundleVerifyCmd = &cobra.Command{
	Use:   "verify PATH",
	Short: "Verify an extracted bundle's signature, checksums, or manifest",
	Long: `Verify the integrity of an extracted bundle directory.

PATH must point to a directory containing a manifest.json (an extracted
bundle, not a .tar.gz / .zip archive).

Verification levels (--level):
  signature   — verify Ed25519 signature against --public-key
  checksum    — verify content checksums against the manifest
  manifest    — validate manifest schema (default)

Exit code is 0 only when verification succeeds. Failed verifications
print structured errors and exit non-zero.`,
	Example: `  # Default: validate manifest schema only
  llmrecon bundle verify ./extracted-bundle

  # Verify signature with provided public key
  llmrecon bundle verify ./extracted-bundle --public-key=keys/release.pub

  # Verify content checksums
  llmrecon bundle verify ./extracted-bundle --level=checksum`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bundlePath := args[0]
		publicKeyPath, _ := cmd.Flags().GetString("public-key")
		level, _ := cmd.Flags().GetString("level")

		// Validate --level BEFORE loading the bundle so a bad level
		// fails fast without I/O. Mirrors bundle_import.go's
		// fail-fast ordering and shares the same TrimSpace+ToLower
		// normalization (otherwise --level=" manifest " silently
		// hits the default branch).
		normLevel := strings.ToLower(strings.TrimSpace(level))
		switch normLevel {
		case "signature":
			if publicKeyPath == "" {
				return fmt.Errorf("--public-key is required for --level=signature")
			}
		case "checksum", "", "manifest":
			// valid; fall through
		default:
			return fmt.Errorf("unknown --level %q (valid: signature, checksum, manifest)", level)
		}

		bndl, err := bundle.LoadBundle(bundlePath)
		if err != nil {
			return fmt.Errorf("load bundle: %w", err)
		}

		// Each level routes to the corresponding purpose-built
		// verification entry point. The level was already validated
		// above; this switch is exhaustive for the accepted set.
		switch normLevel {
		case "signature":
			pubKey, err := os.ReadFile(filepath.Clean(publicKeyPath))
			if err != nil {
				return fmt.Errorf("read public key %q: %w", publicKeyPath, err)
			}
			result, err := bundle.VerifyBundle(bndl, pubKey)
			if err != nil {
				return fmt.Errorf("verify signature: %w", err)
			}
			return reportValidation(cmd, result)

		case "checksum":
			result, err := bundle.VerifyBundleChecksums(bndl)
			if err != nil {
				return fmt.Errorf("verify checksums: %w", err)
			}
			return reportValidation(cmd, result)

		default: // "" or "manifest" — guaranteed by the validation switch above
			validator := bundle.NewBundleValidator(cmd.OutOrStdout())
			result, err := validator.ValidateManifest(&bndl.Manifest)
			if err != nil {
				return fmt.Errorf("validate manifest: %w", err)
			}
			return reportValidation(cmd, result)
		}
	},
}

// reportValidation prints a human-readable summary and returns a
// non-nil error when result.Valid is false so the CLI exits non-zero.
// The "Successfully verified" message only prints on actual success —
// the v0.10.0 honesty invariant: no fake success on a failed verify.
func reportValidation(cmd *cobra.Command, result *bundle.ValidationResult) error {
	out := cmd.OutOrStdout()
	if result == nil {
		return fmt.Errorf("validator returned nil result (this is a bug; please file an issue)")
	}
	if result.Valid {
		if result.Message != "" {
			fmt.Fprintf(out, "✓ %s\n", result.Message)
		} else {
			fmt.Fprintln(out, "✓ Bundle verified.")
		}
		for _, w := range result.Warnings {
			fmt.Fprintf(out, "  warning: %s\n", w)
		}
		return nil
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "✗ Bundle verification failed: %s\n", result.Message)
	for _, e := range result.Errors {
		fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
	}
	return fmt.Errorf("bundle verification failed (level=%s)", result.Level)
}

func init() {
	bundleCmd.AddCommand(bundleVerifyCmd)
	bundleVerifyCmd.Flags().String("public-key", "", "Path to Ed25519 public key for signature verification")
	bundleVerifyCmd.Flags().String("level", "manifest", "Verification level: signature, checksum, manifest")
}
