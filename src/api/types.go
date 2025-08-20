package api

import (
	"time"
)

// API Version
const APIVersion = "v1"

// Common error codes used across the API
const (
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimited        = "RATE_LIMITED"
	ErrCodeTimeout            = "REQUEST_TIMEOUT"
)

// Standard API Response wrapper
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Error represents API error details
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta contains pagination and other metadata
type Meta struct {
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	Total      int    `json:"total,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	Version    string `json:"version,omitempty"`
}

// Scan represents a security scan operation
type Scan struct {
	ID          string       `json:"id"`
	Status      ScanStatus   `json:"status"`
	Target      ScanTarget   `json:"target"`
	Templates   []string     `json:"templates,omitempty"`
	Categories  []string     `json:"categories,omitempty"`
	Config      ScanConfig   `json:"config"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Duration    string       `json:"duration,omitempty"`
	Results     *ScanResults `json:"results,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ScanStatus represents the current state of a scan
type ScanStatus string

const (
	ScanStatusPending    ScanStatus = "pending"
	ScanStatusRunning    ScanStatus = "running"
	ScanStatusCompleted  ScanStatus = "completed"
	ScanStatusFailed     ScanStatus = "failed"
	ScanStatusCancelled  ScanStatus = "cancelled"
)

// ScanTarget defines what is being scanned
type ScanTarget struct {
	Type       string                 `json:"type"` // "api", "model", "application"
	Provider   string                 `json:"provider,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
	Model      string                 `json:"model,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ScanConfig contains scan configuration options
type ScanConfig struct {
	Concurrency   int                    `json:"concurrency,omitempty"`
	Timeout       int                    `json:"timeout,omitempty"` // seconds
	MaxRetries    int                    `json:"max_retries,omitempty"`
	RateLimit     int                    `json:"rate_limit,omitempty"` // requests per minute
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

// ScanResults contains the findings from a scan
type ScanResults struct {
	Summary      ResultSummary       `json:"summary"`
	Findings     []Finding           `json:"findings"`
	Errors       []ScanError         `json:"errors,omitempty"`
	TemplateRuns []TemplateExecution `json:"template_runs"`
}

// ResultSummary provides high-level scan results
type ResultSummary struct {
	TotalTests       int            `json:"total_tests"`
	Passed           int            `json:"passed"`
	Failed           int            `json:"failed"`
	Errors           int            `json:"errors"`
	Skipped          int            `json:"skipped"`
	SeverityCount    map[string]int `json:"severity_count"`
	CategoryCount    map[string]int `json:"category_count"`
	ComplianceScore  float64        `json:"compliance_score"`
}

// Finding represents a security issue discovered
type Finding struct {
	ID           string                 `json:"id"`
	TemplateID   string                 `json:"template_id"`
	TemplateName string                 `json:"template_name"`
	Category     string                 `json:"category"`
	Severity     string                 `json:"severity"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Evidence     map[string]interface{} `json:"evidence,omitempty"`
	Remediation  string                 `json:"remediation,omitempty"`
	References   []string               `json:"references,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ScanError represents an error during scanning
type ScanError struct {
	TemplateID string    `json:"template_id,omitempty"`
	Error      string    `json:"error"`
	Details    string    `json:"details,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// TemplateExecution tracks individual template runs
type TemplateExecution struct {
	TemplateID string    `json:"template_id"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   string    `json:"duration"`
}

// Template represents a security test template
type Template struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Severity     string                 `json:"severity"`
	Author       string                 `json:"author,omitempty"`
	Version      string                 `json:"version"`
	Tags         []string               `json:"tags,omitempty"`
	References   []string               `json:"references,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	LastUpdated  time.Time              `json:"last_updated"`
}

// Module represents a provider module
type Module struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Version      string    `json:"version"`
	Description  string    `json:"description"`
	Provider     string    `json:"provider"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities,omitempty"`
	Config       ModuleConfig `json:"config,omitempty"`
	LoadedAt     time.Time `json:"loaded_at"`
}

