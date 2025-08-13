package update

import (
	"fmt"
)

// UpdateCheck represents the result of checking for updates
type UpdateCheck struct {
	CheckTime        time.Time
	UpdatesAvailable bool
	Components       map[string]*ComponentUpdate
}

// ComponentUpdate represents an available update for a component
type ComponentUpdate struct {
	Component      string
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ChangelogURL   string
	ReleaseNotes   string
	UpdateSize     int64
	Critical       bool
	SecurityUpdate bool
}

// Release represents a software release
type Release struct {
	Version      string
	Name         string
	Description  string
	ChangelogURL string
	ReleaseDate  time.Time
	Assets       []ReleaseAsset
	Prerelease   bool
	Draft        bool
	TagName      string
}

// ReleaseAsset represents a downloadable asset from a release
type ReleaseAsset struct {
	Name         string
	DownloadURL  string
	Size         int64
	ContentType  string
	Checksum     string
	SignatureURL string
	Platform     string
	Architecture string
}

// UpdateManifest represents metadata about updates
type UpdateManifest struct {
	Version     string            `json:"version"`
	Components  []ComponentInfo   `json:"components"`
	Checksums   map[string]string `json:"checksums"`
	Signatures  map[string]string `json:"signatures"`
	Timestamp   time.Time         `json:"timestamp"`
	MinVersion  string            `json:"min_version"`
	MaxVersion  string            `json:"max_version"`
	ChangelogURL string           `json:"changelog_url"`
}

// ComponentInfo represents information about a component
type ComponentInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Bundle represents an offline update bundle
type Bundle struct {
	Metadata     BundleMetadata    `json:"metadata"`
	Components   []ComponentInfo   `json:"components"`
	Checksums    map[string]string `json:"checksums"`
	Signatures   map[string]string `json:"signatures"`
	CreatedBy    string            `json:"created_by"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
}

// BundleMetadata contains metadata about a bundle
type BundleMetadata struct {
	Version       string   `json:"version"`
	BundleType    string   `json:"bundle_type"`
	Description   string   `json:"description"`
	SourceVersion string   `json:"source_version"`
	TargetVersion string   `json:"target_version"`
	Platforms     []string `json:"platforms"`
	Incremental   bool     `json:"incremental"`
}

// TemplateInfo represents information about a template
type TemplateInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Category    string            `json:"category"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Severity    string            `json:"severity"`
	Metadata    map[string]string `json:"metadata"`
	FilePath    string            `json:"file_path"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	LastUpdated time.Time         `json:"last_updated"`
}

// ModuleInfo represents information about a provider module
type ModuleInfo struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Provider        string            `json:"provider"`
	Author          string            `json:"author"`
	Description     string            `json:"description"`
	Capabilities    []string          `json:"capabilities"`
	Dependencies    []string          `json:"dependencies"`
	MinToolVersion  string            `json:"min_tool_version"`
	MaxToolVersion  string            `json:"max_tool_version"`
	Metadata        map[string]string `json:"metadata"`
	FilePath        string            `json:"file_path"`
	Size            int64             `json:"size"`
	Checksum        string            `json:"checksum"`
	LastUpdated     time.Time         `json:"last_updated"`
}

// UpdateProgress represents progress of an update operation
type UpdateProgress struct {
	Component     string
	Operation     string
	Progress      float64
	Total         int64
	Current       int64
	Message       string
	StartTime     time.Time
	EstimatedTime time.Duration
}

// UpdateError represents an error during update
type UpdateError struct {
	Component string
	Operation string
	Message   string
	Err       error
	Fatal     bool
	Retry     bool
}

// Error implements the error interface
func (e *UpdateError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s: %s: %v", e.Component, e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("%s %s: %s", e.Component, e.Operation, e.Message)
}

// Unwrap returns the underlying error
func (e *UpdateError) Unwrap() error {
	return e.Err
}

// UpdateOptions represents options for update operations
type UpdateOptions struct {
	Components       []string
	ForceUpdate      bool
	SkipVerification bool
	SkipBackup       bool
	DryRun           bool
	Verbose          bool
	AutoConfirm      bool
	IncludePrerelease bool
	MaxRetries       int
	Timeout          time.Duration
	ProgressCallback func(*UpdateProgress)
	ErrorCallback    func(*UpdateError) bool // Return true to continue
}

// ChangelogEntry represents an entry in a changelog
type ChangelogEntry struct {
	Version     string    `json:"version"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // added, changed, fixed, removed, security
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Breaking    bool      `json:"breaking"`
	Security    bool      `json:"security"`
	References  []string  `json:"references"`
}

