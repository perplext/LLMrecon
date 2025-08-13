package version

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// DiffType represents the type of difference between two files
type DiffType string

const (
	// Added means the file was added
	Added DiffType = "added"
	// Modified means the file was modified
	Modified DiffType = "modified"
	// Deleted means the file was deleted
	Deleted DiffType = "deleted"
	// Unchanged means the file was not changed
	Unchanged DiffType = "unchanged"
)

// FileDiff represents a difference between two files
type FileDiff struct {
	// Path is the path to the file
	Path string
	
	// Type is the type of difference
	Type DiffType
	
	// OldHash is the hash of the old file
	OldHash string
	
	// NewHash is the hash of the new file
	NewHash string
	
	// OldSize is the size of the old file in bytes
	OldSize int64
	
	// NewSize is the size of the new file in bytes
	NewSize int64
	
	// OldModTime is the modification time of the old file
	OldModTime time.Time
	
	// NewModTime is the modification time of the new file
	NewModTime time.Time
}

// ContentDiff represents a difference between two file contents
type ContentDiff struct {
	// Path is the path to the file
	Path string
	
	// Type is the type of difference
	Type DiffType
	
	// Chunks is the list of changed chunks
	Chunks []*DiffChunk
}

// DiffChunk represents a chunk of a content difference
type DiffChunk struct {
	// OldStart is the starting line number in the old file (1-based)
	OldStart int
	
	// OldLines is the number of lines in the old file
	OldLines int
	
	// NewStart is the starting line number in the new file (1-based)
	NewStart int
	
	// NewLines is the number of lines in the new file
	NewLines int
	
	// Content is the content of the chunk
	Content []string
}

// DiffResult represents the result of a differential analysis
type DiffResult struct {
	// FileDiffs is the list of file differences
	FileDiffs []*FileDiff
	
	// ContentDiffs is the list of content differences
	ContentDiffs []*ContentDiff
	
	// OldVersion is the old version
	OldVersion *Version
	
	// NewVersion is the new version
	NewVersion *Version
	
	// DiffTime is the time the diff was performed
	DiffTime time.Time
}

// DiffOptions represents options for differential analysis
type DiffOptions struct {
	// IgnoreWhitespace ignores whitespace changes
	IgnoreWhitespace bool
	
	// IgnoreCase ignores case changes
	IgnoreCase bool
	
	// IgnoreLineEndings ignores line ending changes
	IgnoreLineEndings bool
	
	// MaxContentSize is the maximum content size to diff in bytes
	MaxContentSize int64
	
	// IncludeContent includes content diffs
	IncludeContent bool
	
	// IncludeUnchanged includes unchanged files
	IncludeUnchanged bool
}

// DefaultDiffOptions returns the default diff options
func DefaultDiffOptions() *DiffOptions {
	return &DiffOptions{
		IgnoreWhitespace: false,
		IgnoreCase:       false,
		IgnoreLineEndings: false,
		MaxContentSize:   1024 * 1024, // 1MB
		IncludeContent:   true,
		IncludeUnchanged: false,
	}
}

// FileInfo represents information about a file
type FileInfo struct {
	// Path is the path to the file
	Path string
	
	// Hash is the hash of the file
	Hash string
	
	// Size is the size of the file in bytes
	Size int64
	
	// ModTime is the modification time of the file
	ModTime time.Time
	
	// Content is the content of the file
	Content []byte
}

