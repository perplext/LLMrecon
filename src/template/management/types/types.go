// Package types provides common types for the template management system
package types

import (
)

// TemplateSourceType represents the type of template source
type TemplateSourceType string

const (
	// LocalSourceType represents templates from the local filesystem
	LocalSourceType TemplateSourceType = "local"
	// RemoteSourceType represents templates from a remote repository
	RemoteSourceType TemplateSourceType = "remote"
	// DatabaseSourceType represents templates from a database
	DatabaseSourceType TemplateSourceType = "database"
)

// TemplateStatus represents the status of a template
type TemplateStatus string

const (
	// StatusLoaded indicates the template has been loaded
	StatusLoaded TemplateStatus = "loaded"
	// StatusValidated indicates the template has been validated
	StatusValidated TemplateStatus = "validated"
	// StatusExecuting indicates the template is being executed
	StatusExecuting TemplateStatus = "executing"
	// StatusCompleted indicates the template execution has completed
	StatusCompleted TemplateStatus = "completed"
	// StatusFailed indicates the template execution has failed
	StatusFailed TemplateStatus = "failed"
)

// TemplateResult represents the result of a template execution
type TemplateResult struct {
	// TemplateID is the ID of the template
	TemplateID string `json:"template_id"`
	// TemplateName is the name of the template
	TemplateName string `json:"template_name"`
	// Description is the description of the template
	Description string `json:"description"`
	// Status is the status of the template
	Status TemplateStatus `json:"status"`
	// StartTime is the time the template execution started
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the template execution ended
	EndTime time.Time `json:"end_time"`
	// Duration is the duration of the template execution
	Duration time.Duration `json:"duration"`
	// Error is any error that occurred during template execution
	Error error `json:"error,omitempty"`
	// Response is the response from the LLM
	Response string `json:"response,omitempty"`
	// Detected indicates whether the vulnerability was detected
	Detected bool `json:"detected"`
	// Score is the score of the template execution (0-100)
	Score int `json:"score"`
	// Details contains additional details about the template execution
	Details map[string]interface{} `json:"details,omitempty"`
	// Tags contains tags associated with the template
	Tags []string `json:"tags,omitempty"`
	// Metadata contains additional metadata about the template
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Input contains the input to the template
	Input string `json:"input,omitempty"`
	// Output contains the expected output from the template
	Output string `json:"output,omitempty"`
}

// Note: TemplateManager and TemplateLoader interfaces are defined in interfaces.go
// to avoid duplicate declarations
