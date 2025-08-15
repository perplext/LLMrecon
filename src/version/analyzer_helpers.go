package version

import (
	"crypto/sha256"
	"fmt"
)

// ReadFileContent reads content from a reader up to maxSize bytes
func ReadFileContent(reader io.Reader, maxSize int64) ([]byte, error) {
	if maxSize <= 0 {
		maxSize = 1024 * 1024 // Default 1MB
	}
	return io.ReadAll(io.LimitReader(reader, maxSize))

// FileInfo represents information about a file
type FileInfo struct {
	Path    string
	Hash    string
	Size    int64
	ModTime time.Time
	Content []byte
}

// ComputeHash computes a hash for the given content
}
func ComputeHash(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	return fmt.Sprintf("%x", hasher.Sum(nil))

// Diff represents the differences between two sets of files
type Diff struct {
	OldFiles   []*FileInfo
	NewFiles   []*FileInfo
	OldVersion *SemVersion
	NewVersion *SemVersion

// DiffFiles creates a diff between two sets of files
}
func DiffFiles(oldFiles, newFiles []*FileInfo, options *DiffOptions) *DiffResult {
	// Stub implementation - return empty diff result
	return &DiffResult{
		LocalVersion:  nil, // Will be set by caller
		RemoteVersion: nil, // Will be set by caller
		Items:         []*DiffItem{},
		Summary:       &DiffSummary{},
		DiffTime:      time.Now(),
	}
