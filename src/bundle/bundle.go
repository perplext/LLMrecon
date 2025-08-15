// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"archive/zip"
	"encoding/json"
	"fmt"
)

// OpenBundle opens a bundle from the given path
func OpenBundle(path string) (*Bundle, error) {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle path does not exist: %w", err)
	}

	// Create bundle
	bundle := &Bundle{
		BundlePath: path,
		IsVerified: false,
	}

	// Read manifest
	err := bundle.readManifest()
	if err != nil {
		return nil, err
	}

	return bundle, nil

// readManifest reads the manifest from the bundle
func (b *Bundle) readManifest() error {
	// Read manifest file
	manifestPath := filepath.Join(b.BundlePath, "manifest.json")
	manifestData, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Unmarshal manifest
	err = json.Unmarshal(manifestData, &b.Manifest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return nil
	

// CreateBundle creates a new bundle with the given manifest and content
func CreateBundle(manifest BundleManifest, contentDir, outputPath string) (*Bundle, error) {
	// Create temporary directory for bundle
	tempDir, err := os.MkdirTemp("", "bundle-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create bundle directory
	bundleDir := filepath.Join(tempDir, "bundle")
	err = os.Mkdir(bundleDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle directory: %w", err)
	}

	// Copy content to bundle directory
	for _, item := range manifest.Content {
		srcPath := filepath.Join(contentDir, item.Path)
		dstPath := filepath.Join(bundleDir, item.Path)

		// Create parent directories
		err = os.MkdirAll(filepath.Dir(dstPath), 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", item.Path, err)
		}

		// Copy file or directory
		err = copyPath(srcPath, dstPath)
		if err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", item.Path, err)
		}
	}

	// Write manifest to bundle directory
	manifestPath := filepath.Join(bundleDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	err = os.WriteFile(filepath.Clean(manifestPath, manifestData, 0600))
	if err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Create zip file
	err = createZipFromDir(bundleDir, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}
	// Open the created bundle
	return OpenBundle(outputPath)

// copyPath copies a file or directory from src to dst
func copyPath(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		// Copy directory
		return copyDir(src, dst)
	}
	
	// Copy file
	return copyFile(src, dst)

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer func() { if err := srcFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { if err := dstFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())

// copyDir copies a directory from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil

// createZipFromDir creates a zip file from a directory
func createZipFromDir(src, dst string) error {
	// Create destination file
	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { if err := zipFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer func() { if err := zipWriter.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Walk the directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}
		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		// Create writer for the file
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it's a directory, just return
		if info.IsDir() {
			return nil
		}

		// Open the file
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

		// Copy file contents to zip
		_, err = io.Copy(writer, file)
		return err
	})

// ExtractBundle extracts a bundle to the given directory
func ExtractBundle(bundlePath, outputDir string) error {
	// Open the zip file
	reader, err := zip.OpenReader(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to open bundle: %w", err)
	}
	defer func() { if err := reader.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Create output directory if it doesn't exist
	err = os.MkdirAll(outputDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	// Extract files
	for _, file := range reader.File {
		err := extractFile(file, outputDir)
		if err != nil {
			return fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}
	}

	return nil

// extractFile extracts a file from a zip archive
func extractFile(file *zip.File, outputDir string) error {
	// Create the file path
	filePath := filepath.Join(outputDir, file.Name)

	// Check for directory traversal
	if !isWithinDir(outputDir, filePath) {
		return fmt.Errorf("illegal file path: %s", file.Name)
	}

	// Create directory for file if needed
	if file.FileInfo().IsDir() {
		return os.MkdirAll(filePath, file.Mode())
	}

	// Create parent directory if needed
	err := os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		return err
	}

	// Open the file
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { if err := rc.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Create the file
	outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer func() { if err := outFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Copy the file
	_, err = io.Copy(outFile, rc)
	return err

// isWithinDir checks if a path is within a directory
func isWithinDir(dir, path string) bool {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	return absPath == absDir || filepath.HasPrefix(absPath, absDir+string(filepath.Separator))

// GetContentPath returns the path to a content item in the bundle
func (b *Bundle) GetContentPath(id string) string {
	for _, item := range b.Manifest.Content {
		if item.ID == id {
			return filepath.Join(b.BundlePath, item.Path)
		}
	}
	return ""

// GetContentItem returns a content item by ID
func (b *Bundle) GetContentItem(id string) *ContentItem {
	for _, item := range b.Manifest.Content {
		if item.ID == id {
			return &item
		}
	}
	return nil

// GetContentItemsByType returns content items by type
func (b *Bundle) GetContentItemsByType(contentType ContentType) []ContentItem {
	var items []ContentItem
	for _, item := range b.Manifest.Content {
		if item.Type == contentType {
			items = append(items, item)
		}
	}
	return items

// CreateBundleManifest creates a new bundle manifest
func CreateBundleManifest(name, description, version string, bundleType BundleType, author Author) BundleManifest {
	return BundleManifest{
		SchemaVersion: "1.0",
		BundleID:      fmt.Sprintf("%s-%s-%d", name, version, time.Now().Unix()),
		BundleType:    bundleType,
		Name:          name,
		Description:   description,
		Version:       version,
		CreatedAt:     time.Now().UTC(),
		Author:        author,
		Content:       []ContentItem{},
		Checksums: Checksums{
			Content: make(map[string]string),
		},
		Compatibility: Compatibility{
			MinVersion: "1.0.0",
		},
	}

// AddContentItem adds a content item to the manifest
func (m *BundleManifest) AddContentItem(path string, contentType ContentType, id, version, description string) {
	item := ContentItem{
		Path:        path,
		Type:        contentType,
		ID:          id,
		Version:     version,
		Description: description,
		Checksum:    "", // Will be calculated during bundle creation
	}
