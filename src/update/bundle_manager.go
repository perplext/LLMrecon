package update

import (
	"archive/zip"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// BundleManager handles offline update bundles
type BundleManager struct {
	config   *Config
	verifier *Verifier
	logger   Logger
}

// NewBundleManager creates a new bundle manager
func NewBundleManager(config *Config, logger Logger) *BundleManager {
	return &BundleManager{
		config:   config,
		verifier: NewVerifier(config, logger),
		logger:   logger,
	}

// ExportBundle exports an offline update bundle
func (bm *BundleManager) ExportBundle(options *BundleExportOptions) (*BundleInfo, error) {
	bm.logger.Info("Creating offline update bundle...")
	
	// Create temporary workspace
	workspaceDir, err := os.MkdirTemp("", "bundle-export-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	defer os.RemoveAll(workspaceDir)
	
	// Create bundle metadata
	bundle := &Bundle{
		Metadata: BundleMetadata{
			Version:       options.Version,
			BundleType:    options.Type,
			Description:   options.Description,
			SourceVersion: options.SourceVersion,
			TargetVersion: options.TargetVersion,
			Platforms:     options.Platforms,
			Incremental:   options.Incremental,
		},
		CreatedBy:  options.CreatedBy,
		CreatedAt:  time.Now(),
		Components: make([]ComponentInfo, 0),
		Checksums:  make(map[string]string),
		Signatures: make(map[string]string),
	}
	
	if options.ExpiresAt != nil {
		bundle.ExpiresAt = options.ExpiresAt
	}
	
	// Collect components based on bundle type
	var totalSize int64
	
	if options.IncludeBinary {
		if err := bm.addBinaryToBundle(bundle, workspaceDir, &totalSize); err != nil {
			return nil, fmt.Errorf("failed to add binary: %w", err)
		}
	}
	
	if options.IncludeTemplates {
		if err := bm.addTemplatesToBundle(bundle, workspaceDir, &totalSize); err != nil {
			return nil, fmt.Errorf("failed to add templates: %w", err)
		}
	}
	
	if options.IncludeModules {
		if err := bm.addModulesToBundle(bundle, workspaceDir, &totalSize); err != nil {
			return nil, fmt.Errorf("failed to add modules: %w", err)
		}
	}
	
	// Create bundle archive
	bundleInfo, err := bm.createBundleArchive(bundle, workspaceDir, options.OutputPath, totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle archive: %w", err)
	}
	
	// Sign bundle if requested
	if options.SignBundle {
		if err := bm.signBundle(bundleInfo.Path, options.SigningKey); err != nil {
			bm.logger.Warn("Failed to sign bundle: " + err.Error())
		} else {
			bundleInfo.Signed = true
		}
	}
	
	bm.logger.Info(fmt.Sprintf("Bundle created successfully: %s (%s)", 
		bundleInfo.Path, FormatFileSize(bundleInfo.Size)))
	
	return bundleInfo, nil

// ImportBundle imports an offline update bundle
func (bm *BundleManager) ImportBundle(bundlePath string, options *BundleImportOptions) (*ImportResult, error) {
	bm.logger.Info(fmt.Sprintf("Importing bundle: %s", bundlePath))
	
	// Verify bundle exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle file not found: %s", bundlePath)
	}
	
	// Verify bundle integrity
	if options.VerifyIntegrity {
		if err := bm.verifyBundleIntegrity(bundlePath); err != nil {
			if !options.ForceImport {
				return nil, fmt.Errorf("bundle verification failed: %w", err)
			}
			bm.logger.Warn("Bundle verification failed, proceeding anyway due to force flag")
		}
	}
	
	// Create workspace for extraction
	workspaceDir, err := os.MkdirTemp("", "bundle-import-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	defer os.RemoveAll(workspaceDir)
	
	// Extract bundle
	bundle, err := bm.extractBundle(bundlePath, workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bundle: %w", err)
	}
	
	// Check bundle expiration
	if bundle.ExpiresAt != nil && time.Now().After(*bundle.ExpiresAt) {
		if !options.ForceImport {
			return nil, fmt.Errorf("bundle has expired on %s", bundle.ExpiresAt.Format("2006-01-02"))
		}
		bm.logger.Warn("Bundle has expired, proceeding anyway due to force flag")
	}
	
	// Check version compatibility
	if err := bm.checkVersionCompatibility(bundle, options); err != nil {
		if !options.ForceImport {
			return nil, fmt.Errorf("version compatibility check failed: %w", err)
		}
		bm.logger.Warn("Version compatibility check failed, proceeding anyway")
	}
	
	// Create backup if requested
	if options.CreateBackup {
		if err := bm.createImportBackup(); err != nil {
			bm.logger.Warn("Failed to create backup: " + err.Error())
		}
	}
	
	// Apply updates
	result, err := bm.applyBundleUpdates(bundle, workspaceDir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to apply updates: %w", err)
	}
	
	bm.logger.Info("Bundle import completed successfully")
	return result, nil

// addBinaryToBundle adds binary to the bundle
func (bm *BundleManager) addBinaryToBundle(bundle *Bundle, workspaceDir string, totalSize *int64) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Copy binary to workspace
	binaryName := filepath.Base(execPath)
	workspacePath := filepath.Join(workspaceDir, "binary", binaryName)
	
	if err := os.MkdirAll(filepath.Dir(workspacePath), 0700); err != nil {
		return err
	}
	
	if err := bm.copyFile(execPath, workspacePath); err != nil {
		return err
	}
	
	// Get file info
	info, err := os.Stat(workspacePath)
	if err != nil {
		return err
	}
	
	// Calculate checksum
	checksum, err := bm.calculateFileChecksum(workspacePath)
	if err != nil {
		return err
	}
	
	// Add to bundle
	component := ComponentInfo{
		Name:        "binary",
		Version:     bm.getCurrentBinaryVersion(),
		Type:        ComponentBinary,
		Path:        filepath.Join("binary", binaryName),
		Size:        info.Size(),
		Checksum:    checksum,
		Required:    true,
		Description: "Main executable binary",
	}
	
	bundle.Components = append(bundle.Components, component)
	bundle.Checksums[component.Path] = checksum
	*totalSize += info.Size()
	
	return nil

// addTemplatesToBundle adds templates to the bundle
func (bm *BundleManager) addTemplatesToBundle(bundle *Bundle, workspaceDir string, totalSize *int64) error {
	templateDir := bm.config.TemplateDirectory
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		bm.logger.Warn("Template directory not found, skipping templates")
		return nil
	}
	
	templatesWorkspaceDir := filepath.Join(workspaceDir, "templates")
	if err := os.MkdirAll(templatesWorkspaceDir, 0700); err != nil {
		return err
	}
	
	// Copy template files
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		
		relPath, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		
		workspacePath := filepath.Join(templatesWorkspaceDir, relPath)
		
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(workspacePath), 0700); err != nil {
			return err
		}
		
		// Copy file
		if err := bm.copyFile(path, workspacePath); err != nil {
			return err
		}
		
		// Calculate checksum
		checksum, err := bm.calculateFileChecksum(workspacePath)
		if err != nil {
			return err
		}
		
		// Add to bundle
		component := ComponentInfo{
			Name:        relPath,
			Version:     bm.getTemplateVersion(path),
			Type:        ComponentTemplates,
			Path:        filepath.Join("templates", relPath),
			Size:        info.Size(),
			Checksum:    checksum,
			Required:    false,
			Description: "Security test template",
		}
		
		bundle.Components = append(bundle.Components, component)
		bundle.Checksums[component.Path] = checksum
		*totalSize += info.Size()
		
		return nil
	})
	
	return err

// addModulesToBundle adds modules to the bundle
func (bm *BundleManager) addModulesToBundle(bundle *Bundle, workspaceDir string, totalSize *int64) error {
	moduleDir := bm.config.ModuleDirectory
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		bm.logger.Warn("Module directory not found, skipping modules")
		return nil
	}
	
	modulesWorkspaceDir := filepath.Join(workspaceDir, "modules")
	if err := os.MkdirAll(modulesWorkspaceDir, 0700); err != nil {
		return err
	}
	
	// Copy module files
	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		
		relPath, err := filepath.Rel(moduleDir, path)
		if err != nil {
			return err
		}
		
		workspacePath := filepath.Join(modulesWorkspaceDir, relPath)
		
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(workspacePath), 0700); err != nil {
			return err
		}
		
		// Copy file
		if err := bm.copyFile(path, workspacePath); err != nil {
			return err
		}
		
		// Calculate checksum
		checksum, err := bm.calculateFileChecksum(workspacePath)
		if err != nil {
			return err
		}
		
		// Add to bundle
		component := ComponentInfo{
			Name:        relPath,
			Version:     bm.getModuleVersion(path),
			Type:        ComponentModules,
			Path:        filepath.Join("modules", relPath),
			Size:        info.Size(),
			Checksum:    checksum,
			Required:    false,
			Description: "Provider module",
		}
		
		bundle.Components = append(bundle.Components, component)
		bundle.Checksums[component.Path] = checksum
		*totalSize += info.Size()
		
		return nil
	})
	
	return err

