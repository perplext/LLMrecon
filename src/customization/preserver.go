package customization

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Preserver preserves and reapplies user customizations during updates
type Preserver struct {
	// Registry is the customization registry
	Registry *Registry
	// InstallDir is the directory where the tool is installed
	InstallDir string
	// BackupDir is the directory for backups during update
	BackupDir string
	// Logger is the logger for preserver operations
	Logger *os.File
}

// PreserverOptions contains options for the Preserver
type PreserverOptions struct {
	// Registry is the customization registry
	Registry *Registry
	// InstallDir is the directory where the tool is installed
	InstallDir string
	// BackupDir is the directory for backups during update
	BackupDir string
	// LogFile is the file to log preserver operations
	LogFile *os.File
}

// NewPreserver creates a new customization preserver
func NewPreserver(options *PreserverOptions) (*Preserver, error) {
	if options.Registry == nil {
		return nil, fmt.Errorf("registry is required")
	}
	if options.InstallDir == "" {
		return nil, fmt.Errorf("install directory is required")
	}
	if options.BackupDir == "" {
		return nil, fmt.Errorf("backup directory is required")
	}

	return &Preserver{
		Registry:   options.Registry,
		InstallDir: options.InstallDir,
		BackupDir:  options.BackupDir,
		Logger:     options.LogFile,
	}, nil
}

// PreserveTemplateCustomizations preserves template customizations before update
func (p *Preserver) PreserveTemplateCustomizations() error {
	// Get all template customization entries
	var entries []*CustomizationEntry
	for _, entry := range p.Registry.Entries {
		if entry.Type == TemplateCustomization {
			entries = append(entries, entry)
		}
	}

	// Group entries by template ID
	templateEntries := make(map[string][]*CustomizationEntry)
	for _, entry := range entries {
		templateEntries[entry.ComponentID] = append(templateEntries[entry.ComponentID], entry)
	}

	// Create backup of customized templates
	for templateID, entries := range templateEntries {
		// Create backup directory for this template
		backupDir := filepath.Join(p.BackupDir, "templates", templateID)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory for template %s: %w", templateID, err)
		}

		// Copy customized files to backup
		for _, entry := range entries {
			// Get source and destination paths
			srcPath := filepath.Join(p.InstallDir, "templates", entry.Path)
			dstPath := filepath.Join(p.BackupDir, "templates", entry.Path)

			// Create destination directory
			dstDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return fmt.Errorf("failed to create backup directory for template file: %w", err)
			}

			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to backup template file: %w", err)
			}

			p.logf("Preserved customization for template %s: %s", templateID, entry.Path)
		}
	}

	return nil
}

// PreserveModuleCustomizations preserves module customizations before update
func (p *Preserver) PreserveModuleCustomizations() error {
	// Get all module customization entries
	var entries []*CustomizationEntry
	for _, entry := range p.Registry.Entries {
		if entry.Type == ModuleCustomization {
			entries = append(entries, entry)
		}
	}

	// Group entries by module ID
	moduleEntries := make(map[string][]*CustomizationEntry)
	for _, entry := range entries {
		moduleEntries[entry.ComponentID] = append(moduleEntries[entry.ComponentID], entry)
	}

	// Create backup of customized modules
	for moduleID, entries := range moduleEntries {
		// Create backup directory for this module
		backupDir := filepath.Join(p.BackupDir, "modules", moduleID)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory for module %s: %w", moduleID, err)
		}

		// Copy customized files to backup
		for _, entry := range entries {
			// Get source and destination paths
			srcPath := filepath.Join(p.InstallDir, "modules", entry.Path)
			dstPath := filepath.Join(p.BackupDir, "modules", entry.Path)

			// Create destination directory
			dstDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return fmt.Errorf("failed to create backup directory for module file: %w", err)
			}

			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to backup module file: %w", err)
			}

			p.logf("Preserved customization for module %s: %s", moduleID, entry.Path)
		}
	}

	return nil
}