// ComputeHash computes the SHA-256 hash of a file
func ComputeHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// DiffFiles compares two sets of files and returns the differences
func DiffFiles(oldFiles, newFiles []*FileInfo, options *DiffOptions) *DiffResult {
	if options == nil {
		options = DefaultDiffOptions()
	}
	
	result := &DiffResult{
		FileDiffs:    make([]*FileDiff, 0),
		ContentDiffs: make([]*ContentDiff, 0),
		DiffTime:     time.Now(),
	}
	
	// Create maps for quick lookup
	oldFileMap := make(map[string]*FileInfo)
	for _, file := range oldFiles {
		oldFileMap[file.Path] = file
	}
	
	newFileMap := make(map[string]*FileInfo)
	for _, file := range newFiles {
		newFileMap[file.Path] = file
	}
	
	// Find added and modified files
	for _, newFile := range newFiles {
		oldFile, exists := oldFileMap[newFile.Path]
		
		if !exists {
			// File was added
			result.FileDiffs = append(result.FileDiffs, &FileDiff{
				Path:       newFile.Path,
				Type:       Added,
				NewHash:    newFile.Hash,
				NewSize:    newFile.Size,
				NewModTime: newFile.ModTime,
			})
			
			if options.IncludeContent {
				result.ContentDiffs = append(result.ContentDiffs, &ContentDiff{
					Path: newFile.Path,
					Type: Added,
					Chunks: []*DiffChunk{
						{
							NewStart: 1,
							NewLines: countLines(newFile.Content),
							Content:  splitLines(newFile.Content),
						},
					},
				})
			}
		} else {
			// File exists in both old and new
			if oldFile.Hash != newFile.Hash {
				// File was modified
				result.FileDiffs = append(result.FileDiffs, &FileDiff{
					Path:       newFile.Path,
					Type:       Modified,
					OldHash:    oldFile.Hash,
					NewHash:    newFile.Hash,
					OldSize:    oldFile.Size,
					NewSize:    newFile.Size,
					OldModTime: oldFile.ModTime,
					NewModTime: newFile.ModTime,
				})
				
				if options.IncludeContent && newFile.Size <= options.MaxContentSize && oldFile.Size <= options.MaxContentSize {
					contentDiff := diffContent(oldFile.Content, newFile.Content, options)
					if contentDiff != nil {
						contentDiff.Path = newFile.Path
						contentDiff.Type = Modified
						result.ContentDiffs = append(result.ContentDiffs, contentDiff)
					}
				}
			} else if options.IncludeUnchanged {
				// File was not changed
				result.FileDiffs = append(result.FileDiffs, &FileDiff{
					Path:       newFile.Path,
					Type:       Unchanged,
					OldHash:    oldFile.Hash,
					NewHash:    newFile.Hash,
					OldSize:    oldFile.Size,
					NewSize:    newFile.Size,
					OldModTime: oldFile.ModTime,
					NewModTime: newFile.ModTime,
				})
			}
			
			// Mark as processed
			delete(oldFileMap, newFile.Path)
		}
	}
	
	// Find deleted files
	for _, oldFile := range oldFileMap {
		result.FileDiffs = append(result.FileDiffs, &FileDiff{
			Path:       oldFile.Path,
			Type:       Deleted,
			OldHash:    oldFile.Hash,
			OldSize:    oldFile.Size,
			OldModTime: oldFile.ModTime,
		})
		
		if options.IncludeContent {
			result.ContentDiffs = append(result.ContentDiffs, &ContentDiff{
				Path: oldFile.Path,
				Type: Deleted,
				Chunks: []*DiffChunk{
					{
						OldStart: 1,
						OldLines: countLines(oldFile.Content),
						Content:  splitLines(oldFile.Content),
					},
				},
			})
		}
	}
	
	return result
}

// diffContent compares the content of two files and returns the differences
func diffContent(oldContent, newContent []byte, options *DiffOptions) *ContentDiff {
	// This is a simplified implementation that just compares the content
	// A real implementation would use a diff algorithm like Myers diff
	
	if bytes.Equal(oldContent, newContent) {
		return nil
	}
	
	// If the content is different, create a simple diff
	return &ContentDiff{
		Type: Modified,
		Chunks: []*DiffChunk{
			{
				OldStart: 1,
				OldLines: countLines(oldContent),
				NewStart: 1,
				NewLines: countLines(newContent),
				Content:  []string{"[Content differs]"},
			},
		},
	}
}

// countLines counts the number of lines in a byte slice
func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	
	count := 1
	for _, b := range content {
		if b == '\n' {
			count++
		}
	}
	
	return count
}

// splitLines splits a byte slice into lines
func splitLines(content []byte) []string {
	if len(content) == 0 {
		return []string{}
	}
	
	lines := bytes.Split(content, []byte{'\n'})
	result := make([]string, len(lines))
	
	for i, line := range lines {
		result[i] = string(line)
	}
	
	return result
}

// GetChangeSummary returns a summary of the changes
func (r *DiffResult) GetChangeSummary() string {
	added := 0
	modified := 0
	deleted := 0
	unchanged := 0
	
	for _, diff := range r.FileDiffs {
		switch diff.Type {
		case Added:
			added++
		case Modified:
			modified++
		case Deleted:
			deleted++
		case Unchanged:
			unchanged++
		}
	}
	
	return fmt.Sprintf("Added: %d, Modified: %d, Deleted: %d, Unchanged: %d", added, modified, deleted, unchanged)
}

// GetChangedFiles returns a list of changed files
func (r *DiffResult) GetChangedFiles() []*FileDiff {
	changed := make([]*FileDiff, 0)
	
	for _, diff := range r.FileDiffs {
		if diff.Type != Unchanged {
			changed = append(changed, diff)
		}
	}
	
	return changed
}

// GetChangedPaths returns a list of changed file paths
func (r *DiffResult) GetChangedPaths() []string {
	changed := make([]string, 0)
	
	for _, diff := range r.FileDiffs {
		if diff.Type != Unchanged {
			changed = append(changed, diff.Path)
		}
	}
	
	return changed
}

// HasChanges returns true if there are any changes
func (r *DiffResult) HasChanges() bool {
	for _, diff := range r.FileDiffs {
		if diff.Type != Unchanged {
			return true
		}
	}
	
	return false
}

// ReadFileContent reads the content of a file from a reader
func ReadFileContent(r io.Reader, maxSize int64) ([]byte, error) {
	// If maxSize is 0, read the entire file
	if maxSize <= 0 {
		return io.ReadAll(r)
	}
	
	// Read up to maxSize bytes
	return io.ReadAll(io.LimitReader(r, maxSize))
}
