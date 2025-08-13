// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"fmt"
)

// clearDirectory removes all files and directories in a directory
// but keeps the directory itself
func clearDirectory(dir string) error {
	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Remove each entry
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		var err error
		if entry.IsDir() {
			// Remove directory and all contents
			err = os.RemoveAll(path)
		} else {
			// Remove file
			err = os.Remove(path)
		}
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// copyDirUtil copies a directory and its contents to another directory (utility version)
func copyDirUtil(src, dst string) error {
	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	// Create destination directory with same permissions
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy directory
			err = copyDirUtil(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = copyFileUtil(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFileUtil copies a file from src to dst (utility version)
func copyFileUtil(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
