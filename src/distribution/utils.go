package distribution

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
)

// calculateFileChecksum calculates the checksum of a file using the specified algorithm
func calculateFileChecksum(filePath, algorithm string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	var hasher hash.Hash
	
	switch algorithm {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
	
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// validateChecksums validates that the artifact's checksums match the calculated values
func validateChecksums(filePath string, expectedChecksums map[string]string) error {
	for algorithm, expectedSum := range expectedChecksums {
		actualSum, err := calculateFileChecksum(filePath, algorithm)
		if err != nil {
			return fmt.Errorf("failed to calculate %s checksum: %w", algorithm, err)
		}
		
		if actualSum != expectedSum {
			return fmt.Errorf("%s checksum mismatch: expected %s, got %s", algorithm, expectedSum, actualSum)
		}
	}
	
	return nil
}

// formatBytes formats byte count as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// sanitizeFileName sanitizes a filename to be safe for filesystem usage
func sanitizeFileName(filename string) string {
	// Replace unsafe characters with underscores
	unsafe := []byte{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
	result := []byte(filename)
	
	for i, b := range result {
		for _, unsafe := range unsafe {
			if b == unsafe {
				result[i] = '_'
				break
			}
		}
	}
	
	return string(result)
}

// ensureDirectory creates a directory if it doesn't exist
func ensureDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
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
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// getFileSize returns the size of a file
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}