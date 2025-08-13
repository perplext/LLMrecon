package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

// ExportOptions defines options for bundle export
type ExportOptions struct {
	OutputPath      string                 // Path to save the bundle
	Format          ExportFormat           // Bundle format (tar.gz, zip, etc.)
	IncludeBinary   bool                   // Include the tool binary
	IncludeTemplates bool                  // Include templates
	IncludeModules  bool                   // Include modules
	IncludeDocs     bool                   // Include documentation
	Compression     CompressionType        // Compression type
	Encryption      *EncryptionOptions     // Optional encryption
	Filters         *ExportFilters         // Filters for selective export
	ProgressHandler ExportProgressHandler        // Progress callback
	Metadata        map[string]interface{} // Additional metadata
}

// ExportFormat defines the bundle format
type ExportFormat string

const (
	FormatTarGz ExportFormat = "tar.gz"
	FormatZip   ExportFormat = "zip"
	FormatTar   ExportFormat = "tar"
)

// CompressionType defines compression options
type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionGzip   CompressionType = "gzip"
	CompressionZstd   CompressionType = "zstd"
	CompressionBrotli CompressionType = "brotli"
)

// EncryptionOptions defines encryption settings
type EncryptionOptions struct {
	Algorithm string // aes-256-gcm, chacha20-poly1305
	Password  string // Encryption password
	KeyFile   string // Path to key file
}

// ExportFilters defines filters for selective export
type ExportFilters struct {
	TemplateCategories []string    // Filter templates by category
	ModuleTypes        []string    // Filter modules by type
	MinVersion         string      // Minimum version to include
	MaxVersion         string      // Maximum version to include
	CreatedAfter       *time.Time  // Include items created after this date
	CreatedBefore      *time.Time  // Include items created before this date
	Tags               []string    // Filter by tags
	ExcludePatterns    []string    // Glob patterns to exclude
	IncludeList        []string    // Explicit list of paths to include
	ExcludeList        []string    // Explicit list of paths to exclude
}

// ExportProgressHandler is called to report export progress
type ExportProgressHandler func(progress ProgressInfo)

// ProgressInfo contains progress information
type ProgressInfo struct {
	Stage           string  // Current stage
	Current         int     // Current item
	Total           int     // Total items
	Percentage      float64 // Completion percentage
	CurrentFile     string  // Current file being processed
	BytesProcessed  int64   // Bytes processed
	TotalBytes      int64   // Total bytes to process
	TimeElapsed     time.Duration
	TimeRemaining   time.Duration
	Message         string  // Status message
}

// BundleExporter handles bundle export operations
type BundleExporter struct {
	options         ExportOptions
	manifest        *BundleManifest
	tempDir         string
	startTime       time.Time
	bytesProcessed  int64
	totalBytes      int64
	currentStage    string
	errors          []error
}

// NewBundleExporter creates a new bundle exporter
func NewBundleExporter(options ExportOptions) *BundleExporter {
	return &BundleExporter{
		options:   options,
		startTime: time.Now(),
		errors:    []error{},
	}
}

// Export creates a bundle with the specified options
func (e *BundleExporter) Export() error {
	// Initialize
	if err := e.initialize(); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}
	defer e.cleanup()

	// Stage 1: Prepare manifest
	e.reportProgress("Preparing manifest", 0, 100)
	if err := e.prepareManifest(); err != nil {
		return fmt.Errorf("manifest preparation failed: %w", err)
	}

	// Stage 2: Collect components
	e.reportProgress("Collecting components", 10, 100)
	if err := e.collectComponents(); err != nil {
		return fmt.Errorf("component collection failed: %w", err)
	}

	// Stage 3: Generate signatures
	e.reportProgress("Generating signatures", 70, 100)
	if err := e.generateSignatures(); err != nil {
		return fmt.Errorf("signature generation failed: %w", err)
	}

	// Stage 4: Create bundle archive
	e.reportProgress("Creating bundle archive", 80, 100)
	if err := e.createArchive(); err != nil {
		return fmt.Errorf("archive creation failed: %w", err)
	}

	// Stage 5: Apply encryption if requested
	if e.options.Encryption != nil {
		e.reportProgress("Encrypting bundle", 95, 100)
		if err := e.encryptBundle(); err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}
	}

	e.reportProgress("Export complete", 100, 100)
	return nil
}

// initialize sets up the export process
func (e *BundleExporter) initialize() error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "bundle-export-*")
	if err != nil {
		return err
	}
	e.tempDir = tempDir

	// Initialize manifest
	e.manifest = &BundleManifest{
		SchemaVersion: "1.0",
		Version:       "1.0.0", // Default version
		CreatedAt:     time.Now().UTC(),
		Author: Author{
			Name:  os.Getenv("USER"),
			Email: "",
		},
		Content: []ContentItem{},
		Checksums: Checksums{
			Content: make(map[string]string),
		},
	}

	// Set default options
	if e.options.Format == "" {
		e.options.Format = FormatTarGz
	}
	if e.options.Compression == "" {
		e.options.Compression = CompressionGzip
	}

	return nil
}

