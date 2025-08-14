// Package scan provides API endpoints for managing red-team scans
package scan

import (
	"time")

// ScanStatus represents the status of a scan
type ScanStatus string

const (
	// ScanStatusPending indicates the scan is pending execution
	ScanStatusPending ScanStatus = "pending"
	// ScanStatusRunning indicates the scan is currently running
	ScanStatusRunning ScanStatus = "running"
	// ScanStatusCompleted indicates the scan has completed successfully
	ScanStatusCompleted ScanStatus = "completed"
	// ScanStatusFailed indicates the scan has failed
	ScanStatusFailed ScanStatus = "failed"
	// ScanStatusCancelled indicates the scan was cancelled
	ScanStatusCancelled ScanStatus = "cancelled"
)

// ScanSeverity represents the severity level of a scan finding
type ScanSeverity string

const (
	// ScanSeverityLow indicates a low severity finding
	ScanSeverityLow ScanSeverity = "low"
	// ScanSeverityMedium indicates a medium severity finding
	ScanSeverityMedium ScanSeverity = "medium"
	// ScanSeverityHigh indicates a high severity finding
	ScanSeverityHigh ScanSeverity = "high"
	// ScanSeverityCritical indicates a critical severity finding
	ScanSeverityCritical ScanSeverity = "critical"
)

// ScanConfig represents the configuration for a scan
type ScanConfig struct {
	// ID is the unique identifier for the scan configuration
	ID string `json:"id"`
	// Name is the name of the scan configuration
	Name string `json:"name"`
	// Description is a description of the scan configuration
	Description string `json:"description"`
	// Target is the target of the scan (e.g., a prompt, system, or model)
	Target string `json:"target"`
	// TargetType is the type of target (e.g., "prompt", "system", "model")
	TargetType string `json:"target_type"`
	// Templates is a list of template IDs to use for the scan
	Templates []string `json:"templates"`
	// Parameters is a map of parameters for the scan
	Parameters map[string]interface{} `json:"parameters"`
	// CreatedAt is the time the scan configuration was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the time the scan configuration was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// CreatedBy is the user who created the scan configuration
	CreatedBy string `json:"created_by"`
}

// Scan represents a scan execution
type Scan struct {
	// ID is the unique identifier for the scan
	ID string `json:"id"`
	// ConfigID is the ID of the scan configuration
	ConfigID string `json:"config_id"`
	// Status is the status of the scan
	Status ScanStatus `json:"status"`
	// StartTime is the time the scan started
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the scan ended
	EndTime time.Time `json:"end_time,omitempty"`
	// Progress is the progress of the scan (0-100)
	Progress int `json:"progress"`
	// Error is the error message if the scan failed
	Error string `json:"error,omitempty"`
	// Results is the results of the scan
	Results []ScanResult `json:"results,omitempty"`
}

// ScanResult represents a result from a scan
type ScanResult struct {
	// ID is the unique identifier for the scan result
	ID string `json:"id"`
	// ScanID is the ID of the scan
	ScanID string `json:"scan_id"`
	// TemplateID is the ID of the template that generated the result
	TemplateID string `json:"template_id"`
	// Severity is the severity of the finding
	Severity ScanSeverity `json:"severity"`
	// Title is the title of the finding
	Title string `json:"title"`
	// Description is a description of the finding
	Description string `json:"description"`
	// Details contains detailed information about the finding
	Details map[string]interface{} `json:"details"`
	// Timestamp is the time the result was generated
	Timestamp time.Time `json:"timestamp"`
}

// PaginationParams represents pagination parameters for list endpoints
type PaginationParams struct {
	// Page is the page number (1-based)
	Page int `json:"page"`
	// PageSize is the number of items per page
	PageSize int `json:"page_size"`
	// TotalItems is the total number of items
	TotalItems int `json:"total_items"`
	// TotalPages is the total number of pages
	TotalPages int `json:"total_pages"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	// Pagination contains pagination information
	Pagination PaginationParams `json:"pagination"`
	// Data contains the response data
	Data interface{} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	// Error is the error message
	Error string `json:"error"`
	// Code is the error code
	Code int `json:"code"`
}

// CreateScanConfigRequest represents a request to create a scan configuration
type CreateScanConfigRequest struct {
	// Name is the name of the scan configuration
	Name string `json:"name" validate:"required"`
	// Description is a description of the scan configuration
	Description string `json:"description"`
	// Target is the target of the scan (e.g., a prompt, system, or model)
	Target string `json:"target" validate:"required"`
	// TargetType is the type of target (e.g., "prompt", "system", "model")
	TargetType string `json:"target_type" validate:"required"`
	// Templates is a list of template IDs to use for the scan
	Templates []string `json:"templates" validate:"required,min=1"`
	// Parameters is a map of parameters for the scan
	Parameters map[string]interface{} `json:"parameters"`
}

// UpdateScanConfigRequest represents a request to update a scan configuration
type UpdateScanConfigRequest struct {
	// Name is the name of the scan configuration
	Name string `json:"name"`
	// Description is a description of the scan configuration
	Description string `json:"description"`
	// Target is the target of the scan (e.g., a prompt, system, or model)
	Target string `json:"target"`
	// TargetType is the type of target (e.g., "prompt", "system", "model")
	TargetType string `json:"target_type"`
	// Templates is a list of template IDs to use for the scan
	Templates []string `json:"templates"`
	// Parameters is a map of parameters for the scan
	Parameters map[string]interface{} `json:"parameters"`
}

// CreateScanRequest represents a request to create a scan
type CreateScanRequest struct {
	// ConfigID is the ID of the scan configuration
	ConfigID string `json:"config_id" validate:"required"`
}

// FilterParams represents filter parameters for list endpoints
type FilterParams struct {
	// Status filters results by status
	Status string `json:"status"`
	// Severity filters results by severity
	Severity string `json:"severity"`
	// StartDate filters results by start date
	StartDate string `json:"start_date"`
	// EndDate filters results by end date
	EndDate string `json:"end_date"`
	// Search is a search term to filter results
	Search string `json:"search"`
}
