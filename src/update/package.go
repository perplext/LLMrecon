// Package update provides functionality for checking and applying updates
package update

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/version"
)

// PackageType represents the type of update package
type PackageType string

const (
	// FullPackage represents a full update package containing complete files
	FullPackage PackageType = "full"
	// DifferentialPackage represents a differential update package containing patches
	DifferentialPackage PackageType = "differential"
)

// ComponentType represents the type of component in an update package
type ComponentType string

const (
	// BinaryComponent represents the core binary component
	BinaryComponent ComponentType = "binary"
	// TemplatesComponent represents the templates component
	TemplatesComponent ComponentType = "templates"
	// ModulesComponent represents the modules component
	ModulesComponent ComponentType = "modules"
)

// Publisher represents the publisher of an update package
type Publisher struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	PublicKeyID string `json:"public_key_id"`
}

// BinaryComponent represents the binary component in an update package
type BinaryComponentInfo struct {
	Version      string            `json:"version"`
	Platforms    []string          `json:"platforms"`
	MinVersion   string            `json:"min_version"`
	Required     bool              `json:"required"`
	ChangelogURL string            `json:"changelog_url"`
	Checksums    map[string]string `json:"checksums"`

// TemplatesComponent represents the templates component in an update package
type TemplatesComponentInfo struct {
	Version      string   `json:"version"`
	MinVersion   string   `json:"min_version"`
	Required     bool     `json:"required"`
	ChangelogURL string   `json:"changelog_url"`
	Checksum     string   `json:"checksum"`
	Categories   []string `json:"categories"`
	TemplateCount int     `json:"template_count"`
}

// ModuleDependency represents a dependency for a module
type ModuleDependency struct {
	ID         string `json:"id"`
	MinVersion string `json:"min_version"`

// ModuleComponent represents a module component in an update package
type ModuleComponentInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	MinVersion   string            `json:"min_version"`
	Required     bool              `json:"required"`
	ChangelogURL string            `json:"changelog_url"`
	Checksum     string            `json:"checksum"`
	Dependencies []ModuleDependency `json:"dependencies"`

// PatchInfo represents information about a patch in a differential update
type PatchInfo struct {
	FromVersion string            `json:"from_version"`
	ToVersion   string            `json:"to_version"`
	Platforms   []string          `json:"platforms,omitempty"`
	Checksums   map[string]string `json:"checksums,omitempty"`
	Checksum    string            `json:"checksum,omitempty"`
	ID          string            `json:"id,omitempty"`

// PatchesInfo represents all patches in a differential update
type PatchesInfo struct {
	Binary    []PatchInfo `json:"binary,omitempty"`
	Templates []PatchInfo `json:"templates,omitempty"`
	Modules   []PatchInfo `json:"modules,omitempty"`
}

// ComplianceInfo represents compliance information for standards
type ComplianceInfo struct {
	Version  string   `json:"version"`
	Coverage []string `json:"coverage,omitempty"`
	Controls []string `json:"controls,omitempty"`
}

// ComplianceMap represents a map of compliance standards to their information
type ComplianceMap struct {
	OWASPLLMTop10 ComplianceInfo `json:"owasp_llm_top10"`
	ISO42001      ComplianceInfo `json:"iso_42001"`

// Components represents all components in an update package
type Components struct {
	Binary    BinaryComponentInfo    `json:"binary"`
	Templates TemplatesComponentInfo `json:"templates"`
	Modules   []ModuleComponentInfo  `json:"modules"`
	Patches   PatchesInfo            `json:"patches,omitempty"`

// PackageManifest represents the manifest of an update package
type PackageManifest struct {
	SchemaVersion string       `json:"schema_version"`
	PackageID     string       `json:"package_id"`
	PackageType   PackageType  `json:"package_type"`
	CreatedAt     time.Time    `json:"created_at"`
	ExpiresAt     time.Time    `json:"expires_at"`
	Publisher     Publisher    `json:"publisher"`
	Components    Components   `json:"components"`
	Compliance    ComplianceMap `json:"compliance"`
	Signature     string       `json:"signature"`
}

// UpdatePackage represents an update package
type UpdatePackage struct {
	Manifest    PackageManifest
	PackagePath string
	reader      *zip.ReadCloser
	verified    bool

// OpenPackage opens an update package from the given path
func OpenPackage(path string) (*UpdatePackage, error) {
	// Open the zip file
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open update package: %w", err)
	}

	// Create update package
	pkg := &UpdatePackage{
		PackagePath: path,
		reader:      reader,
		verified:    false,
	}

	// Read manifest
	err = pkg.readManifest()
	if err != nil {
		reader.Close()
		return nil, err
	}

	return pkg, nil

// Close closes the update package
func (p *UpdatePackage) Close() error {
	if p.reader != nil {
		return p.reader.Close()
	}
	return nil

// readManifest reads the manifest from the update package
func (p *UpdatePackage) readManifest() error {
	// Find manifest file
	var manifestFile *zip.File
	for _, file := range p.reader.File {
		if file.Name == "manifest.json" {
			manifestFile = file
			break
		}
	}

	if manifestFile == nil {
		return fmt.Errorf("manifest.json not found in update package")
	}

	// Open manifest file
	rc, err := manifestFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open manifest.json: %w", err)
	}
	defer func() { if err := rc.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Read manifest file
	manifestData, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read manifest.json: %w", err)
	}

	// Parse manifest
	err = json.Unmarshal(manifestData, &p.Manifest)
	if err != nil {
		return fmt.Errorf("failed to parse manifest.json: %w", err)
	}

	return nil

// Verify verifies the integrity and authenticity of the update package
func (p *UpdatePackage) Verify(publicKey ed25519.PublicKey) error {
	// Check if already verified
	if p.verified {
		return nil
	}

	// Check expiration
	now := time.Now().UTC()
	if now.After(p.Manifest.ExpiresAt) {
		return fmt.Errorf("update package has expired on %s", p.Manifest.ExpiresAt.Format(time.RFC3339))
	}

	// Verify manifest signature
	err := p.verifyManifestSignature(publicKey)
	if err != nil {
		return fmt.Errorf("failed to verify manifest signature: %w", err)
	}

	// Verify component checksums
	err = p.verifyComponentChecksums()
	if err != nil {
		return fmt.Errorf("failed to verify component checksums: %w", err)
	}

	// Mark as verified
	p.verified = true
	return nil

// verifyManifestSignature verifies the signature of the manifest
func (p *UpdatePackage) verifyManifestSignature(publicKey ed25519.PublicKey) error {
	// Create a copy of the manifest without the signature
	manifestCopy := p.Manifest
	signature := manifestCopy.Signature
	manifestCopy.Signature = ""

	// Marshal the manifest to JSON
	manifestData, err := json.Marshal(manifestCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Decode signature
	signatureBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(signature, "base64:"))
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature
	if !ed25519.Verify(publicKey, manifestData, signatureBytes) {
		return fmt.Errorf("invalid manifest signature")
	}

	return nil

// verifyComponentChecksums verifies the checksums of all components in the package
func (p *UpdatePackage) verifyComponentChecksums() error {
	// Verify binary checksums
	if p.Manifest.PackageType == FullPackage {
		// Verify binary checksums for each platform
		for platform, checksum := range p.Manifest.Components.Binary.Checksums {
			binaryPath := fmt.Sprintf("binary/%s/LLMrecon", platform)
			if platform == "windows" {
				binaryPath += ".exe"
			}

			err := p.verifyFileChecksum(binaryPath, checksum)
			if err != nil {
				return fmt.Errorf("failed to verify binary checksum for %s: %w", platform, err)
			}
		}

		// Verify templates checksum
		err := p.verifyDirectoryChecksum("templates", p.Manifest.Components.Templates.Checksum)
		if err != nil {
			return fmt.Errorf("failed to verify templates checksum: %w", err)
		}

		// Verify module checksums
		for _, module := range p.Manifest.Components.Modules {
			modulePath := fmt.Sprintf("modules/%s", module.ID)
			err := p.verifyDirectoryChecksum(modulePath, module.Checksum)
			if err != nil {
				return fmt.Errorf("failed to verify module checksum for %s: %w", module.ID, err)
			}
		}
	} else if p.Manifest.PackageType == DifferentialPackage {
		// Verify binary patch checksums
		for _, patch := range p.Manifest.Components.Patches.Binary {
			for platform, checksum := range patch.Checksums {
				patchPath := fmt.Sprintf("patches/binary/%s/%s-%s.patch", 
					platform, patch.FromVersion, patch.ToVersion)
				
				err := p.verifyFileChecksum(patchPath, checksum)
				if err != nil {
					return fmt.Errorf("failed to verify binary patch checksum for %s: %w", platform, err)
				}
			}
		}

		// Verify templates patch checksums
		for _, patch := range p.Manifest.Components.Patches.Templates {
			patchPath := fmt.Sprintf("patches/templates/%s-%s.patch", 
				patch.FromVersion, patch.ToVersion)
			
			err := p.verifyFileChecksum(patchPath, patch.Checksum)
			if err != nil {
				return fmt.Errorf("failed to verify templates patch checksum: %w", err)
			}
		}

		// Verify module patch checksums
		for _, patch := range p.Manifest.Components.Patches.Modules {
			patchPath := fmt.Sprintf("patches/modules/%s/%s-%s.patch", 
				patch.ID, patch.FromVersion, patch.ToVersion)
			
			err := p.verifyFileChecksum(patchPath, patch.Checksum)
			if err != nil {
				return fmt.Errorf("failed to verify module patch checksum for %s: %w", patch.ID, err)
			}
		}
	}

	return nil
// verifyFileChecksum verifies the checksum of a file in the package
func (p *UpdatePackage) verifyFileChecksum(path, expectedChecksum string) error {
	// Find file
	var file *zip.File
	for _, f := range p.reader.File {
		if f.Name == path {
			file = f
			break
		}
	}

	if file == nil {
		return fmt.Errorf("file not found: %s", path)
	}

	// Open file
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { if err := rc.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Calculate checksum
	hash := sha256.New()
	if _, err := io.Copy(hash, rc); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Verify checksum
	actualChecksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil

// verifyDirectoryChecksum verifies the checksum of a directory in the package
func (p *UpdatePackage) verifyDirectoryChecksum(dirPath, expectedChecksum string) error {
	// Find all files in directory
	var files []*zip.File
	prefix := dirPath + "/"
	for _, file := range p.reader.File {
		if strings.HasPrefix(file.Name, prefix) && !strings.HasSuffix(file.Name, "/") {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("directory not found or empty: %s", dirPath)
	}

	// Sort files by name for consistent hashing
	// Note: In a real implementation, we would sort the files here

	// Calculate combined checksum
	hash := sha256.New()
	for _, file := range files {
		// Open file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}

		// Update hash with filename and content
		hash.Write([]byte(file.Name))
		if _, err := io.Copy(hash, rc); err != nil {
			rc.Close()
			return fmt.Errorf("failed to calculate checksum for %s: %w", file.Name, err)
		}
		rc.Close()
	}

	// Verify checksum
	actualChecksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
// IsCompatible checks if the update package is compatible with the current version
func (p *UpdatePackage) IsCompatible(currentVersions map[string]version.Version) (bool, error) {
	// Check binary compatibility
	binaryVersion, hasBinaryVersion := currentVersions["core"]
	if hasBinaryVersion {
		minVersion, err := version.ParseVersion(p.Manifest.Components.Binary.MinVersion)
		if err != nil {
			return false, fmt.Errorf("failed to parse binary minimum version: %w", err)
		}

		if binaryVersion.LessThan(&minVersion) {
			return false, fmt.Errorf("current binary version %s is less than minimum required version %s",
				binaryVersion.String(), minVersion.String())
		}
	}

	// Check templates compatibility
	templatesVersion, hasTemplatesVersion := currentVersions["templates"]
	if hasTemplatesVersion {
		minVersion, err := version.ParseVersion(p.Manifest.Components.Templates.MinVersion)
		if err != nil {
			return false, fmt.Errorf("failed to parse templates minimum version: %w", err)
		}

		if templatesVersion.LessThan(&minVersion) {
			return false, fmt.Errorf("current templates version %s is less than minimum required version %s",
				templatesVersion.String(), minVersion.String())
		}
	}

	// Check modules compatibility
	for _, moduleInfo := range p.Manifest.Components.Modules {
		moduleVersion, hasModuleVersion := currentVersions[fmt.Sprintf("module.%s", moduleInfo.ID)]
		if hasModuleVersion {
			minVersion, err := version.ParseVersion(moduleInfo.MinVersion)
			if err != nil {
				return false, fmt.Errorf("failed to parse module %s minimum version: %w", moduleInfo.ID, err)
			}

			if moduleVersion.LessThan(&minVersion) {
				return false, fmt.Errorf("current module %s version %s is less than minimum required version %s",
					moduleInfo.ID, moduleVersion.String(), minVersion.String())
			}
		}

		// Check module dependencies
		for _, dep := range moduleInfo.Dependencies {
			depVersion, hasDepVersion := currentVersions[dep.ID]
			if !hasDepVersion {
				return false, fmt.Errorf("module %s depends on %s, which is not installed",
					moduleInfo.ID, dep.ID)
			}

			minVersion, err := version.ParseVersion(dep.MinVersion)
			if err != nil {
				return false, fmt.Errorf("failed to parse dependency %s minimum version: %w", dep.ID, err)
			}
			if depVersion.LessThan(&minVersion) {
				return false, fmt.Errorf("current dependency %s version %s is less than minimum required version %s",
					dep.ID, depVersion.String(), minVersion.String())
			}
		}
	}

	// Check differential update compatibility
	if p.Manifest.PackageType == DifferentialPackage {
		// Check binary patches
		for _, patch := range p.Manifest.Components.Patches.Binary {
			fromVersion, err := version.ParseVersion(patch.FromVersion)
			if err != nil {
				return false, fmt.Errorf("failed to parse binary patch from version: %w", err)
			}

			if hasBinaryVersion && binaryVersion.String() != fromVersion.String() {
				return false, fmt.Errorf("binary patch requires version %s, but current version is %s",
					fromVersion.String(), binaryVersion.String())
			}
		}

		// Check templates patches
		for _, patch := range p.Manifest.Components.Patches.Templates {
			fromVersion, err := version.ParseVersion(patch.FromVersion)
			if err != nil {
				return false, fmt.Errorf("failed to parse templates patch from version: %w", err)
			}

			if hasTemplatesVersion && templatesVersion.String() != fromVersion.String() {
				return false, fmt.Errorf("templates patch requires version %s, but current version is %s",
					fromVersion.String(), templatesVersion.String())
			}
		}

		// Check module patches
		for _, patch := range p.Manifest.Components.Patches.Modules {
			moduleVersion, hasModuleVersion := currentVersions[fmt.Sprintf("module.%s", patch.ID)]
			if !hasModuleVersion {
				return false, fmt.Errorf("module patch for %s requires the module to be installed", patch.ID)
			}

			fromVersion, err := version.ParseVersion(patch.FromVersion)
			if err != nil {
				return false, fmt.Errorf("failed to parse module patch from version: %w", err)
			}

			if moduleVersion.String() != fromVersion.String() {
				return false, fmt.Errorf("module patch requires version %s, but current version is %s",
					fromVersion.String(), moduleVersion.String())
			}
		}
	}
	return true, nil
// HasRequiredUpdates checks if the package contains any required updates
func (p *UpdatePackage) HasRequiredUpdates() bool {
	if p.Manifest.Components.Binary.Required {
		return true
	}

	if p.Manifest.Components.Templates.Required {
		return true
	}

	for _, module := range p.Manifest.Components.Modules {
		if module.Required {
			return true
		}
	}

	return false

// ExtractFile extracts a file from the package to the given path
func (p *UpdatePackage) ExtractFile(filePath, destPath string) error {
	// Find file
	var file *zip.File
	for _, f := range p.reader.File {
		if f.Name == filePath {
			file = f
			break
		}
	}

	if file == nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { if err := src.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Create destination file
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { if err := dest.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Copy file contents
	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Set permissions
	if err := os.Chmod(destPath, file.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
	

// ExtractDirectory extracts a directory from the package to the given path
func (p *UpdatePackage) ExtractDirectory(dirPath, destPath string) error {
	// Find all files in directory
	var files []*zip.File
	prefix := dirPath + "/"
	for _, file := range p.reader.File {
		if strings.HasPrefix(file.Name, prefix) {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("directory not found: %s", dirPath)
	}

	// Create destination directory
	if err := os.MkdirAll(destPath, 0700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract each file
	for _, file := range files {
		// Skip directories
		if strings.HasSuffix(file.Name, "/") {
			// Create directory
			dirName := filepath.Join(destPath, strings.TrimPrefix(file.Name, prefix))
			if err := os.MkdirAll(dirName, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Get relative path
		relPath := strings.TrimPrefix(file.Name, prefix)
		destFilePath := filepath.Join(destPath, relPath)

		// Create parent directories
		destFileDir := filepath.Dir(destFilePath)
		if err := os.MkdirAll(destFileDir, 0700); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		// Open source file
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}

		// Create destination file
		dest, err := os.Create(destFilePath)
		if err != nil {
			src.Close()
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		// Copy file contents
		if _, err := io.Copy(dest, src); err != nil {
			src.Close()
			dest.Close()
			return fmt.Errorf("failed to copy file contents: %w", err)
		}

		// Close files
		src.Close()
		dest.Close()

		// Set permissions
		if err := os.Chmod(destFilePath, file.Mode()); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}
	}

	return nil

// GetBinaryPath returns the path to the binary in the package for the given platform
func (p *UpdatePackage) GetBinaryPath(platform string) string {
	binaryPath := fmt.Sprintf("binary/%s/LLMrecon", platform)
	if platform == "windows" {
		binaryPath += ".exe"
	}
	return binaryPath

// GetTemplatesPath returns the path to the templates in the package
func (p *UpdatePackage) GetTemplatesPath() string {
	return "templates"

// GetModulePath returns the path to a module in the package
func (p *UpdatePackage) GetModulePath(moduleID string) string {
	return fmt.Sprintf("modules/%s", moduleID)

// GetBinaryPatchPath returns the path to a binary patch in the package
func (p *UpdatePackage) GetBinaryPatchPath(platform, fromVersion, toVersion string) string {
	return fmt.Sprintf("patches/binary/%s/%s-%s.patch", platform, fromVersion, toVersion)

// GetTemplatesPatchPath returns the path to a templates patch in the package
func (p *UpdatePackage) GetTemplatesPatchPath(fromVersion, toVersion string) string {
	return fmt.Sprintf("patches/templates/%s-%s.patch", fromVersion, toVersion)

// GetModulePatchPath returns the path to a module patch in the package
func (p *UpdatePackage) GetModulePatchPath(moduleID, fromVersion, toVersion string) string {
	return fmt.Sprintf("patches/modules/%s/%s-%s.patch", moduleID, fromVersion, toVersion)

// CreatePackage creates a new update package with the given manifest
func CreatePackage(manifestPath, outputPath string) error {
	// Read manifest
	manifestData, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest
	var manifest PackageManifest
	err = json.Unmarshal(manifestData, &manifest)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { if err := outputFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Create zip writer
	zipWriter := zip.NewWriter(outputFile)
	defer func() { if err := zipWriter.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Add manifest
	manifestWriter, err := zipWriter.Create("manifest.json")
	if err != nil {
		return fmt.Errorf("failed to create manifest entry: %w", err)
	}

	if _, err := manifestWriter.Write(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// TODO: Add files to package based on manifest

}
}
}
}
}
}
}
}
}
}
}
}
}
}