// cleanup removes temporary files
func (e *BundleExporter) cleanup() {
	if e.tempDir != "" {
		os.RemoveAll(e.tempDir)
	}
}

// prepareManifest prepares the bundle manifest
func (e *BundleExporter) prepareManifest() error {
	// Add bundle metadata
	e.manifest.BundleID = generateBundleID()
	e.manifest.Name = fmt.Sprintf("llm-redteam-bundle-%s", time.Now().Format("20060102-150405"))
	e.manifest.Description = "LLMrecon offline bundle"
	e.manifest.BundleType = MixedBundleType // Default to mixed type

	// Add compatibility information
	e.manifest.Compatibility = Compatibility{
		MinVersion:   "1.0.0",
		MaxVersion:   "",
		Dependencies: []string{},
		Incompatible: []string{},
	}

	return nil
}

// collectComponents collects all components for the bundle
func (e *BundleExporter) collectComponents() error {
	totalSteps := 0
	if e.options.IncludeBinary {
		totalSteps++
	}
	if e.options.IncludeTemplates {
		totalSteps++
	}
	if e.options.IncludeModules {
		totalSteps++
	}
	if e.options.IncludeDocs {
		totalSteps++
	}

	currentStep := 0

	// Collect binary
	if e.options.IncludeBinary {
		e.reportProgress("Collecting binary", currentStep*25, 100)
		if err := e.collectBinary(); err != nil {
			return fmt.Errorf("failed to collect binary: %w", err)
		}
		currentStep++
	}

	// Collect templates
	if e.options.IncludeTemplates {
		e.reportProgress("Collecting templates", currentStep*25, 100)
		if err := e.collectTemplates(); err != nil {
			return fmt.Errorf("failed to collect templates: %w", err)
		}
		currentStep++
	}

	// Collect modules
	if e.options.IncludeModules {
		e.reportProgress("Collecting modules", currentStep*25, 100)
		if err := e.collectModules(); err != nil {
			return fmt.Errorf("failed to collect modules: %w", err)
		}
		currentStep++
	}

	// Collect documentation
	if e.options.IncludeDocs {
		e.reportProgress("Collecting documentation", currentStep*25, 100)
		if err := e.collectDocumentation(); err != nil {
			return fmt.Errorf("failed to collect documentation: %w", err)
		}
		currentStep++
	}

	return nil
}

// collectBinary collects the tool binary
func (e *BundleExporter) collectBinary() error {
	// Find the current executable
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Copy to bundle directory
	destPath := filepath.Join(e.tempDir, "binary", filepath.Base(execPath))
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	if err := copyFileWithProgress(execPath, destPath, e.updateFileProgress); err != nil {
		return err
	}

	// Add to manifest
	checksum, _ := e.calculateChecksum(destPath)
	
	e.manifest.Content = append(e.manifest.Content, ContentItem{
		Path:        filepath.Join("binary", filepath.Base(execPath)),
		Type:        "binary",
		Checksum:    checksum,
		Version:     "1.0.0", // Using default version
		Description: fmt.Sprintf("Binary for %s/%s", getOS(), getArch()),
	})

	return nil
}

// collectTemplates collects security test templates
func (e *BundleExporter) collectTemplates() error {
	templateDir := getTemplateDirectory()
	
	// Apply filters if specified
	templates, err := e.findTemplates(templateDir)
	if err != nil {
		return err
	}

	e.reportProgress(fmt.Sprintf("Found %d templates", len(templates)), 0, 0)

	// Copy templates
	for i, tmplPath := range templates {
		relPath, _ := filepath.Rel(templateDir, tmplPath)
		destPath := filepath.Join(e.tempDir, "templates", relPath)
		
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if err := copyFileWithProgress(tmplPath, destPath, e.updateFileProgress); err != nil {
			return err
		}

		// Parse template metadata
		tmplData, _ := os.ReadFile(tmplPath)
		metadata := e.parseTemplateMetadata(tmplData)

		// Add to manifest
		checksum, _ := e.calculateChecksum(destPath)
		
		versionStr := ""
		if v, ok := metadata["version"].(string); ok {
			versionStr = v
		}
		
		e.manifest.Content = append(e.manifest.Content, ContentItem{
			Path:        filepath.Join("templates", relPath),
			Type:        TemplateContentType,
			Checksum:    checksum,
			Version:     versionStr,
			Description: fmt.Sprintf("Template: %s", filepath.Base(tmplPath)),
		})

		e.reportProgress(fmt.Sprintf("Collected template %d/%d", i+1, len(templates)), 0, 0)
	}

	return nil
}