// ReapplyTemplateCustomizations reapplies template customizations after update
func (p *Preserver) ReapplyTemplateCustomizations(updatedTemplates []string) error {
	// Get all template customization entries
	var entries []*CustomizationEntry
	for _, entry := range p.Registry.Entries {
		if entry.Type == TemplateCustomization {
			// Check if this template was updated
			templateUpdated := false
			for _, updatedTemplate := range updatedTemplates {
				if entry.ComponentID == updatedTemplate {
					templateUpdated = true
					break
				}
			}

			if templateUpdated {
				entries = append(entries, entry)
			}
		}
	}

	// Reapply customizations
	for _, entry := range entries {
		// Get paths
		backupPath := filepath.Join(p.BackupDir, "templates", entry.Path)
		installPath := filepath.Join(p.InstallDir, "templates", entry.Path)

		// Check if backup exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			p.logf("Warning: Backup not found for template customization: %s", entry.Path)
			continue
		}

		// Check if install path exists
		if _, err := os.Stat(installPath); os.IsNotExist(err) {
			p.logf("Warning: Template file not found after update: %s", entry.Path)
			continue
		}

		// Read files
		backupContent, err := ioutil.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup file: %w", err)
		}

		installContent, err := ioutil.ReadFile(installPath)
		if err != nil {
			return fmt.Errorf("failed to read updated file: %w", err)
		}

		// Calculate hashes
		_ = calculateHash(backupContent) // backupHash not used
		installHash := calculateHash(installContent)

		// Check if the file was modified in the update
		if installHash != entry.OriginalHash {
			// Apply customization based on policy
			switch entry.Policy {
			case AlwaysPreserve:
				// Always use the customized version
				if err := copyFile(backupPath, installPath); err != nil {
					return fmt.Errorf("failed to restore customized file: %w", err)
				}
				p.logf("Reapplied customization (AlwaysPreserve) for template %s: %s", entry.ComponentID, entry.Path)

			case PreserveWithConflictResolution:
				// Merge changes
				mergedContent, err := mergeCustomizations(string(installContent), string(backupContent), entry.Markers)
				if err != nil {
					return fmt.Errorf("failed to merge customizations: %w", err)
				}

				// Write merged content
				if err := ioutil.WriteFile(installPath, []byte(mergedContent), 0644); err != nil {
					return fmt.Errorf("failed to write merged file: %w", err)
				}
				p.logf("Reapplied customization (PreserveWithConflictResolution) for template %s: %s", entry.ComponentID, entry.Path)

			case AskUser:
				// For now, preserve the customization
				// In a real implementation, this would prompt the user
				if err := copyFile(backupPath, installPath); err != nil {
					return fmt.Errorf("failed to restore customized file: %w", err)
				}
				p.logf("Reapplied customization (AskUser) for template %s: %s", entry.ComponentID, entry.Path)

			case Discard:
				// Keep the updated version
				p.logf("Discarded customization (Discard) for template %s: %s", entry.ComponentID, entry.Path)
			}
		} else {
			// File wasn't modified in the update, restore customization
			if err := copyFile(backupPath, installPath); err != nil {
				return fmt.Errorf("failed to restore customized file: %w", err)
			}
			p.logf("Reapplied customization (unchanged file) for template %s: %s", entry.ComponentID, entry.Path)
		}

		// Update registry entry with new base version
		templatePath := filepath.Join(p.InstallDir, "templates", entry.ComponentID, "template.yaml")
		newVersion, err := getTemplateVersion(templatePath)
		if err != nil {
			p.logf("Warning: Failed to get new template version: %v", err)
		} else {
			entry.BaseVersion = newVersion
			p.Registry.AddEntry(entry)
		}
	}

	// Save registry
	if err := p.Registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	return nil
}