// ModuleConfig contains module configuration
type ModuleConfig struct {
	Enabled     bool                   `json:"enabled"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Credentials map[string]string      `json:"credentials,omitempty"` // Will be filtered in responses
}

// VersionInfo contains version information
type VersionInfo struct {
	Core      ComponentVersion `json:"core"`
	Templates ComponentVersion `json:"templates"`
	Modules   ComponentVersion `json:"modules"`
	API       ComponentVersion `json:"api"`
}

// ComponentVersion represents version details for a component
type ComponentVersion struct {
	Current        string    `json:"current"`
	Latest         string    `json:"latest,omitempty"`
	UpdateAvailable bool     `json:"update_available"`
	ReleaseDate    time.Time `json:"release_date,omitempty"`
	Changelog      string    `json:"changelog_url,omitempty"`
}

// UpdateRequest represents an update operation request
type UpdateRequest struct {
	Components []string `json:"components,omitempty"` // ["core", "templates", "modules"]
	Force      bool     `json:"force,omitempty"`
	DryRun     bool     `json:"dry_run,omitempty"`
}

// UpdateResponse contains update operation results
type UpdateResponse struct {
	Status   string          `json:"status"`
	Updates  []UpdateResult  `json:"updates"`
	Messages []string        `json:"messages,omitempty"`
}

// UpdateResult represents the result of updating a component
type UpdateResult struct {
	Component   string `json:"component"`
	OldVersion  string `json:"old_version"`
	NewVersion  string `json:"new_version"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

// CreateScanRequest represents a request to create a new scan
type CreateScanRequest struct {
	Target     ScanTarget `json:"target"`
	Templates  []string   `json:"templates,omitempty"`
	Categories []string   `json:"categories,omitempty"`
	Config     ScanConfig `json:"config,omitempty"`
}

// ListScansRequest represents scan listing parameters
type ListScansRequest struct {
	Status   ScanStatus `json:"status,omitempty"`
	Page     int        `json:"page,omitempty"`
	PerPage  int        `json:"per_page,omitempty"`
	SortBy   string     `json:"sort_by,omitempty"`
	OrderBy  string     `json:"order_by,omitempty"`
}

// ListTemplatesRequest represents template listing parameters
type ListTemplatesRequest struct {
	Category string   `json:"category,omitempty"`
	Severity string   `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Search   string   `json:"search,omitempty"`
	Page     int      `json:"page,omitempty"`
	PerPage  int      `json:"per_page,omitempty"`
}

// AuthRequest represents authentication request
type AuthRequest struct {
	APIKey string `json:"api_key,omitempty"`
	Token  string `json:"token,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success     bool      `json:"success"`
	Token       string    `json:"token,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
}

// Filter types for API operations

// TemplateFilter represents filtering criteria for templates
type TemplateFilter struct {
	Category string   `json:"category,omitempty"`
	Severity string   `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Search   string   `json:"search,omitempty"`
	Author   string   `json:"author,omitempty"`
	Version  string   `json:"version,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

// ScanFilter represents filtering criteria for scans
type ScanFilter struct {
	Status    ScanStatus `json:"status,omitempty"`
	Target    string     `json:"target,omitempty"`
	Provider  string     `json:"provider,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DateFrom  *time.Time `json:"date_from,omitempty"`
	DateTo    *time.Time `json:"date_to,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// Bundle types for API operations

// BundleInfo represents metadata about a bundle
type BundleInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Tags        []string               `json:"tags,omitempty"`
	Templates   []string               `json:"templates"`
	Modules     []string               `json:"modules,omitempty"`
	Size        int64                  `json:"size"`
	Checksum    string                 `json:"checksum"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExportBundleRequest represents a request to export a bundle
type ExportBundleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Templates   []string `json:"templates,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Modules     []string `json:"modules,omitempty"`
	Format      string   `json:"format,omitempty"` // "zip", "tar.gz"
	Compress    bool     `json:"compress,omitempty"`
}

// ImportBundleRequest represents a request to import a bundle
type ImportBundleRequest struct {
	Source        string `json:"source"` // file path or URL
	ValidateOnly  bool   `json:"validate_only,omitempty"`
	Overwrite     bool   `json:"overwrite,omitempty"`
	SkipConflicts bool   `json:"skip_conflicts,omitempty"`
}

// BundleOperationResult represents the result of a bundle operation
type BundleOperationResult struct {
	BundleID   string            `json:"bundle_id,omitempty"`
	Status     string            `json:"status"`
	Message    string            `json:"message,omitempty"`
	Templates  []string          `json:"templates,omitempty"`
	Modules    []string          `json:"modules,omitempty"`
	Conflicts  []string          `json:"conflicts,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Compliance types for API operations

// ComplianceReportRequest represents a request to generate a compliance report
type ComplianceReportRequest struct {
	Framework    string   `json:"framework"`     // "owasp", "iso42001", "nist"
	ScanIDs      []string `json:"scan_ids,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	Format       string   `json:"format,omitempty"` // "json", "pdf", "html", "csv"
	IncludePassed bool    `json:"include_passed,omitempty"`
	IncludeFailed bool    `json:"include_failed,omitempty"`
	IncludeSkipped bool   `json:"include_skipped,omitempty"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ID          string                 `json:"id"`
	Framework   string                 `json:"framework"`
	GeneratedAt time.Time              `json:"generated_at"`
	Period      CompliancePeriod       `json:"period"`
	Summary     ComplianceSummary      `json:"summary"`
	Results     []ComplianceResult     `json:"results"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CompliancePeriod represents the time period covered by a compliance report
type CompliancePeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ComplianceSummary provides high-level compliance information
type ComplianceSummary struct {
	TotalControls     int     `json:"total_controls"`
	CompliantControls int     `json:"compliant_controls"`
	FailedControls    int     `json:"failed_controls"`
	SkippedControls   int     `json:"skipped_controls"`
	ComplianceScore   float64 `json:"compliance_score"`
	RiskLevel         string  `json:"risk_level"`
}

// ComplianceResult represents the compliance status of a specific control
type ComplianceResult struct {
	ControlID    string                 `json:"control_id"`
	ControlName  string                 `json:"control_name"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"` // "compliant", "non-compliant", "not-applicable"
	Evidence     []ComplianceEvidence   `json:"evidence,omitempty"`
	Recommendations []string             `json:"recommendations,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceEvidence represents evidence supporting a compliance result
type ComplianceEvidence struct {
	Type        string                 `json:"type"` // "scan", "template", "manual"
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// ComplianceStatus represents overall compliance status
type ComplianceStatus struct {
	Framework       string                 `json:"framework"`
	OverallScore    float64                `json:"overall_score"`
	RiskLevel       string                 `json:"risk_level"`
	LastAssessment  time.Time              `json:"last_assessment"`
	ControlsSummary ComplianceSummary      `json:"controls_summary"`
	Trends          []ComplianceTrend      `json:"trends,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceTrend represents compliance trends over time
type ComplianceTrend struct {
	Date  time.Time `json:"date"`
	Score float64   `json:"score"`
}

// Service interfaces and types

// Services contains all service dependencies for the API server
type Services struct {
	ScanEngine        ScanService
	TemplateManager   TemplateService
	ModuleManager     ModuleService
	UpdateManager     UpdateService
	BundleManager     BundleService
	ComplianceManager ComplianceService
}

// ScanService interface for scan operations
type ScanService interface {
	CreateScan(request CreateScanRequest) (*Scan, error)
	GetScan(id string) (*Scan, error)
	ListScans(filter ScanFilter) ([]Scan, error)
	CancelScan(id string) error
	GetScanResults(id string) (*ScanResults, error)
}

// TemplateService interface for template operations
type TemplateService interface {
	ListTemplates(filter TemplateFilter) ([]Template, error)
	GetTemplate(id string) (*Template, error)
	GetCategories() ([]string, error)
	ValidateTemplate(template *Template) error
}

// ModuleService interface for module operations
type ModuleService interface {
	ListModules() ([]Module, error)
	GetModule(id string) (*Module, error)
	UpdateModuleConfig(id string, config ModuleConfig) error
	ReloadModule(id string) error
}

// UpdateService interface for update operations
type UpdateService interface {
	CheckForUpdates() (*VersionInfo, error)
	PerformUpdate(request UpdateRequest) (*UpdateResponse, error)
	GetUpdateHistory() ([]UpdateResult, error)
}

// BundleService interface for bundle operations
type BundleService interface {
	ListBundles() ([]BundleInfo, error)
	GetBundle(id string) (*BundleInfo, error)
	ExportBundle(request ExportBundleRequest) (*BundleOperationResult, error)
	ImportBundle(request ImportBundleRequest) (*BundleOperationResult, error)
	DeleteBundle(id string) error
}

// ComplianceService interface for compliance operations
type ComplianceService interface {
	GenerateReport(request ComplianceReportRequest) (*ComplianceReport, error)
	CheckCompliance(framework string) (*ComplianceStatus, error)
	GetComplianceHistory(framework string, days int) ([]ComplianceTrend, error)
}