// collectModules collects plugin modules
func (e *BundleExporter) collectModules() error {
	moduleDir := getModuleDirectory()
	
	modules, err := e.findModules(moduleDir)
	if err != nil {
		return err
	}

	for i, modPath := range modules {
		relPath, _ := filepath.Rel(moduleDir, modPath)
		destPath := filepath.Join(e.tempDir, "modules", relPath)
		
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if err := copyFileWithProgress(modPath, destPath, e.updateFileProgress); err != nil {
			return err
		}

		// Add to manifest
		checksum, _ := e.calculateChecksum(destPath)
		
		e.manifest.Content = append(e.manifest.Content, ContentItem{
			Path:        filepath.Join("modules", relPath),
			Type:        ModuleContentType,
			Checksum:    checksum,
			Description: fmt.Sprintf("Module: %s", e.getModuleType(modPath)),
		})

		e.reportProgress(fmt.Sprintf("Collected module %d/%d", i+1, len(modules)), 0, 0)
	}

	return nil
}

// collectDocumentation collects documentation files
func (e *BundleExporter) collectDocumentation() error {
	docDir := getDocumentationDirectory()
	
	// Find all documentation files
	var docs []string
	err := filepath.Walk(docDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && e.isDocumentationFile(path) {
			docs = append(docs, path)
		}
		return nil
	})
	
	if err != nil {
		return err
	}

	// Copy documentation
	for i, docPath := range docs {
		relPath, _ := filepath.Rel(docDir, docPath)
		destPath := filepath.Join(e.tempDir, "documentation", relPath)
		
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if err := copyFileWithProgress(docPath, destPath, e.updateFileProgress); err != nil {
			return err
		}

		// Add to manifest
		checksum, _ := e.calculateChecksum(destPath)
		
		e.manifest.Content = append(e.manifest.Content, ContentItem{
			Path:        filepath.Join("documentation", relPath),
			Type:        "documentation",
			Checksum:    checksum,
			Description: fmt.Sprintf("Documentation: %s", filepath.Base(docPath)),
		})

		e.reportProgress(fmt.Sprintf("Collected documentation %d/%d", i+1, len(docs)), 0, 0)
	}

	return nil
}

// generateSignatures generates digital signatures for the bundle
func (e *BundleExporter) generateSignatures() error {
	// Save manifest
	manifestPath := filepath.Join(e.tempDir, "manifest.json")
	manifestData, err := json.MarshalIndent(e.manifest, "", "  ")
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return err
	}

	// Calculate manifest checksum
	checksum, err := e.calculateChecksum(manifestPath)
	if err != nil {
		return err
	}
	e.manifest.Checksums.Manifest = checksum

	// Generate bundle signature
	if privateKey := e.getSigningKey(); privateKey != nil {
		privateKeyBytes, ok := privateKey.(ed25519.PrivateKey)
		if !ok {
			return fmt.Errorf("invalid private key type")
		}
		
		signer := NewSigner(privateKeyBytes, "bundle-signing-key", SignatureMetadata{
			Signer:      e.manifest.Author.Name,
			Environment: "production",
			BuildID:     e.manifest.BundleID,
		})

		signature, err := signer.SignBundle(e.tempDir)
		if err != nil {
			return fmt.Errorf("failed to sign bundle: %w", err)
		}

		if err := SaveSignature(e.tempDir, signature); err != nil {
			return fmt.Errorf("failed to save signature: %w", err)
		}

		e.manifest.Signature = signature.Signature
	}

	// Update manifest with signature
	manifestData, _ = json.MarshalIndent(e.manifest, "", "  ")
	return os.WriteFile(manifestPath, manifestData, 0644)
}

// createArchive creates the final bundle archive
func (e *BundleExporter) createArchive() error {
	switch e.options.Format {
	case FormatTarGz:
		return e.createTarGzArchive()
	case FormatZip:
		return e.createZipArchive()
	case FormatTar:
		return e.createTarArchive()
	default:
		return fmt.Errorf("unsupported format: %s", e.options.Format)
	}
}

