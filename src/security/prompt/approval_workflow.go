// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ApprovalWorkflow manages approval workflows for high-risk operations
type ApprovalWorkflow struct {
	config           *ProtectionConfig
	workflowConfig   *ApprovalWorkflowConfig
	pendingRequests  map[string]*ApprovalRequest
	approvedRequests map[string]bool
	deniedRequests   map[string]bool
	mu               sync.RWMutex
}

// NewApprovalWorkflow creates a new approval workflow
func NewApprovalWorkflow(config *ProtectionConfig) *ApprovalWorkflow {
	// Create default workflow config if not specified
	workflowConfig := &ApprovalWorkflowConfig{
		ApprovalThreshold: config.ApprovalThreshold,
		ApprovalTimeout:   time.Minute * 5, // 5 minute default timeout
		EnableAutoApproval: false,
		AutoApprovalRules: make(map[string]interface{}),
	}

	return &ApprovalWorkflow{
		config:           config,
		workflowConfig:   workflowConfig,
		pendingRequests:  make(map[string]*ApprovalRequest),
		approvedRequests: make(map[string]bool),
		deniedRequests:   make(map[string]bool),
	}
}

// RequestApproval requests approval for a high-risk operation
func (w *ApprovalWorkflow) RequestApproval(ctx context.Context, result *ProtectionResult) (bool, *ProtectionResult, error) {
	// Create a new approval request
	requestID := uuid.New().String()
	
	request := &ApprovalRequest{
		OriginalPrompt: result.OriginalPrompt,
		ProtectedPrompt: result.ProtectedPrompt,
		Detections:     result.Detections,
		RiskScore:      result.RiskScore,
		RequestID:      requestID,
		Timestamp:      time.Now(),
		Reason:         generateApprovalReason(result),
	}
	
	// Check if auto-approval is enabled and applicable
	if w.workflowConfig.EnableAutoApproval {
		autoApproved, autoResult := w.checkAutoApproval(ctx, request)
		if autoApproved {
			return true, autoResult, nil
		}
	}
	
	// Store the request
	w.mu.Lock()
	w.pendingRequests[requestID] = request
	w.mu.Unlock()
	
	// Call the approval callback if provided
	if w.config.ApprovalCallback != nil {
		approved, err := w.config.ApprovalCallback(ctx, request)
		if err != nil {
			return false, result, err
		}
		
		// Store the result
		w.mu.Lock()
		delete(w.pendingRequests, requestID)
		if approved {
			w.approvedRequests[requestID] = true
		} else {
			w.deniedRequests[requestID] = true
		}
		w.mu.Unlock()
		
		return approved, result, nil
	}
	
	// If no callback is provided, use a blocking approach with timeout
	approvalChan := make(chan bool, 1)
	errorChan := make(chan error, 1)
	
	// Start a goroutine to wait for approval
	go func() {
		// This would typically be implemented with a UI or API endpoint
		// For now, we'll simulate approval based on risk score
		time.Sleep(time.Second * 2) // Simulate processing time
		
		// Auto-deny very high risk requests
		if result.RiskScore >= 0.95 {
			approvalChan <- false
			return
		}
		
		// Auto-approve lower risk requests
		if result.RiskScore < 0.85 {
			approvalChan <- true
			return
		}
		
		// For medium risk, flip a coin (simulating human decision)
		approvalChan <- (time.Now().UnixNano()%2 == 0)
	}()
	
	// Wait for approval or timeout
	select {
	case approved := <-approvalChan:
		// Store the result
		w.mu.Lock()
		delete(w.pendingRequests, requestID)
		if approved {
			w.approvedRequests[requestID] = true
		} else {
			w.deniedRequests[requestID] = true
		}
		w.mu.Unlock()
		
		return approved, result, nil
	case err := <-errorChan:
		return false, result, err
	case <-time.After(w.workflowConfig.ApprovalTimeout):
		// Handle timeout
		w.mu.Lock()
		delete(w.pendingRequests, requestID)
		w.deniedRequests[requestID] = true
		w.mu.Unlock()
		
		return false, result, fmt.Errorf("approval request timed out after %v", w.workflowConfig.ApprovalTimeout)
	case <-ctx.Done():
		// Handle context cancellation
		w.mu.Lock()
		delete(w.pendingRequests, requestID)
		w.deniedRequests[requestID] = true
		w.mu.Unlock()
		
		return false, result, ctx.Err()
	}
}