// ReapplyModuleCustomizations reapplies module customizations after update
func (p *Preserver) ReapplyModuleCustomizations(updatedModules []string) error {
	// Get all module customization entries
	var entries []*CustomizationEntry
	for _, entry := range p.Registry.Entries {
		if entry.Type == ModuleCustomization {
			// Check if this module was updated
			moduleUpdated := false
			for _, updatedModule := range updatedModules {
				if entry.ComponentID == updatedModule {
					moduleUpdated = true
					break
				}
			}

			if moduleUpdated {
				entries = append(entries, entry)
			}
		}
	}

	// Reapply customizations
	for _, entry := range entries {
		// Get paths
		backupPath := filepath.Join(p.BackupDir, "modules", entry.Path)
		installPath := filepath.Join(p.InstallDir, "modules", entry.Path)

		// Check if backup exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			p.logf("Warning: Backup not found for module customization: %s", entry.Path)
			continue
		}

		// Check if install path exists
		if _, err := os.Stat(installPath); os.IsNotExist(err) {
			p.logf("Warning: Module file not found after update: %s", entry.Path)
			continue
		}

		// Read files
		backupContent, err := ioutil.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup file: %w", err)
		}

		installContent, err := ioutil.ReadFile(installPath)
		if err != nil {
			return fmt.Errorf("failed to read updated file: %w", err)
		}

		// Calculate hashes
		_ = calculateHash(backupContent) // backupHash not used
		installHash := calculateHash(installContent)

		// Check if the file was modified in the update
		if installHash != entry.OriginalHash {
			// Apply customization based on policy
			switch entry.Policy {
			case AlwaysPreserve:
				// Always use the customized version
				if err := copyFile(backupPath, installPath); err != nil {
					return fmt.Errorf("failed to restore customized file: %w", err)
				}
				p.logf("Reapplied customization (AlwaysPreserve) for module %s: %s", entry.ComponentID, entry.Path)

			case PreserveWithConflictResolution:
				// Merge changes
				mergedContent, err := mergeCustomizations(string(installContent), string(backupContent), entry.Markers)
				if err != nil {
					return fmt.Errorf("failed to merge customizations: %w", err)
				}

				// Write merged content
				if err := ioutil.WriteFile(installPath, []byte(mergedContent), 0644); err != nil {
					return fmt.Errorf("failed to write merged file: %w", err)
				}
				p.logf("Reapplied customization (PreserveWithConflictResolution) for module %s: %s", entry.ComponentID, entry.Path)

			case AskUser:
				// For now, preserve the customization
				// In a real implementation, this would prompt the user
				if err := copyFile(backupPath, installPath); err != nil {
					return fmt.Errorf("failed to restore customized file: %w", err)
				}
				p.logf("Reapplied customization (AskUser) for module %s: %s", entry.ComponentID, entry.Path)

			case Discard:
				// Keep the updated version
				p.logf("Discarded customization (Discard) for module %s: %s", entry.ComponentID, entry.Path)
			}
		} else {
			// File wasn't modified in the update, restore customization
			if err := copyFile(backupPath, installPath); err != nil {
				return fmt.Errorf("failed to restore customized file: %w", err)
			}
			p.logf("Reapplied customization (unchanged file) for module %s: %s", entry.ComponentID, entry.Path)
		}

		// Update registry entry with new base version
		moduleDir := filepath.Join(p.InstallDir, "modules", entry.ComponentID)
		newVersion, err := getModuleVersion(moduleDir)
		if err != nil {
			p.logf("Warning: Failed to get new module version: %v", err)
		} else {
			entry.BaseVersion = newVersion
			p.Registry.AddEntry(entry)
		}
	}

	// Save registry
	if err := p.Registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	return nil
}

// Helper functions

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read source file
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create destination directory
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write destination file
	if err := ioutil.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// mergeCustomizations merges customizations from customized content into updated content
func mergeCustomizations(updatedContent, customizedContent string, markers []CustomizationMarker) (string, error) {
	// If there are no markers, just return the customized content
	if len(markers) == 0 {
		return customizedContent, nil
	}

	// Split content into lines
	updatedLines := strings.Split(updatedContent, "\n")
	_ = strings.Split(customizedContent, "\n") // customizedLines not used

	// For each marker, replace the content in the updated file with the customized content
	for _, marker := range markers {
		// Find marker in updated content
		startLine := -1
		endLine := -1
		for i, line := range updatedLines {
			if strings.Contains(line, fmt.Sprintf("%s BEGIN", strings.ToUpper(strings.Replace(marker.Type, "_", " ", -1)))) {
				startLine = i
			} else if startLine != -1 && strings.Contains(line, fmt.Sprintf("%s END", strings.ToUpper(strings.Replace(marker.Type, "_", " ", -1)))) {
				endLine = i
				break
			}
		}

		// If marker not found in updated content, skip
		if startLine == -1 || endLine == -1 {
			continue
		}

		// Replace content between markers
		newLines := make([]string, 0)
		newLines = append(newLines, updatedLines[:startLine+1]...)
		newLines = append(newLines, strings.Split(marker.Content, "\n")...)
		newLines = append(newLines, updatedLines[endLine:]...)
		updatedLines = newLines
	}

	return strings.Join(updatedLines, "\n"), nil
}

// logf logs a message to the logger
func (p *Preserver) logf(format string, args ...interface{}) {
	if p.Logger != nil {
		fmt.Fprintf(p.Logger, "[%s] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	}
}
