// Package customization provides customization preservation
package customization

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CustomizationPreserver preserves user customizations
type CustomizationPreserver struct {
	BackupPath string
}

// NewCustomizationPreserver creates a new customization preserver
func NewCustomizationPreserver(backupPath string) *CustomizationPreserver {
	return &CustomizationPreserver{
		BackupPath: backupPath,
	}
}

// PreserveCustomization preserves a customization
func (p *CustomizationPreserver) PreserveCustomization(custom Customization) error {
	// Create backup directory
	if err := os.MkdirAll(p.BackupPath, 0750); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Determine backup file path
	relPath, err := filepath.Rel(".", custom.Path)
	if err != nil {
		relPath = filepath.Base(custom.Path)
	}
	
	backupFile := filepath.Join(p.BackupPath, relPath)
	backupDir := filepath.Dir(backupFile)
	
	// Create backup subdirectory if needed
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return fmt.Errorf("failed to create backup subdirectory: %w", err)
	}
	
	// Copy the file
	return p.copyFile(custom.Path, backupFile)
}

// copyFile copies a file from source to destination
func (p *CustomizationPreserver) copyFile(src, dst string) error {
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	
	return os.Chmod(dst, info.Mode())
}

// RestoreCustomization restores a customization from backup
func (p *CustomizationPreserver) RestoreCustomization(custom Customization) error {
	// Determine backup file path
	relPath, err := filepath.Rel(".", custom.Path)
	if err != nil {
		relPath = filepath.Base(custom.Path)
	}
	
	backupFile := filepath.Join(p.BackupPath, relPath)
	
	// Check if backup exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}
	
	// Restore the file
	return p.copyFile(backupFile, custom.Path)
}