// checkAutoApproval checks if a request can be auto-approved based on rules
func (w *ApprovalWorkflow) checkAutoApproval(ctx context.Context, request *ApprovalRequest) (bool, *ProtectionResult) {
	result := &ProtectionResult{
		OriginalPrompt:   request.OriginalPrompt,
		ProtectedPrompt:  request.ProtectedPrompt,
		Detections:       request.Detections,
		RiskScore:        request.RiskScore,
		ActionTaken:      ActionNone,
		Timestamp:        time.Now(),
	}
	
	// Check if the risk score is below a certain threshold
	if request.RiskScore < 0.6 {
		// Low risk, auto-approve
		result.Detections = append(result.Detections, &Detection{
			Type:        DetectionTypeUnusualPattern,
			Confidence:  0.5,
			Description: "Request auto-approved due to low risk score",
			Remediation: "Monitor for potential false negatives",
			Metadata: map[string]interface{}{
				"auto_approval": true,
				"reason":        "low_risk_score",
			},
		})
		
		return true, result
	}
	
	// Check if all detections are of low severity
	allLowSeverity := true
	for _, detection := range request.Detections {
		if detection.Confidence > 0.7 {
			allLowSeverity = false
			break
		}
	}
	
	if allLowSeverity {
		// All low severity detections, auto-approve
		result.Detections = append(result.Detections, &Detection{
			Type:        DetectionTypeUnusualPattern,
			Confidence:  0.5,
			Description: "Request auto-approved due to low severity detections",
			Remediation: "Monitor for potential false negatives",
			Metadata: map[string]interface{}{
				"auto_approval": true,
				"reason":        "low_severity_detections",
			},
		})
		
		return true, result
	}
	
	// Apply custom auto-approval rules
	// This would be implemented based on specific requirements
	
	// No auto-approval rules matched
	return false, result
}

// GetPendingRequests gets all pending approval requests
func (w *ApprovalWorkflow) GetPendingRequests() []*ApprovalRequest {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	requests := make([]*ApprovalRequest, 0, len(w.pendingRequests))
	for _, request := range w.pendingRequests {
		requests = append(requests, request)
	}
	
	return requests
}

// GetRequestStatus gets the status of an approval request
func (w *ApprovalWorkflow) GetRequestStatus(requestID string) (string, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if _, ok := w.pendingRequests[requestID]; ok {
		return "pending", true
	}
	
	if _, ok := w.approvedRequests[requestID]; ok {
		return "approved", true
	}
	
	if _, ok := w.deniedRequests[requestID]; ok {
		return "denied", true
	}
	
	return "", false
}

// ApproveRequest approves a pending request
func (w *ApprovalWorkflow) ApproveRequest(requestID string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	_, ok := w.pendingRequests[requestID]
	if !ok {
		return false
	}
	
	delete(w.pendingRequests, requestID)
	w.approvedRequests[requestID] = true
	
	// Log the approval
	// This would be implemented with a proper logging system
	
	return true
}

// DenyRequest denies a pending request
func (w *ApprovalWorkflow) DenyRequest(requestID string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	_, ok := w.pendingRequests[requestID]
	if !ok {
		return false
	}
	
	delete(w.pendingRequests, requestID)
	w.deniedRequests[requestID] = true
	
	// Log the denial
	// This would be implemented with a proper logging system
	
	return true
}

// generateApprovalReason generates a reason for the approval request
func generateApprovalReason(result *ProtectionResult) string {
	if len(result.Detections) == 0 {
		return fmt.Sprintf("High risk score (%.2f) without specific detections", result.RiskScore)
	}
	
	// Get the top 3 detections by confidence
	topDetections := getTopDetections(result.Detections, 3)
	
	reason := fmt.Sprintf("High risk score (%.2f) with the following detections:\n", result.RiskScore)
	for i, detection := range topDetections {
		reason += fmt.Sprintf("%d. %s (Confidence: %.2f)\n", i+1, detection.Description, detection.Confidence)
	}
	
	return reason
}

// getTopDetections gets the top N detections by confidence
func getTopDetections(detections []*Detection, n int) []*Detection {
	// Sort detections by confidence (descending)
	sortedDetections := make([]*Detection, len(detections))
	copy(sortedDetections, detections)
	
	// Simple bubble sort
	for i := 0; i < len(sortedDetections); i++ {
		for j := i + 1; j < len(sortedDetections); j++ {
			if sortedDetections[i].Confidence < sortedDetections[j].Confidence {
				sortedDetections[i], sortedDetections[j] = sortedDetections[j], sortedDetections[i]
			}
		}
	}
	
	// Get the top N
	if len(sortedDetections) <= n {
		return sortedDetections
	}
	
	return sortedDetections[:n]
}
