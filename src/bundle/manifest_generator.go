// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// ManifestGenerator generates bundle manifests
type ManifestGenerator struct {
	// SigningKey is the key used to sign bundles
	SigningKey ed25519.PrivateKey
	// Author is the default author for generated manifests
	Author Author
}

// NewManifestGenerator creates a new manifest generator
func NewManifestGenerator(signingKey ed25519.PrivateKey, author Author) *ManifestGenerator {
	return &ManifestGenerator{
		SigningKey: signingKey,
		Author:     author,
	}
}

// GenerateManifest generates a bundle manifest
func (g *ManifestGenerator) GenerateManifest(name, description, version string, bundleType BundleType) *BundleManifest {
	return &BundleManifest{
		SchemaVersion: "1.0",
		BundleID:      uuid.New().String(),
		BundleType:    bundleType,
		Name:          name,
		Description:   description,
		Version:       version,
		CreatedAt:     time.Now().UTC(),
		Author:        g.Author,
		Content:       []ContentItem{},
		Checksums: Checksums{
			Manifest: "",
			Content:  make(map[string]string),
		},
		Compatibility: Compatibility{
			MinVersion:   "1.0.0",
			Dependencies: []string{},
			Incompatible: []string{},
		},
	}
}

// GenerateEnhancedManifest generates an enhanced bundle manifest for offline bundles
func (g *ManifestGenerator) GenerateEnhancedManifest(name, description, version string, bundleType BundleType) *EnhancedBundleManifest {
	baseManifest := g.GenerateManifest(name, description, version, bundleType)
	
	enhancedManifest := &EnhancedBundleManifest{
		BundleManifest: *baseManifest,
	}
	
	// Initialize compliance mappings
	enhancedManifest.Compliance.OwaspLLMTop10 = make(map[string][]string)
	enhancedManifest.Compliance.ISOIEC42001 = make(map[string][]string)
	
	// Initialize changelog
	enhancedManifest.Changelog = []ChangelogEntry{
		{
			Version: version,
			Date:    time.Now().UTC(),
			Changes: []string{
				"Initial release",
			},
		},
	}
	
	// Initialize documentation
	enhancedManifest.Documentation = make(map[string]string)
	
	return enhancedManifest
}

// GenerateIncrementalManifest generates an incremental bundle manifest
func (g *ManifestGenerator) GenerateIncrementalManifest(baseManifest *EnhancedBundleManifest, newVersion string, changes []string) *EnhancedBundleManifest {
	// Create a copy of the base manifest
	manifestData, _ := json.Marshal(baseManifest)
	var incrementalManifest EnhancedBundleManifest
	json.Unmarshal(manifestData, &incrementalManifest)
	
	// Update version and creation time
	incrementalManifest.Version = newVersion
	incrementalManifest.CreatedAt = time.Now().UTC()
	
	// Set incremental flag and base version
	incrementalManifest.IsIncremental = true
	incrementalManifest.BaseVersion = baseManifest.Version
	
	// Add changelog entry
	incrementalManifest.Changelog = append(incrementalManifest.Changelog, ChangelogEntry{
		Version: newVersion,
		Date:    time.Now().UTC(),
		Changes: changes,
	})
	
	// Clear content and checksums (will be populated later)
	incrementalManifest.Content = []ContentItem{}
	incrementalManifest.Checksums.Content = make(map[string]string)
	
	return &incrementalManifest
}

// AddContentItem adds a content item to a manifest
func (g *ManifestGenerator) AddContentItem(manifest *BundleManifest, path string, contentType ContentType, id, version, description string) {
	// Generate ID if not provided
	if id == "" {
		id = generateContentID(path, contentType)
	}
	
	// Create content item
	item := ContentItem{
		Path:        path,
		Type:        contentType,
		ID:          id,
		Version:     version,
		Description: description,
		Checksum:    "",
		BundleID:    manifest.BundleID,
	}
	
	// Add to manifest
	manifest.Content = append(manifest.Content, item)
}

// AddContentItemToEnhancedManifest adds a content item to an enhanced manifest
func (g *ManifestGenerator) AddContentItemToEnhancedManifest(manifest *EnhancedBundleManifest, path string, contentType ContentType, id, version, description string) {
	g.AddContentItem(&manifest.BundleManifest, path, contentType, id, version, description)
}

