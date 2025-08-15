// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

// EnhancedApprovalWorkflow extends the ApprovalWorkflow with more sophisticated approval capabilities
type EnhancedApprovalWorkflow struct {
	*ApprovalWorkflow
	config              *ProtectionConfig
	approvalConfig      *EnhancedApprovalConfig
	pendingRequests     map[string]*EnhancedApprovalRequest
	approvalHistory     []*EnhancedApprovalRequest
	approvalHandlers    map[string]ApprovalHandlerFunc
	maxPendingRequests  int
	maxApprovalHistory  int
	defaultExpiration   time.Duration
	dataDir             string
	mu                  sync.RWMutex
}

// EnhancedApprovalConfig defines the configuration for enhanced approval workflow
type EnhancedApprovalConfig struct {
	EnableAutoApproval      bool                   `json:"enable_auto_approval"`
	EnableApprovalHistory   bool                   `json:"enable_approval_history"`
	EnableApprovalExpiration bool                  `json:"enable_approval_expiration"`
	ApprovalThresholds      map[string]float64     `json:"approval_thresholds"`
	AutoApprovalRules       map[string]AutoApprovalRule `json:"auto_approval_rules"`
	ApprovalLevels          map[string]ApprovalLevel `json:"approval_levels"`
	DefaultApprovalLevel    string                 `json:"default_approval_level"`
	ApprovalTimeout         time.Duration          `json:"approval_timeout"`

// EnhancedApprovalRequest extends the ApprovalRequest with more information
type EnhancedApprovalRequest struct {
	*ApprovalRequest
	Status              ApprovalStatus       `json:"status"`
	ApprovalLevel       string               `json:"approval_level"`
	ApproverID          string               `json:"approver_id,omitempty"`
	ApprovalTime        time.Time            `json:"approval_time,omitempty"`
	ExpirationTime      time.Time            `json:"expiration_time"`
	ApprovalHistory     []*ApprovalAction    `json:"approval_history,omitempty"`
	AutoApprovalEligible bool                `json:"auto_approval_eligible"`
	AutoApprovalReason   string              `json:"auto_approval_reason,omitempty"`
}

// ApprovalStatus defines the status of an approval request
type ApprovalStatus string

const (
	// ApprovalStatusPending indicates the approval request is pending
	ApprovalStatusPending ApprovalStatus = "pending"
	// ApprovalStatusApproved indicates the approval request was approved
	ApprovalStatusApproved ApprovalStatus = "approved"
	// ApprovalStatusRejected indicates the approval request was rejected
	ApprovalStatusRejected ApprovalStatus = "rejected"
	// ApprovalStatusExpired indicates the approval request expired
	ApprovalStatusExpired ApprovalStatus = "expired"
	// ApprovalStatusCancelled indicates the approval request was cancelled
	ApprovalStatusCancelled ApprovalStatus = "cancelled"
)

// ApprovalAction defines an action taken on an approval request
type ApprovalAction struct {
	ActionType          string               `json:"action_type"`
	ActorID             string               `json:"actor_id"`
	Timestamp           time.Time            `json:"timestamp"`
	Reason              string               `json:"reason,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// AutoApprovalRule defines a rule for auto-approval
type AutoApprovalRule struct {
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	MaxRiskScore        float64              `json:"max_risk_score"`
	AllowedDetectionTypes []DetectionType    `json:"allowed_detection_types"`
	DisallowedDetectionTypes []DetectionType `json:"disallowed_detection_types"`
	RequiredMetadata    map[string]interface{} `json:"required_metadata,omitempty"`
	ApplicableUsers     []string             `json:"applicable_users,omitempty"`
	Enabled             bool                 `json:"enabled"`

// ApprovalLevel defines a level of approval
type ApprovalLevel struct {
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	RequiredApprovers   int                  `json:"required_approvers"`
	AllowedApprovers    []string             `json:"allowed_approvers,omitempty"`
	MaxRiskScore        float64              `json:"max_risk_score"`
	ApprovalTimeout     time.Duration        `json:"approval_timeout"`

// ApprovalHandlerFunc defines a function that handles approval requests
type ApprovalHandlerFunc func(context.Context, *EnhancedApprovalRequest) (bool, error)

// NewEnhancedApprovalWorkflow creates a new enhanced approval workflow
func NewEnhancedApprovalWorkflow(config *ProtectionConfig, dataDir string) (*EnhancedApprovalWorkflow, error) {
	baseWorkflow := NewApprovalWorkflow(config)
	
	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Initialize approval config
	approvalConfig := &EnhancedApprovalConfig{
		EnableAutoApproval:      true,
		EnableApprovalHistory:   true,
		EnableApprovalExpiration: true,
		ApprovalThresholds:      make(map[string]float64),
		AutoApprovalRules:       make(map[string]AutoApprovalRule),
		ApprovalLevels:          make(map[string]ApprovalLevel),
		DefaultApprovalLevel:    "standard",
		ApprovalTimeout:         time.Hour * 24,
	}
	
	// Set default approval thresholds
	approvalConfig.ApprovalThresholds["low"] = 0.5
	approvalConfig.ApprovalThresholds["medium"] = 0.7
	approvalConfig.ApprovalThresholds["high"] = 0.9
	
	// Set default auto-approval rules
	approvalConfig.AutoApprovalRules["low_risk"] = AutoApprovalRule{
		Name:                "Low Risk Auto-Approval",
		Description:         "Auto-approve low-risk requests",
		MaxRiskScore:        0.3,
		AllowedDetectionTypes: []DetectionType{
			DetectionTypeUnusualPattern,
			DetectionTypeDelimiterMisuse,
		},
		DisallowedDetectionTypes: []DetectionType{
			DetectionTypeJailbreak,
			DetectionTypePromptInjection,
			DetectionTypeSystemPrompt,
			DetectionTypeRoleChange,
		},
		Enabled:             true,
	}
	
	// Set default approval levels
	approvalConfig.ApprovalLevels["standard"] = ApprovalLevel{
		Name:                "Standard Approval",
		Description:         "Standard approval level for most requests",
		RequiredApprovers:   1,
		MaxRiskScore:        0.7,
		ApprovalTimeout:     time.Hour * 24,
	}
	
	approvalConfig.ApprovalLevels["high"] = ApprovalLevel{
		Name:                "High-Risk Approval",
		Description:         "High-risk approval level requiring multiple approvers",
		RequiredApprovers:   2,
		MaxRiskScore:        0.9,
		ApprovalTimeout:     time.Hour * 12,
	}
	
	approvalConfig.ApprovalLevels["critical"] = ApprovalLevel{
		Name:                "Critical Approval",
		Description:         "Critical approval level for highest-risk requests",
		RequiredApprovers:   3,
		MaxRiskScore:        1.0,
		ApprovalTimeout:     time.Hour * 6,
	}
	
	return &EnhancedApprovalWorkflow{
		ApprovalWorkflow:    baseWorkflow,
		config:              config,
		approvalConfig:      approvalConfig,
		pendingRequests:     make(map[string]*EnhancedApprovalRequest),
		approvalHistory:     make([]*EnhancedApprovalRequest, 0),
		approvalHandlers:    make(map[string]ApprovalHandlerFunc),
		maxPendingRequests:  100,
		maxApprovalHistory:  1000,
		defaultExpiration:   time.Hour * 24,
		dataDir:             dataDir,
	}, nil

// RequestApprovalEnhanced requests approval with enhanced capabilities
func (w *EnhancedApprovalWorkflow) RequestApprovalEnhanced(ctx context.Context, result *ProtectionResult) (bool, *ProtectionResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Create approval request
	requestID := fmt.Sprintf("request-%d", time.Now().UnixNano())
	
	// Determine approval level
	approvalLevel := w.determineApprovalLevel(result)
	
	// Create enhanced approval request
	request := &EnhancedApprovalRequest{
		ApprovalRequest: &ApprovalRequest{
			OriginalPrompt: result.OriginalPrompt,
			ProtectedPrompt: result.ProtectedPrompt,
			Detections:     result.Detections,
			RiskScore:      result.RiskScore,
			RequestID:      requestID,
			Timestamp:      time.Now(),
			Reason:         "High-risk operation requires approval",
			Metadata:       make(map[string]interface{}),
		},
		Status:              ApprovalStatusPending,
		ApprovalLevel:       approvalLevel,
		ExpirationTime:      time.Now().Add(w.approvalConfig.ApprovalLevels[approvalLevel].ApprovalTimeout),
		ApprovalHistory:     make([]*ApprovalAction, 0),
		AutoApprovalEligible: false,
	}
	
	// Check for auto-approval eligibility
	if w.approvalConfig.EnableAutoApproval {
		eligible, reason := w.checkAutoApprovalEligibility(request)
		request.AutoApprovalEligible = eligible
		request.AutoApprovalReason = reason
		
		// Auto-approve if eligible
		if eligible {
			request.Status = ApprovalStatusApproved
			request.ApproverID = "auto-approval"
			request.ApprovalTime = time.Now()
			
			// Add to approval history
			action := &ApprovalAction{
				ActionType:  "auto-approve",
				ActorID:     "system",
				Timestamp:   time.Now(),
				Reason:      reason,
			}
			request.ApprovalHistory = append(request.ApprovalHistory, action)
			
			// Add to approval history
			if w.approvalConfig.EnableApprovalHistory {
				w.approvalHistory = append(w.approvalHistory, request)
				if len(w.approvalHistory) > w.maxApprovalHistory {
					w.approvalHistory = w.approvalHistory[1:]
				}
			}
			
			// Save to disk
			w.saveApprovalToDisk(request)
			
			// Return approved result
			return true, result, nil
		}
	}
	
	// Add to pending requests
	w.pendingRequests[requestID] = request
	
	// Trim if too many pending requests
	if len(w.pendingRequests) > w.maxPendingRequests {
		// Find oldest request
		var oldestID string
		var oldestTime time.Time
		for id, req := range w.pendingRequests {
			if oldestID == "" || req.Timestamp.Before(oldestTime) {
				oldestID = id
				oldestTime = req.Timestamp
			}
		}
		
		// Remove oldest request
		if oldestID != "" {
			delete(w.pendingRequests, oldestID)
		}
	}
	
	// Save to disk
	w.saveApprovalToDisk(request)
	
	// Process with approval handlers
	approved := false
	for _, handler := range w.approvalHandlers {
		isApproved, err := handler(ctx, request)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Error processing approval request: %v\n", err)
			continue
		}
		
		if isApproved {
			approved = true
			break
		}
	}
	
	// If no handlers or all handlers rejected, use the default callback
	if !approved && w.config.ApprovalCallback != nil {
		isApproved, err := w.config.ApprovalCallback(ctx, request.ApprovalRequest)
		if err != nil {
			return false, result, err
		}
		approved = isApproved
	}
	
	// Update request status
	if approved {
		request.Status = ApprovalStatusApproved
		request.ApprovalTime = time.Now()
		
		// Add to approval history
		action := &ApprovalAction{
			ActionType:  "approve",
			ActorID:     "user",
			Timestamp:   time.Now(),
			Reason:      "User approved",
		}
		request.ApprovalHistory = append(request.ApprovalHistory, action)
	} else {
		request.Status = ApprovalStatusRejected
		
		// Add to approval history
		action := &ApprovalAction{
			ActionType:  "reject",
			ActorID:     "user",
			Timestamp:   time.Now(),
			Reason:      "User rejected",
		}
		request.ApprovalHistory = append(request.ApprovalHistory, action)
		
		// Update result
		result.ActionTaken = ActionBlocked
		result.ProtectedPrompt = "This operation requires approval but was rejected."
	}
	
	// Add to approval history
	if w.approvalConfig.EnableApprovalHistory {
		w.approvalHistory = append(w.approvalHistory, request)
		if len(w.approvalHistory) > w.maxApprovalHistory {
			w.approvalHistory = w.approvalHistory[1:]
		}
	}
	
	// Remove from pending requests
	delete(w.pendingRequests, requestID)
	
	// Save to disk
	w.saveApprovalToDisk(request)
	
	return approved, result, nil

// determineApprovalLevel determines the approval level for a request
func (w *EnhancedApprovalWorkflow) determineApprovalLevel(result *ProtectionResult) string {
	// Default to standard level
	level := w.approvalConfig.DefaultApprovalLevel
	
	// Check risk score against approval levels
	for name, approvalLevel := range w.approvalConfig.ApprovalLevels {
		if result.RiskScore <= approvalLevel.MaxRiskScore {
			level = name
			break
		}
	}
	
	// Check for specific detection types that might require higher approval
	for _, detection := range result.Detections {
		switch detection.Type {
		case DetectionTypeJailbreak, DetectionTypeSystemPrompt:
			// These are high-risk detection types, use critical level
			return "critical"
		case DetectionTypePromptInjection, DetectionTypeRoleChange:
			// These are medium-high risk detection types, use high level if not already critical
			if level != "critical" {
				level = "high"
			}
		}
	}
	
	return level

// checkAutoApprovalEligibility checks if a request is eligible for auto-approval
func (w *EnhancedApprovalWorkflow) checkAutoApprovalEligibility(request *EnhancedApprovalRequest) (bool, string) {
	// Check each auto-approval rule
	for _, rule := range w.approvalConfig.AutoApprovalRules {
		if !rule.Enabled {
			continue
		}
		
		// Check risk score
		if request.RiskScore > rule.MaxRiskScore {
			continue
		}
		
		// Check detection types
		eligible := true
		
		// Check for disallowed detection types
		for _, detection := range request.Detections {
			for _, disallowedType := range rule.DisallowedDetectionTypes {
				if detection.Type == disallowedType {
					eligible = false
					break
				}
			}
			if !eligible {
				break
			}
		}
		
		if !eligible {
			continue
		}
		
		// Check if all detections are allowed
		for _, detection := range request.Detections {
			allowed := false
			for _, allowedType := range rule.AllowedDetectionTypes {
				if detection.Type == allowedType {
					allowed = true
					break
				}
			}
			if !allowed {
				eligible = false
				break
			}
		}
		
		if !eligible {
			continue
		}
		
		// Check required metadata
		for key, value := range rule.RequiredMetadata {
			if metaValue, ok := request.Metadata[key]; !ok || metaValue != value {
				eligible = false
				break
			}
		}
		
		if !eligible {
			continue
		}
		
		// All checks passed, request is eligible for auto-approval
		return true, fmt.Sprintf("Auto-approved by rule: %s", rule.Name)
	}
	
	return false, ""
// saveApprovalToDisk saves an approval request to disk
func (w *EnhancedApprovalWorkflow) saveApprovalToDisk(request *EnhancedApprovalRequest) error {
	// Create approval directory if it doesn't exist
	approvalDir := filepath.Join(w.dataDir, "approvals")
	if err := os.MkdirAll(approvalDir, 0700); err != nil {
		return fmt.Errorf("failed to create approval directory: %w", err)
	}
	
	// Create file path
	filePath := filepath.Join(approvalDir, fmt.Sprintf("%s.json", request.RequestID))
	
	// Marshal to JSON
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal approval request: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write approval request to file: %w", err)
	}
	
	return nil

// GetPendingApprovals gets all pending approval requests
func (w *EnhancedApprovalWorkflow) GetPendingApprovals() []*EnhancedApprovalRequest {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	// Check for expired requests
	now := time.Now()
	for id, request := range w.pendingRequests {
		if now.After(request.ExpirationTime) {
			request.Status = ApprovalStatusExpired
			
			// Add to approval history
			action := &ApprovalAction{
				ActionType:  "expire",
				ActorID:     "system",
				Timestamp:   now,
				Reason:      "Request expired",
			}
			request.ApprovalHistory = append(request.ApprovalHistory, action)
			
			// Add to approval history
			if w.approvalConfig.EnableApprovalHistory {
				w.approvalHistory = append(w.approvalHistory, request)
				if len(w.approvalHistory) > w.maxApprovalHistory {
					w.approvalHistory = w.approvalHistory[1:]
				}
			}
			
			// Remove from pending requests
			delete(w.pendingRequests, id)
			
			// Save to disk
			w.saveApprovalToDisk(request)
		}
	}
	
	// Convert map to slice
	pendingApprovals := make([]*EnhancedApprovalRequest, 0, len(w.pendingRequests))
	for _, request := range w.pendingRequests {
		pendingApprovals = append(pendingApprovals, request)
	}
	
	return pendingApprovals

// GetApprovalHistory gets the approval history
func (w *EnhancedApprovalWorkflow) GetApprovalHistory() []*EnhancedApprovalRequest {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	return w.approvalHistory

// GetApprovalRequest gets an approval request by ID
func (w *EnhancedApprovalWorkflow) GetApprovalRequest(requestID string) (*EnhancedApprovalRequest, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	// Check pending requests
	if request, ok := w.pendingRequests[requestID]; ok {
		return request, nil
	}
	
	// Check approval history
	for _, request := range w.approvalHistory {
		if request.RequestID == requestID {
			return request, nil
		}
	}
	
	// Check on disk
	filePath := filepath.Join(w.dataDir, "approvals", fmt.Sprintf("%s.json", requestID))
	if _, err := os.Stat(filePath); err == nil {
		data, err := ioutil.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read approval request file: %w", err)
		}
		
		var request EnhancedApprovalRequest
		if err := json.Unmarshal(data, &request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal approval request: %w", err)
		}
		
		return &request, nil
	}
	
	return nil, fmt.Errorf("approval request not found")

// ApproveRequest approves an approval request
func (w *EnhancedApprovalWorkflow) ApproveRequest(requestID string, approverID string, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Get request
	request, ok := w.pendingRequests[requestID]
	if !ok {
		return fmt.Errorf("approval request not found")
	}
	
	// Check if already approved or rejected
	if request.Status != ApprovalStatusPending {
		return fmt.Errorf("approval request is not pending")
	}
	
	// Update request status
	request.Status = ApprovalStatusApproved
	request.ApproverID = approverID
	request.ApprovalTime = time.Now()
	
	// Add to approval history
	action := &ApprovalAction{
		ActionType:  "approve",
		ActorID:     approverID,
		Timestamp:   time.Now(),
		Reason:      reason,
	}
	request.ApprovalHistory = append(request.ApprovalHistory, action)
	
	// Add to approval history
	if w.approvalConfig.EnableApprovalHistory {
		w.approvalHistory = append(w.approvalHistory, request)
		if len(w.approvalHistory) > w.maxApprovalHistory {
			w.approvalHistory = w.approvalHistory[1:]
		}
	}
	
	// Remove from pending requests
	delete(w.pendingRequests, requestID)
	
	// Save to disk
	return w.saveApprovalToDisk(request)

// RejectRequest rejects an approval request
func (w *EnhancedApprovalWorkflow) RejectRequest(requestID string, approverID string, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Get request
	request, ok := w.pendingRequests[requestID]
	if !ok {
		return fmt.Errorf("approval request not found")
	}
	
	// Check if already approved or rejected
	if request.Status != ApprovalStatusPending {
		return fmt.Errorf("approval request is not pending")
	}
	
	// Update request status
	request.Status = ApprovalStatusRejected
	request.ApproverID = approverID
	
	// Add to approval history
	action := &ApprovalAction{
		ActionType:  "reject",
		ActorID:     approverID,
		Timestamp:   time.Now(),
		Reason:      reason,
	}
	request.ApprovalHistory = append(request.ApprovalHistory, action)
	
	// Add to approval history
	if w.approvalConfig.EnableApprovalHistory {
		w.approvalHistory = append(w.approvalHistory, request)
		if len(w.approvalHistory) > w.maxApprovalHistory {
			w.approvalHistory = w.approvalHistory[1:]
		}
	}
	
	// Remove from pending requests
	delete(w.pendingRequests, requestID)
	
	// Save to disk
	return w.saveApprovalToDisk(request)

// RegisterApprovalHandler registers a handler for approval requests
func (w *EnhancedApprovalWorkflow) RegisterApprovalHandler(name string, handler ApprovalHandlerFunc) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.approvalHandlers[name] = handler

// SetApprovalThreshold sets the threshold for requiring approval
func (w *EnhancedApprovalWorkflow) SetApprovalThreshold(level string, threshold float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.approvalConfig.ApprovalThresholds[level] = threshold

// EnableAutoApproval enables or disables auto-approval
func (w *EnhancedApprovalWorkflow) EnableAutoApproval(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.approvalConfig.EnableAutoApproval = enabled

// AddAutoApprovalRule adds an auto-approval rule
func (w *EnhancedApprovalWorkflow) AddAutoApprovalRule(rule AutoApprovalRule) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.approvalConfig.AutoApprovalRules[rule.Name] = rule

// SetDefaultApprovalLevel sets the default approval level
func (w *EnhancedApprovalWorkflow) SetDefaultApprovalLevel(level string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if _, ok := w.approvalConfig.ApprovalLevels[level]; !ok {
		return fmt.Errorf("approval level not found")
	}
	
	w.approvalConfig.DefaultApprovalLevel = level
	return nil

// AddApprovalLevel adds an approval level
func (w *EnhancedApprovalWorkflow) AddApprovalLevel(name string, level ApprovalLevel) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
}
}
}
}
}
}
}
}
}
}
}
}
}