// createBundleArchive creates the final bundle archive
func (bm *BundleManager) createBundleArchive(bundle *Bundle, workspaceDir, outputPath string, totalSize int64) (*BundleInfo, error) {
	// Create bundle manifest
	manifestPath := filepath.Join(workspaceDir, "manifest.json")
	manifestData, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	if err := os.WriteFile(filepath.Clean(manifestPath, manifestData, 0600)); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}
	
	// Create ZIP archive
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle file: %w", err)
	}
	defer func() { if err := zipFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	zipWriter := zip.NewWriter(zipFile)
	defer func() { if err := zipWriter.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	// Add all files to ZIP
	err = filepath.Walk(workspaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		relPath, err := filepath.Rel(workspaceDir, path)
		if err != nil {
			return err
		}
		
		// Create file in ZIP
		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		
		// Copy file content
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
		
		_, err = io.Copy(zipFileWriter, file)
		return err
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP archive: %w", err)
	}
		
	// Get final file size
	zipWriter.Close()
	zipFile.Close()
	
	finalInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat final bundle: %w", err)
	}
	
	return &BundleInfo{
		Path:         outputPath,
		Size:         finalInfo.Size(),
		Created:      time.Now(),
		ComponentCount: len(bundle.Components),
		Signed:       false,
		Bundle:       bundle,
	}, nil

// verifyBundleIntegrity verifies bundle integrity
func (bm *BundleManager) verifyBundleIntegrity(bundlePath string) error {
	result, err := bm.verifier.VerifyBundle(bundlePath)
	if err != nil {
		return err
	}
	
	if !result.Verified {
		return fmt.Errorf("bundle verification failed")
	}
	
	return nil

// extractBundle extracts and parses a bundle
func (bm *BundleManager) extractBundle(bundlePath, workspaceDir string) (*Bundle, error) {
	// Open ZIP file
	reader, err := zip.OpenReader(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer func() { if err := reader.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	// Extract files
	for _, file := range reader.File {
		if err := bm.extractBundleFile(file, workspaceDir); err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}
	}
	
	// Read manifest
	manifestPath := filepath.Join(workspaceDir, "manifest.json")
	manifestData, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var bundle Bundle
	if err := json.Unmarshal(manifestData, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	return &bundle, nil

// extractBundleFile extracts a single file from bundle
func (bm *BundleManager) extractBundleFile(file *zip.File, destDir string) error {
	// Clean path
	cleanPath := filepath.Clean(file.Name)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: %s", file.Name)
	}
	
	destPath := filepath.Join(destDir, cleanPath)
	
	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.FileInfo().Mode())
	}
	
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
		return err
	}
	
	// Extract file
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { if err := reader.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { if err := destFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	_, err = io.Copy(destFile, reader)
	return err

// checkVersionCompatibility checks if bundle is compatible
func (bm *BundleManager) checkVersionCompatibility(bundle *Bundle, options *BundleImportOptions) error {
	// Check tool version compatibility
	currentVersion := bm.getCurrentBinaryVersion()
	
	if bundle.Metadata.TargetVersion != "" && bundle.Metadata.TargetVersion != currentVersion {
		return fmt.Errorf("bundle targets version %s, current version is %s", 
			bundle.Metadata.TargetVersion, currentVersion)
	}
	
	// Check platform compatibility
	if len(bundle.Metadata.Platforms) > 0 {
		currentPlatform := GetPlatformString()
		platformSupported := false
		
		for _, platform := range bundle.Metadata.Platforms {
			if platform == currentPlatform {
				platformSupported = true
				break
			}
		}
		
		if !platformSupported {
			return fmt.Errorf("bundle does not support platform %s", currentPlatform)
		}
	}
	
	return nil
	
// Helper functions

func (bm *BundleManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer func() { if err := sourceFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { if err := destFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	_, err = io.Copy(destFile, sourceFile)
	return err

func (bm *BundleManager) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil

func (bm *BundleManager) getCurrentBinaryVersion() string {
	version, _ := GetCurrentVersion()
	return version

func (bm *BundleManager) getTemplateVersion(filePath string) string {
	// Simple version extraction - could be more sophisticated
	return "1.0.0"

func (bm *BundleManager) getModuleVersion(filePath string) string {
	// Simple version extraction - could be more sophisticated
	return "1.0.0"

// Data structures

type BundleExportOptions struct {
	Version          string
	Type             string
	Description      string
	SourceVersion    string
	TargetVersion    string
	Platforms        []string
	Incremental      bool
	OutputPath       string
	CreatedBy        string
	ExpiresAt        *time.Time
	IncludeBinary    bool
	IncludeTemplates bool
	IncludeModules   bool
	SignBundle       bool
	SigningKey       *rsa.PrivateKey
}

type BundleImportOptions struct {
	VerifyIntegrity bool
	ForceImport     bool
	CreateBackup    bool
	DryRun          bool

type BundleInfo struct {
	Path           string
	Size           int64
	Created        time.Time
	ComponentCount int
	Signed         bool
	Bundle         *Bundle

type ImportResult struct {
	Success        bool
	ComponentsUpdated []string
	Errors         []string
	BackupPath     string
	RestartRequired bool
}

// Placeholder implementations for missing methods

func (bm *BundleManager) signBundle(bundlePath string, signingKey *rsa.PrivateKey) error {
	if signingKey == nil {
		return fmt.Errorf("no signing key provided")
	}
	
	// Calculate bundle hash
	file, err := os.Open(filepath.Clean(bundlePath))
	if err != nil {
		return err
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	
	hash := hasher.Sum(nil)
	
	// Sign hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, signingKey, crypto.SHA256, hash)
	if err != nil {
		return err
	}
	
	// Write signature file
	signaturePath := bundlePath + ".sig"
	signatureData := base64.StdEncoding.EncodeToString(signature)
	
	return os.WriteFile(filepath.Clean(signaturePath, []byte(signatureData)), 0600)

func (bm *BundleManager) createImportBackup() error {
	// Implementation for creating backup before import
	return nil

func (bm *BundleManager) applyBundleUpdates(bundle *Bundle, workspaceDir string, options *BundleImportOptions) (*ImportResult, error) {
	result := &ImportResult{
		Success:           true,
		ComponentsUpdated: make([]string, 0),
		Errors:           make([]string, 0),
		RestartRequired:  false,
	}
	
	// Apply each component
	for _, component := range bundle.Components {
		sourcePath := filepath.Join(workspaceDir, component.Path)
		
		switch component.Type {
		case ComponentBinary:
			if err := bm.applyBinaryUpdate(sourcePath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("binary update failed: %v", err))
				result.Success = false
			} else {
				result.ComponentsUpdated = append(result.ComponentsUpdated, component.Name)
				result.RestartRequired = true
			}
		case ComponentTemplates:
			if err := bm.applyTemplateUpdate(sourcePath, component); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("template %s update failed: %v", component.Name, err))
			} else {
				result.ComponentsUpdated = append(result.ComponentsUpdated, component.Name)
			}
		case ComponentModules:
			if err := bm.applyModuleUpdate(sourcePath, component); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("module %s update failed: %v", component.Name, err))
			} else {
				result.ComponentsUpdated = append(result.ComponentsUpdated, component.Name)
			}
		}
	}
	
	return result, nil

func (bm *BundleManager) applyBinaryUpdate(sourcePath string) error {
	// Implementation for applying binary update
	return nil

func (bm *BundleManager) applyTemplateUpdate(sourcePath string, component ComponentInfo) error {
	// Implementation for applying template update
	return nil

func (bm *BundleManager) applyModuleUpdate(sourcePath string, component ComponentInfo) error {
	// Implementation for applying module update
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
}
}
}
}
}