// AddComplianceMapping adds a compliance mapping to an enhanced manifest
func (g *ManifestGenerator) AddComplianceMapping(manifest *EnhancedBundleManifest, contentID string, owaspCategories, isoControls []string) error {
	// Verify that the content ID exists
	contentExists := false
	for _, item := range manifest.Content {
		if item.ID == contentID {
			contentExists = true
			break
		}
	}
	
	if !contentExists {
		return fmt.Errorf("content ID %s does not exist in the manifest", contentID)
	}
	
	// Add OWASP LLM Top 10 mappings
	for _, category := range owaspCategories {
		if manifest.Compliance.OwaspLLMTop10 == nil {
			manifest.Compliance.OwaspLLMTop10 = make(map[string][]string)
		}
		
		// Check if content ID is already mapped to this category
		exists := false
		for _, id := range manifest.Compliance.OwaspLLMTop10[category] {
			if id == contentID {
				exists = true
				break
			}
		}
		
		// Add mapping if it doesn't exist
		if !exists {
			manifest.Compliance.OwaspLLMTop10[category] = append(
				manifest.Compliance.OwaspLLMTop10[category], contentID)
		}
	}
	
	// Add ISO/IEC 42001 mappings
	for _, control := range isoControls {
		if manifest.Compliance.ISOIEC42001 == nil {
			manifest.Compliance.ISOIEC42001 = make(map[string][]string)
		}
		
		// Check if content ID is already mapped to this control
		exists := false
		for _, id := range manifest.Compliance.ISOIEC42001[control] {
			if id == contentID {
				exists = true
				break
			}
		}
		
		// Add mapping if it doesn't exist
		if !exists {
			manifest.Compliance.ISOIEC42001[control] = append(
				manifest.Compliance.ISOIEC42001[control], contentID)
		}
	}
	
	return nil
}

// AddDocumentation adds a documentation file to an enhanced manifest
func (g *ManifestGenerator) AddDocumentation(manifest *EnhancedBundleManifest, docType, path string) {
	if manifest.Documentation == nil {
		manifest.Documentation = make(map[string]string)
	}
	
	manifest.Documentation[docType] = path
}

// UpdateChecksums updates the checksums in a manifest based on content in a directory
func (g *ManifestGenerator) UpdateChecksums(manifest *BundleManifest, contentDir string) error {
	// Calculate checksums for content items
	for i, item := range manifest.Content {
		itemPath := filepath.Join(contentDir, item.Path)
		
		// Check if path exists
		if _, err := os.Stat(itemPath); os.IsNotExist(err) {
			return fmt.Errorf("content item path does not exist: %s", itemPath)
		}
		
		// Calculate checksum
		checksum, err := calculateFileChecksum(itemPath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %w", itemPath, err)
		}
		
		// Update checksum in content item
		manifest.Content[i].Checksum = checksum
		
		// Add to checksums map
		manifest.Checksums.Content[item.Path] = checksum
	}
	
	// Calculate manifest checksum (excluding the checksums field)
	manifestCopy := *manifest
	manifestCopy.Checksums.Manifest = ""
	manifestCopy.Checksums.Content = nil
	manifestCopy.Signature = ""
	
	manifestData, err := json.Marshal(manifestCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for checksum calculation: %w", err)
	}
	
	manifestChecksum := calculateChecksum(manifestData)
	manifest.Checksums.Manifest = manifestChecksum
	
	return nil
}

// UpdateChecksumsForEnhancedManifest updates the checksums in an enhanced manifest
func (g *ManifestGenerator) UpdateChecksumsForEnhancedManifest(manifest *EnhancedBundleManifest, contentDir string) error {
	return g.UpdateChecksums(&manifest.BundleManifest, contentDir)
}

// SignManifest signs a manifest using the signing key
func (g *ManifestGenerator) SignManifest(manifest *BundleManifest) error {
	if g.SigningKey == nil {
		return fmt.Errorf("signing key is not set")
	}
	
	// Create a copy of the manifest without the signature
	manifestCopy := *manifest
	manifestCopy.Signature = ""
	
	// Marshal the manifest
	manifestData, err := json.Marshal(manifestCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for signing: %w", err)
	}
	
	// Sign the manifest
	signature := ed25519.Sign(g.SigningKey, manifestData)
	
	// Set the signature
	manifest.Signature = base64.StdEncoding.EncodeToString(signature)
	
	return nil
}

// SignEnhancedManifest signs an enhanced manifest
func (g *ManifestGenerator) SignEnhancedManifest(manifest *EnhancedBundleManifest) error {
	return g.SignManifest(&manifest.BundleManifest)
}

// WriteManifest writes a manifest to a file
func (g *ManifestGenerator) WriteManifest(manifest *BundleManifest, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal manifest
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}
	
	return nil
}

// WriteEnhancedManifest writes an enhanced manifest to a file
func (g *ManifestGenerator) WriteEnhancedManifest(manifest *EnhancedBundleManifest, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal manifest
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}
	
	return nil
}

// GenerateManifestKeyPair generates a new Ed25519 key pair for manifest signing
func GenerateManifestKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// generateContentID generates a content ID based on path and type
func generateContentID(path string, contentType ContentType) string {
	// Create a hash of the path and content type
	hasher := sha256.New()
	hasher.Write([]byte(path))
	hasher.Write([]byte(string(contentType)))
	hash := hasher.Sum(nil)
	
	// Use first 8 bytes of hash as ID
	return fmt.Sprintf("%s-%x", contentType, hash[:8])
}

// calculateFileChecksum calculates the SHA-256 checksum of a file
func calculateFileChecksum(filePath string) (string, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	return calculateChecksum(data), nil
}

// calculateChecksum calculates the SHA-256 checksum of data
func calculateChecksum(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return fmt.Sprintf("sha256:%x", hasher.Sum(nil))
}