// Changelog represents a collection of changelog entries
type Changelog struct {
	Project     string           `json:"project"`
	Format      string           `json:"format"`
	LastUpdated time.Time        `json:"last_updated"`
	Entries     []ChangelogEntry `json:"entries"`
}

// RepositoryStatus represents the status of a repository
type RepositoryStatus struct {
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Type         string    `json:"type"`
	Available    bool      `json:"available"`
	LastChecked  time.Time `json:"last_checked"`
	LastUpdated  time.Time `json:"last_updated"`
	ItemCount    int       `json:"item_count"`
	TotalSize    int64     `json:"total_size"`
	Error        string    `json:"error,omitempty"`
	Version      string    `json:"version"`
}

// UpdateStatistics represents statistics about updates
type UpdateStatistics struct {
	TotalUpdates        int           `json:"total_updates"`
	SuccessfulUpdates   int           `json:"successful_updates"`
	FailedUpdates       int           `json:"failed_updates"`
	LastUpdateTime      time.Time     `json:"last_update_time"`
	AverageUpdateTime   time.Duration `json:"average_update_time"`
	TotalDownloadSize   int64         `json:"total_download_size"`
	UpdatesByComponent  map[string]int `json:"updates_by_component"`
	ErrorsByType        map[string]int `json:"errors_by_type"`
}

// Constants for update operations
const (
	// Component types
	ComponentBinary    = "binary"
	ComponentTemplates = "templates"
	ComponentModules   = "modules"
	
	// Update types
	UpdateTypeFull        = "full"
	UpdateTypeIncremental = "incremental"
	UpdateTypeSecurity    = "security"
	
	// Bundle types
	BundleTypeFull      = "full"
	BundleTypeTemplates = "templates"
	BundleTypeModules   = "modules"
	BundleTypeMixed     = "mixed"
	
	// Operations
	OperationCheck     = "check"
	OperationDownload  = "download"
	OperationVerify    = "verify"
	OperationInstall   = "install"
	OperationBackup    = "backup"
	OperationRollback  = "rollback"
	
	// Error types
	ErrorTypeNetwork      = "network"
	ErrorTypeVerification = "verification"
	ErrorTypePermission   = "permission"
	ErrorTypeCompatibility = "compatibility"
	ErrorTypeCorruption   = "corruption"
	ErrorTypeTimeout      = "timeout"
)

// Update status constants
const (
	StatusPending    = "pending"
	StatusRunning    = "running"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
	StatusSkipped    = "skipped"
)

// Priority levels for updates
const (
	PriorityLow      = "low"
	PriorityNormal   = "normal"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
	PrioritySecurity = "security"
)

// Platform and architecture constants
const (
	PlatformLinux   = "linux"
	PlatformDarwin  = "darwin"
	PlatformWindows = "windows"
	PlatformFreeBSD = "freebsd"
	
	ArchAMD64 = "amd64"
	ArchARM64 = "arm64"
	Arch386   = "386"
	ArchARM   = "arm"
)

// Validation functions

// IsValidVersion checks if a version string is valid
func IsValidVersion(version string) bool {
	// Simple validation - in real implementation would use semver
	return version != "" && len(version) > 0
}

// IsValidComponent checks if a component name is valid
func IsValidComponent(component string) bool {
	validComponents := []string{ComponentBinary, ComponentTemplates, ComponentModules}
	for _, valid := range validComponents {
		if component == valid {
			return true
		}
	}
	return false
}

// GetPlatformString returns the current platform string
func GetPlatformString() string {
	// Would be implemented based on runtime.GOOS
	return "linux" // placeholder
}

// GetArchString returns the current architecture string
func GetArchString() string {
	// Would be implemented based on runtime.GOARCH
	return "amd64" // placeholder
}

// FormatFileSize formats a file size in bytes to human readable format
func FormatFileSize(bytes int64) string {
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

// FormatDuration formats a duration to human readable format
func FormatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.0fs", duration.Seconds())
	}
	if duration < time.Hour {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	return fmt.Sprintf("%.1fh", duration.Hours())
}