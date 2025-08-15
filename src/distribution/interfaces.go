package distribution

import (
	"context"
)

// Logger interface for logging distribution operations
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})

// BuildPipeline interface for building cross-platform binaries
type BuildPipeline interface {
	Build(ctx context.Context, version string, targets []BuildTarget) (*BuildResult, error)
	GetBuildStatus(buildID string) (*BuildStatus, error)
	ListBuilds(ctx context.Context, filters BuildFilters) ([]BuildInfo, error)
	CleanupBuilds(ctx context.Context, olderThan time.Time) error

// PackageManager interface for package manager integration
type PackageManager interface {
	// Package operations
	CreatePackage(ctx context.Context, artifact *BuildArtifact, metadata PackageMetadata) (*Package, error)
	UpdatePackage(ctx context.Context, packageName string, artifact *BuildArtifact) error
	DeletePackage(ctx context.Context, packageName, version string) error
	
	// Repository operations
	PublishToRepository(ctx context.Context, pkg *Package) error
	GetPackageInfo(ctx context.Context, packageName string) (*PackageInfo, error)
	ListPackages(ctx context.Context, filters PackageFilters) ([]PackageInfo, error)
	
	// Metadata operations
	GetSupportedPlatforms() []Platform
	GetType() PackageManagerType
	IsAvailable() bool
	Validate() error

// DistributionChannel interface for distribution channels
type DistributionChannel interface {
	// Release operations
	CreateRelease(ctx context.Context, release *Release) error
	UpdateRelease(ctx context.Context, releaseID string, updates map[string]interface{}) error
	DeleteRelease(ctx context.Context, releaseID string) error
	
	// Asset operations
	UploadAsset(ctx context.Context, releaseID string, asset *Asset) error
	DownloadAsset(ctx context.Context, releaseID, assetName string, writer io.Writer) error
	DeleteAsset(ctx context.Context, releaseID, assetName string) error
	
	// Query operations
	GetRelease(ctx context.Context, releaseID string) (*Release, error)
	ListReleases(ctx context.Context, filters ReleaseFilters) ([]Release, error)
	GetLatestRelease(ctx context.Context) (*Release, error)
	
	// Channel metadata
	GetType() ChannelType
	GetSupportedPlatforms() []Platform
	IsAvailable() bool
	Validate() error

// UpdateVerifier interface for verifying updates
type UpdateVerifier interface {
	// Verification operations
	VerifyChecksum(ctx context.Context, artifact *BuildArtifact) error
	VerifySignature(ctx context.Context, artifact *BuildArtifact) error
	VerifyChain(ctx context.Context, artifact *BuildArtifact) error
	
	// Key management
	AddTrustedKey(ctx context.Context, key *PublicKey) error
	RemoveTrustedKey(ctx context.Context, keyID string) error
	ListTrustedKeys(ctx context.Context) ([]PublicKey, error)
	
	// Configuration
	GetVerificationConfig() VerificationConfig
	UpdateConfig(ctx context.Context, config VerificationConfig) error

// InstallationAnalytics interface for tracking installations
type InstallationAnalytics interface {
	// Event tracking
	TrackInstallation(ctx context.Context, event *InstallationEvent) error
	TrackUpdate(ctx context.Context, event *UpdateEvent) error
	TrackUsage(ctx context.Context, event *UsageEvent) error
	TrackError(ctx context.Context, event *ErrorEvent) error
	
	// Query operations
	GetInstallationStats(ctx context.Context, filters AnalyticsFilters) (*InstallationStats, error)
	GetUsageStats(ctx context.Context, filters AnalyticsFilters) (*UsageStats, error)
	GetErrorStats(ctx context.Context, filters AnalyticsFilters) (*ErrorStats, error)
	
	// Reports
	GenerateReport(ctx context.Context, reportType ReportType, period TimePeriod) (*AnalyticsReport, error)
	ExportData(ctx context.Context, format ExportFormat, filters AnalyticsFilters, writer io.Writer) error
	
	// Configuration
	IsEnabled() bool
	GetRetentionPeriod() time.Duration

// ReleaseManager interface for managing releases
type ReleaseManager interface {
	// Release lifecycle
	CreateRelease(ctx context.Context, release *ReleaseDefinition) (*ReleaseExecution, error)
	PromoteRelease(ctx context.Context, releaseID string, targetChannel ReleaseChannel) error
	RollbackRelease(ctx context.Context, releaseID string) error
	CancelRelease(ctx context.Context, releaseID string) error
	
	// Health monitoring
	CheckReleaseHealth(ctx context.Context, releaseID string) (*HealthStatus, error)
	GetHealthChecks(ctx context.Context, releaseID string) ([]HealthCheckResult, error)
	
	// Strategy management
	GetStrategy() ReleaseStrategy
	UpdateStrategy(ctx context.Context, strategy ReleaseStrategy) error
	
	// Reporting
	GetReleaseStatus(ctx context.Context, releaseID string) (*ReleaseStatus, error)
	ListReleases(ctx context.Context, filters ReleaseFilters) ([]ReleaseInfo, error)

// ArtifactStorage interface for storing build artifacts
type ArtifactStorage interface {
	// Storage operations
	Store(ctx context.Context, artifact *BuildArtifact) (*StorageLocation, error)
	Retrieve(ctx context.Context, location *StorageLocation, writer io.Writer) error
	Delete(ctx context.Context, location *StorageLocation) error
	Exists(ctx context.Context, location *StorageLocation) (bool, error)
	
	// Metadata operations
	GetMetadata(ctx context.Context, location *StorageLocation) (*ArtifactMetadata, error)
	UpdateMetadata(ctx context.Context, location *StorageLocation, metadata *ArtifactMetadata) error
	
	// Query operations
	List(ctx context.Context, filters StorageFilters) ([]StorageLocation, error)
	Search(ctx context.Context, query string) ([]StorageLocation, error)
	
	// Cleanup operations
	Cleanup(ctx context.Context, retentionPolicy RetentionPolicy) error
	GetStorageUsage(ctx context.Context) (*StorageUsage, error)

// Data types for interfaces

// Build-related types
type BuildResult struct {
	BuildID     string          `json:"build_id"`
	Version     string          `json:"version"`
	Artifacts   []BuildArtifact `json:"artifacts"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	Duration    time.Duration   `json:"duration"`
	Status      BuildStatus     `json:"status"`
	LogsURL     string          `json:"logs_url"`
	Errors      []BuildError    `json:"errors,omitempty"`
}

type BuildArtifact struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Platform     Platform          `json:"platform"`
	Architecture Architecture      `json:"architecture"`
	Type         ArtifactType      `json:"type"`
	Size         int64             `json:"size"`
	Checksum     map[string]string `json:"checksum"`
	Signature    *Signature        `json:"signature,omitempty"`
	Location     *StorageLocation  `json:"location"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
}

type BuildStatus struct {
	BuildID   string      `json:"build_id"`
	Status    BuildState  `json:"status"`
	Progress  int         `json:"progress"`
	Message   string      `json:"message"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type BuildInfo struct {
	BuildID   string      `json:"build_id"`
	Version   string      `json:"version"`
	Status    BuildState  `json:"status"`
	Targets   []BuildTarget `json:"targets"`
	StartTime time.Time   `json:"start_time"`
	Duration  time.Duration `json:"duration"`

type BuildFilters struct {
	Status    []BuildState `json:"status,omitempty"`
	Platform  []Platform   `json:"platform,omitempty"`
	Version   string       `json:"version,omitempty"`
	StartDate *time.Time   `json:"start_date,omitempty"`
	EndDate   *time.Time   `json:"end_date,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
}

type BuildError struct {
	Target  BuildTarget `json:"target"`
	Error   string      `json:"error"`
	Stage   string      `json:"stage"`
	Details map[string]interface{} `json:"details,omitempty"`

type BuildState string

const (
	BuildStatePending    BuildState = "pending"
	BuildStateRunning    BuildState = "running"
	BuildStateCompleted  BuildState = "completed"
	BuildStateFailed     BuildState = "failed"
	BuildStateCancelled  BuildState = "cancelled"
)

type ArtifactType string

const (
	ArtifactTypeBinary  ArtifactType = "binary"
	ArtifactTypeArchive ArtifactType = "archive"
	ArtifactTypePackage ArtifactType = "package"
	ArtifactTypeImage   ArtifactType = "image"
)

// Package-related types
type Package struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Version  string          `json:"version"`
	Type     PackageType     `json:"type"`
	Platform Platform        `json:"platform"`
	Arch     Architecture    `json:"architecture"`
	Metadata PackageMetadata `json:"metadata"`
	Artifact *BuildArtifact  `json:"artifact"`
	Location *StorageLocation `json:"location"`
	CreatedAt time.Time      `json:"created_at"`

type PackageInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	LatestVersion string           `json:"latest_version"`
	Description  string            `json:"description"`
	Homepage     string            `json:"homepage"`
	License      string            `json:"license"`
	Platforms    []Platform        `json:"platforms"`
	Downloads    int64             `json:"downloads"`
	LastUpdated  time.Time         `json:"last_updated"`
	Repository   string            `json:"repository"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type PackageFilters struct {
	Platform  []Platform `json:"platform,omitempty"`
	Version   string     `json:"version,omitempty"`
	UpdatedAfter *time.Time `json:"updated_after,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

type PackageType string

const (
	PackageTypeDEB     PackageType = "deb"
	PackageTypeRPM     PackageType = "rpm"
	PackageTypeSnap    PackageType = "snap"
	PackageTypeFlatpak PackageType = "flatpak"
	PackageTypeMSI     PackageType = "msi"
	PackageTypeDMG     PackageType = "dmg"
	PackageTypePKG     PackageType = "pkg"
	PackageTypeTarball PackageType = "tarball"
	PackageTypeZip     PackageType = "zip"
)

// Release-related types
type Release struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Channel     ReleaseChannel     `json:"channel"`
	Description string             `json:"description"`
	Assets      []Asset            `json:"assets"`
	CreatedAt   time.Time          `json:"created_at"`
	PublishedAt *time.Time         `json:"published_at,omitempty"`
	Status      ReleaseStatus      `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`

type Asset struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ContentType  string            `json:"content_type"`
	Size         int64             `json:"size"`
	DownloadURL  string            `json:"download_url"`
	Checksum     map[string]string `json:"checksum"`
	Platform     Platform          `json:"platform"`
	Architecture Architecture      `json:"architecture"`
	CreatedAt    time.Time         `json:"created_at"`
}

type ReleaseFilters struct {
	Channel   []ReleaseChannel `json:"channel,omitempty"`
	Status    []ReleaseStatus  `json:"status,omitempty"`
	Version   string           `json:"version,omitempty"`
	CreatedAfter *time.Time    `json:"created_after,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`

type ReleaseStatus string

const (
	ReleaseStatusDraft     ReleaseStatus = "draft"
	ReleaseStatusPublished ReleaseStatus = "published"
	ReleaseStatusArchived  ReleaseStatus = "archived"
	ReleaseStatusDeleted   ReleaseStatus = "deleted"
)

// Verification-related types
type Signature struct {
	Algorithm string    `json:"algorithm"`
	KeyID     string    `json:"key_id"`
	Signature string    `json:"signature"`
	Timestamp time.Time `json:"timestamp"`

type PublicKey struct {
	ID          string    `json:"id"`
	Algorithm   string    `json:"algorithm"`
	Key         string    `json:"key"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`

// Analytics-related types
type InstallationEvent struct {
	EventID      string            `json:"event_id"`
	Timestamp    time.Time         `json:"timestamp"`
	Version      string            `json:"version"`
	Platform     Platform          `json:"platform"`
	Architecture Architecture      `json:"architecture"`
	Source       string            `json:"source"`
	UserAgent    string            `json:"user_agent"`
	IPAddress    string            `json:"ip_address"`
	Country      string            `json:"country,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`

type UpdateEvent struct {
	EventID         string            `json:"event_id"`
	Timestamp       time.Time         `json:"timestamp"`
	FromVersion     string            `json:"from_version"`
	ToVersion       string            `json:"to_version"`
	Platform        Platform          `json:"platform"`
	Architecture    Architecture      `json:"architecture"`
	UpdateMethod    string            `json:"update_method"`
	Success         bool              `json:"success"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	Duration        time.Duration     `json:"duration"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type UsageEvent struct {
	EventID      string            `json:"event_id"`
	Timestamp    time.Time         `json:"timestamp"`
	Version      string            `json:"version"`
	Command      string            `json:"command"`
	Args         []string          `json:"args"`
	Duration     time.Duration     `json:"duration"`
	Success      bool              `json:"success"`
	ErrorCode    string            `json:"error_code,omitempty"`
	Platform     Platform          `json:"platform"`
	Architecture Architecture      `json:"architecture"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type ErrorEvent struct {
	EventID      string            `json:"event_id"`
	Timestamp    time.Time         `json:"timestamp"`
	Version      string            `json:"version"`
	ErrorType    string            `json:"error_type"`
	ErrorMessage string            `json:"error_message"`
	StackTrace   string            `json:"stack_trace,omitempty"`
	Context      map[string]interface{} `json:"context"`
	Platform     Platform          `json:"platform"`
	Architecture Architecture      `json:"architecture"`
	Metadata     map[string]interface{} `json:"metadata"`

type InstallationStats struct {
	TotalInstalls     int64                    `json:"total_installs"`
	UniqueInstalls    int64                    `json:"unique_installs"`
	ByPlatform        map[Platform]int64       `json:"by_platform"`
	ByVersion         map[string]int64         `json:"by_version"`
	BySource          map[string]int64         `json:"by_source"`
	ByCountry         map[string]int64         `json:"by_country"`
	GrowthRate        float64                  `json:"growth_rate"`
	Period            TimePeriod               `json:"period"`

type UsageStats struct {
	TotalSessions     int64                    `json:"total_sessions"`
	ActiveUsers       int64                    `json:"active_users"`
	ByCommand         map[string]int64         `json:"by_command"`
	ByVersion         map[string]int64         `json:"by_version"`
	AverageSession    time.Duration            `json:"average_session"`
	SuccessRate       float64                  `json:"success_rate"`
	Period            TimePeriod               `json:"period"`
}

type ErrorStats struct {
	TotalErrors       int64                    `json:"total_errors"`
	ByType            map[string]int64         `json:"by_type"`
	ByVersion         map[string]int64         `json:"by_version"`
	ErrorRate         float64                  `json:"error_rate"`
	TopErrors         []ErrorSummary           `json:"top_errors"`
	Period            TimePeriod               `json:"period"`

type ErrorSummary struct {
	ErrorType    string  `json:"error_type"`
	Count        int64   `json:"count"`
	Percentage   float64 `json:"percentage"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
}

type AnalyticsFilters struct {
	Platform     []Platform    `json:"platform,omitempty"`
	Version      []string      `json:"version,omitempty"`
	Source       []string      `json:"source,omitempty"`
	Country      []string      `json:"country,omitempty"`
	StartDate    *time.Time    `json:"start_date,omitempty"`
	EndDate      *time.Time    `json:"end_date,omitempty"`
	Limit        int           `json:"limit,omitempty"`
	Offset       int           `json:"offset,omitempty"`

type AnalyticsReport struct {
	ID               string               `json:"id"`
	Type             ReportType           `json:"type"`
	Period           TimePeriod           `json:"period"`
	GeneratedAt      time.Time            `json:"generated_at"`
	InstallationStats *InstallationStats  `json:"installation_stats,omitempty"`
	UsageStats       *UsageStats          `json:"usage_stats,omitempty"`
	ErrorStats       *ErrorStats          `json:"error_stats,omitempty"`
	Summary          string               `json:"summary"`
	Recommendations  []string             `json:"recommendations"`
}

type ReportType string

const (
	ReportTypeInstallation ReportType = "installation"
	ReportTypeUsage        ReportType = "usage"
	ReportTypeError        ReportType = "error"
	ReportTypeComprehensive ReportType = "comprehensive"
)

type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`

type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatXML  ExportFormat = "xml"
)

// Release management types
type ReleaseDefinition struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Channel      ReleaseChannel      `json:"channel"`
	Artifacts    []BuildArtifact     `json:"artifacts"`
	Description  string              `json:"description"`
	Changelog    string              `json:"changelog"`
	Strategy     ReleaseStrategy     `json:"strategy"`
	HealthChecks []HealthCheck       `json:"health_checks"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type ReleaseExecution struct {
	ID          string               `json:"id"`
	Definition  ReleaseDefinition    `json:"definition"`
	Status      ReleaseExecutionStatus `json:"status"`
	Progress    int                  `json:"progress"`
	StartedAt   time.Time            `json:"started_at"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	Logs        []ReleaseLog         `json:"logs"`
	HealthStatus *HealthStatus       `json:"health_status,omitempty"`

type ReleaseExecutionStatus string

const (
	ReleaseExecutionStatusPending   ReleaseExecutionStatus = "pending"
	ReleaseExecutionStatusRunning   ReleaseExecutionStatus = "running"
	ReleaseExecutionStatusCompleted ReleaseExecutionStatus = "completed"
	ReleaseExecutionStatusFailed    ReleaseExecutionStatus = "failed"
	ReleaseExecutionStatusCancelled ReleaseExecutionStatus = "cancelled"
	ReleaseExecutionStatusRolledBack ReleaseExecutionStatus = "rolled_back"
)

type ReleaseLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`

type HealthStatus struct {
	Overall     HealthState           `json:"overall"`
	Checks      []HealthCheckResult   `json:"checks"`
	Score       float64               `json:"score"`
	UpdatedAt   time.Time             `json:"updated_at"`

type HealthCheckResult struct {
	Name      string      `json:"name"`
	Status    HealthState `json:"status"`
	Message   string      `json:"message"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

type HealthState string

const (
	HealthStateHealthy   HealthState = "healthy"
	HealthStateUnhealthy HealthState = "unhealthy"
	HealthStateUnknown   HealthState = "unknown"
)

type ReleaseInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Channel     ReleaseChannel         `json:"channel"`
	Status      ReleaseExecutionStatus `json:"status"`
	Progress    int                    `json:"progress"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`

// Storage-related types
type StorageLocation struct {
	Provider string `json:"provider"`
	Bucket   string `json:"bucket,omitempty"`
	Key      string `json:"key"`
	URL      string `json:"url,omitempty"`
	Region   string `json:"region,omitempty"`
}

type ArtifactMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Platform     Platform          `json:"platform"`
	Architecture Architecture      `json:"architecture"`
	Size         int64             `json:"size"`
	Checksum     map[string]string `json:"checksum"`
	ContentType  string            `json:"content_type"`
	CreatedAt    time.Time         `json:"created_at"`
	ModifiedAt   time.Time         `json:"modified_at"`
	Tags         map[string]string `json:"tags"`
	Custom       map[string]interface{} `json:"custom"`
}

type StorageFilters struct {
	Platform     []Platform    `json:"platform,omitempty"`
	Architecture []Architecture `json:"architecture,omitempty"`
	Version      string        `json:"version,omitempty"`
	CreatedAfter *time.Time    `json:"created_after,omitempty"`
	CreatedBefore *time.Time   `json:"created_before,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Limit        int           `json:"limit,omitempty"`
	Offset       int           `json:"offset,omitempty"`

type RetentionPolicy struct {
	MaxAge       time.Duration `json:"max_age"`
	MaxVersions  int           `json:"max_versions"`
	MinVersions  int           `json:"min_versions"`
	KeepLatest   bool          `json:"keep_latest"`

type StorageUsage struct {
	TotalSize    int64     `json:"total_size"`
	TotalFiles   int64     `json:"total_files"`
	ByPlatform   map[Platform]int64 `json:"by_platform"`
	ByVersion    map[string]int64   `json:"by_version"`
	LastUpdated  time.Time `json:"last_updated"`