// createTarGzArchive creates a tar.gz archive
func (e *BundleExporter) createTarGzArchive() error {
	// Create output file
	outputFile, err := os.Create(e.options.OutputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outputFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through temp directory and add files
	return filepath.Walk(e.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update header name to be relative to temp dir
		relPath, _ := filepath.Rel(e.tempDir, path)
		if relPath == "." {
			return nil
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		e.updateFileProgress(path, info.Size())
		return nil
	})
}

// Helper methods

func (e *BundleExporter) reportProgress(stage string, current, total int) {
	if e.options.ProgressHandler == nil {
		return
	}

	elapsed := time.Since(e.startTime)
	percentage := float64(current) / float64(total) * 100
	if total == 0 {
		percentage = 0
	}

	info := ProgressInfo{
		Stage:          stage,
		Current:        current,
		Total:          total,
		Percentage:     percentage,
		BytesProcessed: e.bytesProcessed,
		TotalBytes:     e.totalBytes,
		TimeElapsed:    elapsed,
		Message:        stage,
	}

	// Estimate time remaining
	if percentage > 0 {
		totalTime := elapsed.Seconds() / (percentage / 100)
		remaining := totalTime - elapsed.Seconds()
		info.TimeRemaining = time.Duration(remaining) * time.Second
	}

	e.options.ProgressHandler(info)
}

func (e *BundleExporter) updateFileProgress(path string, size int64) {
	e.bytesProcessed += size
	e.reportProgress(e.currentStage, int(e.bytesProcessed), int(e.totalBytes))
}

func (e *BundleExporter) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func (e *BundleExporter) getIncludeFlags() []string {
	var flags []string
	if e.options.IncludeBinary {
		flags = append(flags, "binary")
	}
	if e.options.IncludeTemplates {
		flags = append(flags, "templates")
	}
	if e.options.IncludeModules {
		flags = append(flags, "modules")
	}
	if e.options.IncludeDocs {
		flags = append(flags, "documentation")
	}
	return flags
}

func (e *BundleExporter) findTemplates(dir string) ([]string, error) {
	var templates []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			// Apply filters
			if e.shouldIncludeTemplate(path) {
				templates = append(templates, path)
			}
		}
		return nil
	})
	
	return templates, err
}

func (e *BundleExporter) shouldIncludeTemplate(path string) bool {
	if e.options.Filters == nil {
		return true
	}

	// Check exclude patterns
	for _, pattern := range e.options.Filters.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return false
		}
	}

	// Check category filters
	if len(e.options.Filters.TemplateCategories) > 0 {
		// Parse template to check category
		data, _ := os.ReadFile(path)
		metadata := e.parseTemplateMetadata(data)
		category, _ := metadata["category"].(string)
		
		found := false
		for _, cat := range e.options.Filters.TemplateCategories {
			if cat == category {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (e *BundleExporter) parseTemplateMetadata(data []byte) map[string]interface{} {
	// Simple YAML front matter parsing
	metadata := make(map[string]interface{})
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "id:") {
			metadata["id"] = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "category:") {
			metadata["category"] = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
		} else if strings.HasPrefix(line, "version:") {
			metadata["version"] = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}
	}
	
	return metadata
}

func (e *BundleExporter) findModules(dir string) ([]string, error) {
	var modules []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && (strings.HasSuffix(path, ".so") || strings.HasSuffix(path, ".dll")) {
			modules = append(modules, path)
		}
		return nil
	})
	
	return modules, err
}

func (e *BundleExporter) getModuleType(path string) string {
	if strings.Contains(path, "provider") {
		return "provider"
	} else if strings.Contains(path, "detector") {
		return "detector"
	} else if strings.Contains(path, "util") {
		return "utility"
	}
	return "unknown"
}

func (e *BundleExporter) isDocumentationFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".txt" || ext == ".pdf" || ext == ".html"
}

func (e *BundleExporter) getSigningKey() interface{} {
	// TODO: Implement key management
	return nil
}

func (e *BundleExporter) encryptBundle() error {
	// TODO: Implement encryption
	return nil
}

// Utility functions

func generateBundleID() string {
	return fmt.Sprintf("BDL-%d-%s", time.Now().Unix(), generateRandomString(8))
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func copyFileWithProgress(src, dst string, progressFunc func(string, int64)) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy with progress updates
	buf := make([]byte, 32*1024)
	var written int64
	
	for {
		n, err := sourceFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destFile.Write(buf[:n]); err != nil {
			return err
		}
		
		written += int64(n)
		if progressFunc != nil && written%1048576 == 0 { // Update every MB
			progressFunc(src, written)
		}
	}

	// Copy file permissions
	srcInfo, _ := os.Stat(src)
	return os.Chmod(dst, srcInfo.Mode())
}

func getOS() string {
	return os.Getenv("GOOS")
}

func getArch() string {
	return os.Getenv("GOARCH")
}

func getTemplateDirectory() string {
	// TODO: Make configurable
	return "./templates"
}

func getModuleDirectory() string {
	// TODO: Make configurable
	return "./modules"
}

func getDocumentationDirectory() string {
	// TODO: Make configurable
	return "./docs"
}

// createZipArchive creates a ZIP archive (stub for now)
func (e *BundleExporter) createZipArchive() error {
	return fmt.Errorf("ZIP format not yet implemented")
}

// createTarArchive creates a plain TAR archive (stub for now)
func (e *BundleExporter) createTarArchive() error {
	return fmt.Errorf("plain TAR format not yet implemented")
}