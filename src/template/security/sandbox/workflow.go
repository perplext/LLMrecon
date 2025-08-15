package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// ApprovalStatus represents the approval status of a template
type ApprovalStatus string

const (
	// StatusDraft indicates a template is in draft status
	StatusDraft ApprovalStatus = "draft"
	// StatusPendingReview indicates a template is pending review
	StatusPendingReview ApprovalStatus = "pending_review"
	// StatusApproved indicates a template is approved
	StatusApproved ApprovalStatus = "approved"
	// StatusRejected indicates a template is rejected
	StatusRejected ApprovalStatus = "rejected"
	// StatusDeprecated indicates a template is deprecated
	StatusDeprecated ApprovalStatus = "deprecated"
)

// TemplateVersion represents a version of a template
type TemplateVersion struct {
	// ID is the version ID
	ID string
	// TemplateID is the ID of the template
	TemplateID string
	// Version is the version number
	Version string
	// Content is the template content
	Content string
	// Status is the approval status
	Status ApprovalStatus
	// CreatedAt is the creation time
	CreatedAt time.Time
	// UpdatedAt is the last update time
	UpdatedAt time.Time
	// ApprovedBy is the user who approved the template
	ApprovedBy string
	// ApprovedAt is the approval time
	ApprovedAt time.Time
	// RiskScore is the risk score
	RiskScore *RiskScore
	// SecurityIssues are the security issues found in the template
	SecurityIssues []*security.SecurityIssue
	// Comments are comments on the template
	Comments []string

// ApprovalWorkflow manages the template approval workflow
type ApprovalWorkflow struct {
	// validator is the template validator
	validator *TemplateValidator
	// scorer is the template scorer
	scorer *TemplateScorer
	// versions is a map of template versions
	versions map[string][]*TemplateVersion
	// approvers is a list of users who can approve templates
	approvers []string
	// storageDir is the directory for storing template versions
	storageDir string

// NewApprovalWorkflow creates a new approval workflow
func NewApprovalWorkflow(validator *TemplateValidator, scorer *TemplateScorer, storageDir string) *ApprovalWorkflow {
	return &ApprovalWorkflow{
		validator:  validator,
		scorer:     scorer,
		versions:   make(map[string][]*TemplateVersion),
		approvers:  []string{},
		storageDir: storageDir,
	}

// AddApprover adds an approver to the workflow
func (w *ApprovalWorkflow) AddApprover(approver string) {
	w.approvers = append(w.approvers, approver)

// IsApprover checks if a user is an approver
func (w *ApprovalWorkflow) IsApprover(user string) bool {
	for _, approver := range w.approvers {
		if approver == user {
			return true
		}
	}
	return false

// CreateVersion creates a new version of a template
func (w *ApprovalWorkflow) CreateVersion(ctx context.Context, template *format.Template, user string) (*TemplateVersion, error) {
	// Validate the template
	issues, err := w.validator.Validate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}
	
	// Score the template
	riskScore := w.scorer.ScoreTemplate(template, issues)
	
	// Get the current versions
	versions := w.versions[template.ID]
	
	// Determine the new version number
	var version string
	if len(versions) == 0 {
		version = "1.0.0"
	} else {
		lastVersion := versions[len(versions)-1]
		// Simple version increment for now
		version = fmt.Sprintf("1.0.%d", len(versions))
	}
	
	// Create the new version
	templateVersion := &TemplateVersion{
		ID:            fmt.Sprintf("%s-v%s", template.ID, version),
		TemplateID:    template.ID,
		Version:       version,
		Content:       template.Content,
		Status:        StatusDraft,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		RiskScore:     riskScore,
		SecurityIssues: issues,
		Comments:      []string{},
	}
	
	// Add the version to the map
	w.versions[template.ID] = append(w.versions[template.ID], templateVersion)
	
	// Save the version to disk
	if err := w.saveVersion(templateVersion); err != nil {
		return nil, fmt.Errorf("failed to save template version: %w", err)
	}
	
	return templateVersion, nil

// saveVersion saves a template version to disk
func (w *ApprovalWorkflow) saveVersion(version *TemplateVersion) error {
	if w.storageDir == "" {
		return nil
	}
	
	// Create the template directory if it doesn't exist
	templateDir := filepath.Join(w.storageDir, version.TemplateID)
	if err := os.MkdirAll(templateDir, 0700); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}
	
	// Write the template content to a file
	filename := fmt.Sprintf("%s.tmpl", version.ID)
	filePath := filepath.Join(templateDir, filename)
	if err := ioutil.WriteFile(filePath, []byte(version.Content), 0600); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}
	
	return nil
	

// SubmitForReview submits a template version for review
func (w *ApprovalWorkflow) SubmitForReview(templateID, versionID, user string) error {
	// Find the template version
	version, err := w.findVersion(templateID, versionID)
	if err != nil {
		return err
	}
	
	// Check if the template is in draft status
	if version.Status != StatusDraft {
		return fmt.Errorf("template version is not in draft status")
	}
	
	// Update the status
	version.Status = StatusPendingReview
	version.UpdatedAt = time.Now()
	
	// Add a comment
	version.Comments = append(version.Comments, fmt.Sprintf("Submitted for review by %s at %s", user, version.UpdatedAt.Format(time.RFC3339)))
	
	// Save the version
	if err := w.saveVersion(version); err != nil {
		return fmt.Errorf("failed to save template version: %w", err)
	}
	
	return nil

// ApproveVersion approves a template version
func (w *ApprovalWorkflow) ApproveVersion(templateID, versionID, user string) error {
	// Find the template version
	version, err := w.findVersion(templateID, versionID)
	if err != nil {
		return err
	}
	
	// Check if the template is pending review
	if version.Status != StatusPendingReview {
		return fmt.Errorf("template version is not pending review")
	}
	
	// Check if the user is an approver
	if !w.IsApprover(user) {
		return fmt.Errorf("user is not an approver")
	}
	
	// Update the status
	version.Status = StatusApproved
	version.UpdatedAt = time.Now()
	version.ApprovedBy = user
	version.ApprovedAt = time.Now()
	
	// Add a comment
	version.Comments = append(version.Comments, fmt.Sprintf("Approved by %s at %s", user, version.ApprovedAt.Format(time.RFC3339)))
	
	// Save the version
	if err := w.saveVersion(version); err != nil {
		return fmt.Errorf("failed to save template version: %w", err)
	}
	
	return nil

// RejectVersion rejects a template version
func (w *ApprovalWorkflow) RejectVersion(templateID, versionID, user, reason string) error {
	// Find the template version
	version, err := w.findVersion(templateID, versionID)
	if err != nil {
		return err
	}
	
	// Check if the template is pending review
	if version.Status != StatusPendingReview {
		return fmt.Errorf("template version is not pending review")
	}
	
	// Check if the user is an approver
	if !w.IsApprover(user) {
		return fmt.Errorf("user is not an approver")
	}
	
	// Update the status
	version.Status = StatusRejected
	version.UpdatedAt = time.Now()
	
	// Add a comment
	comment := fmt.Sprintf("Rejected by %s at %s", user, version.UpdatedAt.Format(time.RFC3339))
	if reason != "" {
		comment += fmt.Sprintf(": %s", reason)
	}
	version.Comments = append(version.Comments, comment)
	
	// Save the version
	if err := w.saveVersion(version); err != nil {
		return fmt.Errorf("failed to save template version: %w", err)
	}
	
	return nil

// DeprecateVersion deprecates a template version
func (w *ApprovalWorkflow) DeprecateVersion(templateID, versionID, user, reason string) error {
	// Find the template version
	version, err := w.findVersion(templateID, versionID)
	if err != nil {
		return err
	}
	
	// Check if the template is approved
	if version.Status != StatusApproved {
		return fmt.Errorf("template version is not approved")
	}
	
	// Check if the user is an approver
	if !w.IsApprover(user) {
		return fmt.Errorf("user is not an approver")
	}
	
	// Update the status
	version.Status = StatusDeprecated
	version.UpdatedAt = time.Now()
	
	// Add a comment
	comment := fmt.Sprintf("Deprecated by %s at %s", user, version.UpdatedAt.Format(time.RFC3339))
	if reason != "" {
		comment += fmt.Sprintf(": %s", reason)
	}
	version.Comments = append(version.Comments, comment)
	
	// Save the version
	if err := w.saveVersion(version); err != nil {
		return fmt.Errorf("failed to save template version: %w", err)
	}
	
	return nil

// GetVersion gets a template version
func (w *ApprovalWorkflow) GetVersion(templateID, versionID string) (*TemplateVersion, error) {
	return w.findVersion(templateID, versionID)

// GetVersions gets all versions of a template
func (w *ApprovalWorkflow) GetVersions(templateID string) ([]*TemplateVersion, error) {
	versions, ok := w.versions[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	return versions, nil

// GetLatestVersion gets the latest version of a template
func (w *ApprovalWorkflow) GetLatestVersion(templateID string) (*TemplateVersion, error) {
	versions, ok := w.versions[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for template: %s", templateID)
	}
	
	return versions[len(versions)-1], nil

// GetLatestApprovedVersion gets the latest approved version of a template
func (w *ApprovalWorkflow) GetLatestApprovedVersion(templateID string) (*TemplateVersion, error) {
	versions, ok := w.versions[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	// Find the latest approved version
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].Status == StatusApproved {
			return versions[i], nil
		}
	}
	
	return nil, fmt.Errorf("no approved versions found for template: %s", templateID)
	

// findVersion finds a template version
func (w *ApprovalWorkflow) findVersion(templateID, versionID string) (*TemplateVersion, error) {
	versions, ok := w.versions[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	
	for _, version := range versions {
		if version.ID == versionID {
			return version, nil
		}
	}
	
	return nil, fmt.Errorf("version not found: %s", versionID)

// AddComment adds a comment to a template version
func (w *ApprovalWorkflow) AddComment(templateID, versionID, user, comment string) error {
	// Find the template version
	version, err := w.findVersion(templateID, versionID)
	if err != nil {
		return err
	}
	
	// Add the comment
	version.Comments = append(version.Comments, fmt.Sprintf("%s (%s): %s", user, time.Now().Format(time.RFC3339), comment))
	
	// Save the version
	if err := w.saveVersion(version); err != nil {
		return fmt.Errorf("failed to save template version: %w", err)
	}
	
