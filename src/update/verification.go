// Package update provides functionality for checking and applying updates
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/version"
)

// VerificationResult represents the result of a verification operation
type VerificationResult struct {
	// Success indicates whether the verification was successful
	Success bool
	// Message contains a human-readable message about the verification
	Message string
	// Details contains additional details about the verification
	Details map[string]interface{}
}

// IntegrityVerifier handles verification of update package integrity
type IntegrityVerifier struct {
	// Logger is the logger for verification operations
	Logger io.Writer
}

// NewIntegrityVerifier creates a new integrity verifier
func NewIntegrityVerifier(logger io.Writer) *IntegrityVerifier {
	return &IntegrityVerifier{
		Logger: logger,
	}
}

// VerifyPackage verifies the integrity of an update package
func (v *IntegrityVerifier) VerifyPackage(pkg *UpdatePackage) (*VerificationResult, error) {
	// Log verification start
	fmt.Fprintf(v.Logger, "Verifying package integrity: %s\n", pkg.PackagePath)

	// Check if manifest exists
	manifestPath := filepath.Join(pkg.PackagePath, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return &VerificationResult{
			Success: false,
			Message: "Manifest file not found",
		}, fmt.Errorf("manifest file not found")
	}

	// Verify digital signature if provided — a hard refusal that precedes
	// checksum work.
	//
	// v0.10.0 #174 Tier 1: refuse rather than silently log+pass when a
	// signature is present but the verifier is unimplemented. Operators
	// who packaged a signed bundle expect verification to either succeed
	// or fail visibly — silently bypassing means a tampered bundle could
	// pass through this code path with `Success: true`.
	if pkg.Manifest.Signature != "" {
		fmt.Fprintf(v.Logger, "Digital signature verification not implemented; refusing to claim verification success on a signed bundle\n")
		return &VerificationResult{
			Success: false,
			Message: "Digital signature verification not implemented in this version (deferred to v0.11.0); refuse-on-signed-bundle is the v0.10.0 #174 Tier 1 fail-safe",
		}, fmt.Errorf("digital signature verification not implemented; cannot verify signed bundle")
	}

	// Verify the component checksums declared in the manifest against the
	// package payloads on disk. Honesty invariant (#224): an IntegrityVerifier
	// must not return Success:true unless it actually verified integrity. This
	// replaces the prior "skip checksum verification" path that returned
	// success unconditionally.
	verified, err := v.verifyComponentChecksums(pkg)
	if err != nil {
		fmt.Fprintf(v.Logger, "Component checksum verification failed: %v\n", err)
		return &VerificationResult{
			Success: false,
			Message: fmt.Sprintf("Component checksum verification failed: %v", err),
			Details: map[string]interface{}{"verified_components": verified},
		}, fmt.Errorf("component checksum verification failed: %w", err)
	}
	if verified == 0 {
		// Nothing was verifiable — the manifest declares no component checksums.
		// Refuse to claim integrity success rather than silently passing.
		return &VerificationResult{
			Success: false,
			Message: "Manifest declares no component checksums; package integrity cannot be verified",
		}, fmt.Errorf("no component checksums declared; cannot verify package integrity")
	}

	// Log verification success
	fmt.Fprintf(v.Logger, "Package integrity verification successful (%d component checksum(s) verified)\n", verified)

	return &VerificationResult{
		Success: true,
		Message: "Package integrity verification successful",
		Details: map[string]interface{}{"verified_components": verified},
	}, nil
}

// verifyComponentChecksums verifies every component checksum declared in the
// package manifest against the corresponding payload on disk, and returns the
// number of checksums successfully verified.
//
// Payload layout convention (relative to pkg.PackagePath, an extracted package
// directory):
//   - templates:  templates/            (directory hash)
//   - modules:    modules/<module-id>   (file hash)
//   - binary:     binary/<platform>     (file hash, per declared platform)
//
// A declared checksum whose payload is missing, unreadable, or mismatched is a
// hard failure (returns an error) — the verifier never treats an unverifiable
// declared checksum as a pass.
func (v *IntegrityVerifier) verifyComponentChecksums(pkg *UpdatePackage) (int, error) {
	verified := 0
	c := pkg.Manifest.Components

	// Templates component — directory hash.
	if c.Templates.Checksum != "" {
		dir := filepath.Join(pkg.PackagePath, "templates")
		got, err := calculateDirectoryHash(dir)
		if err != nil {
			return verified, fmt.Errorf("templates payload: %w", err)
		}
		if !hashEqual(got, c.Templates.Checksum) {
			return verified, fmt.Errorf("templates checksum mismatch (manifest %s, computed %s)", c.Templates.Checksum, got)
		}
		verified++
	}

	// Module components — per-module file hash.
	for _, m := range c.Modules {
		if m.Checksum == "" {
			continue
		}
		path := filepath.Join(pkg.PackagePath, "modules", m.ID)
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return verified, fmt.Errorf("module %q payload: %w", m.ID, err)
		}
		if !hashEqual(calculateFileHash(data), m.Checksum) {
			return verified, fmt.Errorf("module %q checksum mismatch", m.ID)
		}
		verified++
	}

	// Binary component — per-platform file hash.
	for platform, sum := range c.Binary.Checksums {
		if sum == "" {
			continue
		}
		path := filepath.Join(pkg.PackagePath, "binary", platform)
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return verified, fmt.Errorf("binary[%s] payload: %w", platform, err)
		}
		if !hashEqual(calculateFileHash(data), sum) {
			return verified, fmt.Errorf("binary[%s] checksum mismatch", platform)
		}
		verified++
	}

	return verified, nil
}

// hashEqual compares two hex-encoded SHA-256 digests case-insensitively.
func hashEqual(a, b string) bool {
	return strings.EqualFold(a, b)
}

// VerifyCompatibility verifies that the update package is compatible with the current installation
func (v *IntegrityVerifier) VerifyCompatibility(pkg *UpdatePackage, currentVersions map[string]version.Version) (*VerificationResult, error) {
	// Log verification start
	fmt.Fprintf(v.Logger, "Verifying package compatibility\n")

	// Check if package is compatible with current versions
	compatible, err := pkg.IsCompatible(currentVersions)
	if err != nil {
		return &VerificationResult{
			Success: false,
			Message: "Failed to check package compatibility",
		}, fmt.Errorf("failed to check package compatibility: %w", err)
	}

	if !compatible {
		return &VerificationResult{
			Success: false,
			Message: "Package is not compatible with current versions",
		}, fmt.Errorf("package is not compatible with current versions")
	}

	// Log verification success
	fmt.Fprintf(v.Logger, "Package compatibility verification successful\n")

	return &VerificationResult{
		Success: true,
		Message: "Package compatibility verification successful",
	}, nil
}

// calculateFileHash calculates the SHA-256 hash of a file
func calculateFileHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// calculateDirectoryHash calculates the SHA-256 hash of a directory
func calculateDirectoryHash(dirPath string) (string, error) {
	// Create a hash object
	h := sha256.New()

	// Walk through the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read file
		data, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Update hash with file path and content
		h.Write([]byte(relPath))
		h.Write(data)

		return nil
	})

	if err != nil {
		return "", err
	}

	// Return hash as hex string
	return hex.EncodeToString(h.Sum(nil)), nil
}
