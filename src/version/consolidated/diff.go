package version

import (
	"fmt"
	"time"
)

// DiffType represents the type of difference
type DiffType string

const (
	// AddedDiff represents an added item
	AddedDiff DiffType = "added"
	
	// RemovedDiff represents a removed item
	RemovedDiff DiffType = "removed"
	
	// ModifiedDiff represents a modified item
	ModifiedDiff DiffType = "modified"
	
	// UnchangedDiff represents an unchanged item
	UnchangedDiff DiffType = "unchanged"
)

// DiffOptions represents options for diffing
type DiffOptions struct {
	// IgnoreWhitespace determines if whitespace should be ignored
	IgnoreWhitespace bool
	
	// IgnoreCase determines if case should be ignored
	IgnoreCase bool
	
	// IgnoreComments determines if comments should be ignored
	IgnoreComments bool
	
	// IgnoreFormatting determines if formatting should be ignored
	IgnoreFormatting bool
	
	// IncludeUnchanged determines if unchanged items should be included
	IncludeUnchanged bool
}

// DefaultDiffOptions returns the default diff options
func DefaultDiffOptions() *DiffOptions {
	return &DiffOptions{
		IgnoreWhitespace: true,
		IgnoreCase:       false,
		IgnoreComments:   true,
		IgnoreFormatting: true,
		IncludeUnchanged: false,
	}
}

// DiffItem represents a difference item
type DiffItem struct {
	// Type is the type of difference
	Type DiffType
	
	// Path is the path to the item
	Path string
	
	// OldContent is the old content
	OldContent string
	
	// NewContent is the new content
	NewContent string
	
	// LineNumber is the line number of the difference
	LineNumber int
	
	// LineCount is the number of lines affected
	LineCount int
	
	// Metadata is additional metadata for the difference
	Metadata map[string]interface{}
}

// NewDiffItem creates a new diff item
func NewDiffItem(diffType DiffType, path string, oldContent, newContent string, lineNumber, lineCount int) *DiffItem {
	return &DiffItem{
		Type:       diffType,
		Path:       path,
		OldContent: oldContent,
		NewContent: newContent,
		LineNumber: lineNumber,
		LineCount:  lineCount,
		Metadata:   make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the diff item
func (d *DiffItem) WithMetadata(key string, value interface{}) *DiffItem {
	d.Metadata[key] = value
	return d
}

// String returns a string representation of the diff item
func (d *DiffItem) String() string {
	switch d.Type {
	case AddedDiff:
		return fmt.Sprintf("+ %s (line %d, %d lines)", d.Path, d.LineNumber, d.LineCount)
	case RemovedDiff:
		return fmt.Sprintf("- %s (line %d, %d lines)", d.Path, d.LineNumber, d.LineCount)
	case ModifiedDiff:
		return fmt.Sprintf("M %s (line %d, %d lines)", d.Path, d.LineNumber, d.LineCount)
	case UnchangedDiff:
		return fmt.Sprintf("  %s (line %d, %d lines)", d.Path, d.LineNumber, d.LineCount)
	default:
		return fmt.Sprintf("? %s (line %d, %d lines)", d.Path, d.LineNumber, d.LineCount)
	}
}

// DiffResult represents the result of a diff operation
type DiffResult struct {
	// LocalVersion is the local version
	LocalVersion *VersionInfo
	
	// RemoteVersion is the remote version
	RemoteVersion *VersionInfo
	
	// Items is the list of diff items
	Items []*DiffItem
	
	// Summary is a summary of the differences
	Summary *DiffSummary
	
	// DiffTime is the time the diff was performed
	DiffTime time.Time
}

// NewDiffResult creates a new diff result
func NewDiffResult(localVersion, remoteVersion *VersionInfo) *DiffResult {
	return &DiffResult{
		LocalVersion:  localVersion,
		RemoteVersion: remoteVersion,
		Items:         []*DiffItem{},
		Summary:       NewDiffSummary(),
		DiffTime:      time.Now(),
	}
}

// AddItem adds a diff item to the result
func (r *DiffResult) AddItem(item *DiffItem) {
	r.Items = append(r.Items, item)
	
	// Update summary
	switch item.Type {
	case AddedDiff:
		r.Summary.Added++
	case RemovedDiff:
		r.Summary.Removed++
	case ModifiedDiff:
		r.Summary.Modified++
	case UnchangedDiff:
		r.Summary.Unchanged++
	}
	
	r.Summary.Total++
}

// HasDifferences returns true if there are differences
func (r *DiffResult) HasDifferences() bool {
	return r.Summary.Added > 0 || r.Summary.Removed > 0 || r.Summary.Modified > 0
}

// GetAddedItems returns all added items
func (r *DiffResult) GetAddedItems() []*DiffItem {
	result := make([]*DiffItem, 0)
	for _, item := range r.Items {
		if item.Type == AddedDiff {
			result = append(result, item)
		}
	}
	return result
}

// GetRemovedItems returns all removed items
func (r *DiffResult) GetRemovedItems() []*DiffItem {
	result := make([]*DiffItem, 0)
	for _, item := range r.Items {
		if item.Type == RemovedDiff {
			result = append(result, item)
		}
	}
	return result
}

// GetModifiedItems returns all modified items
func (r *DiffResult) GetModifiedItems() []*DiffItem {
	result := make([]*DiffItem, 0)
	for _, item := range r.Items {
		if item.Type == ModifiedDiff {
			result = append(result, item)
		}
	}
	return result
}

// GetUnchangedItems returns all unchanged items
func (r *DiffResult) GetUnchangedItems() []*DiffItem {
	result := make([]*DiffItem, 0)
	for _, item := range r.Items {
		if item.Type == UnchangedDiff {
			result = append(result, item)
		}
	}
	return result
}

// DiffSummary represents a summary of differences
type DiffSummary struct {
	// Added is the number of added items
	Added int
	
	// Removed is the number of removed items
	Removed int
	
	// Modified is the number of modified items
	Modified int
	
	// Unchanged is the number of unchanged items
	Unchanged int
	
	// Total is the total number of items
	Total int
}

// NewDiffSummary creates a new diff summary
func NewDiffSummary() *DiffSummary {
	return &DiffSummary{
		Added:     0,
		Removed:   0,
		Modified:  0,
		Unchanged: 0,
		Total:     0,
	}
}

// String returns a string representation of the diff summary
func (s *DiffSummary) String() string {
	return fmt.Sprintf("Added: %d, Removed: %d, Modified: %d, Unchanged: %d, Total: %d",
		s.Added, s.Removed, s.Modified, s.Unchanged, s.Total)
}
