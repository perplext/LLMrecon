package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// DeploymentInfo represents information about a deployment
type DeploymentInfo struct {
	ID             string    `json:"id"`
	Organization   string    `json:"organization"`
	Version        string    `json:"version"`
	Environment    string    `json:"environment"`
	DeployedAt     time.Time `json:"deployed_at"`
	Status         string    `json:"status"`
	Metrics        Metrics   `json:"metrics"`
	Issues         []Issue   `json:"issues"`
	LastHealthCheck time.Time `json:"last_health_check"`
}

// Metrics represents deployment performance metrics
type Metrics struct {
	Uptime           time.Duration `json:"uptime"`
	TotalRequests    int64        `json:"total_requests"`
	SuccessRate      float64      `json:"success_rate"`
	AvgResponseTime  float64      `json:"avg_response_time"`
	PeakConcurrent   int          `json:"peak_concurrent"`
	ErrorCount       int64        `json:"error_count"`
	LastUpdated      time.Time    `json:"last_updated"`
}

// Issue represents a deployment issue
type Issue struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"` // P0, P1, P2, P3
	Type        string    `json:"type"`
	Description string    `json:"description"`
	ReportedAt  time.Time `json:"reported_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Status      string    `json:"status"` // open, investigating, resolved
}

// DeploymentTracker tracks all v0.2.0 deployments
type DeploymentTracker struct {
	deployments map[string]*DeploymentInfo
	mu          sync.RWMutex
}

// NewDeploymentTracker creates a new deployment tracker
func NewDeploymentTracker() *DeploymentTracker {
	return &DeploymentTracker{
		deployments: make(map[string]*DeploymentInfo),
	}
}

// RegisterDeployment registers a new deployment
func (dt *DeploymentTracker) RegisterDeployment(info DeploymentInfo) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	
	if info.ID == "" {
		info.ID = fmt.Sprintf("%s-%d", info.Organization, time.Now().Unix())
	}
	
	info.Status = "active"
	info.DeployedAt = time.Now()
	info.LastHealthCheck = time.Now()
	
	dt.deployments[info.ID] = &info
	
	log.Printf("Registered deployment: %s for %s", info.ID, info.Organization)
	return nil
}

// UpdateMetrics updates deployment metrics
func (dt *DeploymentTracker) UpdateMetrics(deploymentID string, metrics Metrics) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	
	deployment, exists := dt.deployments[deploymentID]
	if !exists {
		return fmt.Errorf("deployment %s not found", deploymentID)
	}
	
	metrics.LastUpdated = time.Now()
	deployment.Metrics = metrics
	deployment.LastHealthCheck = time.Now()
	
	// Check for issues based on metrics
	dt.checkMetricsForIssues(deployment)
	
	return nil
}

// ReportIssue reports an issue for a deployment
func (dt *DeploymentTracker) ReportIssue(deploymentID string, issue Issue) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	
	deployment, exists := dt.deployments[deploymentID]
	if !exists {
		return fmt.Errorf("deployment %s not found", deploymentID)
	}
	
	issue.ID = fmt.Sprintf("ISSUE-%d", time.Now().Unix())
	issue.ReportedAt = time.Now()
	issue.Status = "open"
	
	deployment.Issues = append(deployment.Issues, issue)
	
	log.Printf("Issue reported for %s: %s (%s)", deploymentID, issue.Description, issue.Severity)
	
	// Send alert for P0/P1 issues
	if issue.Severity == "P0" || issue.Severity == "P1" {
		dt.sendAlert(deployment, issue)
	}
	
	return nil
}

// GetDeploymentStatus returns the status of all deployments
func (dt *DeploymentTracker) GetDeploymentStatus() map[string]*DeploymentInfo {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	status := make(map[string]*DeploymentInfo)
	for k, v := range dt.deployments {
		status[k] = v
	}
	
	return status
}

// GetDeploymentHealth returns health summary
func (dt *DeploymentTracker) GetDeploymentHealth() map[string]interface{} {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	
	totalDeployments := len(dt.deployments)
	healthyDeployments := 0
	totalRequests := int64(0)
	totalErrors := int64(0)
	openIssues := 0
	
	for _, deployment := range dt.deployments {
		if deployment.Status == "active" && deployment.Metrics.SuccessRate >= 0.95 {
			healthyDeployments++
		}
		
		totalRequests += deployment.Metrics.TotalRequests
		totalErrors += deployment.Metrics.ErrorCount
		
		for _, issue := range deployment.Issues {
			if issue.Status != "resolved" {
				openIssues++
			}
		}
	}
	
	overallSuccessRate := float64(0)
	if totalRequests > 0 {
		overallSuccessRate = float64(totalRequests-totalErrors) / float64(totalRequests)
	}
	
	return map[string]interface{}{
		"total_deployments":   totalDeployments,
		"healthy_deployments": healthyDeployments,
		"health_percentage":   float64(healthyDeployments) / float64(totalDeployments) * 100,
		"total_requests":      totalRequests,
		"overall_success_rate": overallSuccessRate,
		"open_issues":         openIssues,
		"timestamp":           time.Now(),
	}
}

// checkMetricsForIssues automatically detects issues based on metrics
func (dt *DeploymentTracker) checkMetricsForIssues(deployment *DeploymentInfo) {
	// Check for low success rate
	if deployment.Metrics.SuccessRate < 0.95 {
		issue := Issue{
			Severity:    "P1",
			Type:        "performance",
			Description: fmt.Sprintf("Success rate below threshold: %.2f%%", deployment.Metrics.SuccessRate*100),
		}
		dt.addIssueIfNew(deployment, issue)
	}
	
	// Check for high response time
	if deployment.Metrics.AvgResponseTime > 2.0 {
		issue := Issue{
			Severity:    "P2",
			Type:        "performance",
			Description: fmt.Sprintf("Response time above threshold: %.2fs", deployment.Metrics.AvgResponseTime),
		}
		dt.addIssueIfNew(deployment, issue)
	}
}

// addIssueIfNew adds an issue if it doesn't already exist
func (dt *DeploymentTracker) addIssueIfNew(deployment *DeploymentInfo, newIssue Issue) {
	// Check if similar issue already exists
	for _, existingIssue := range deployment.Issues {
		if existingIssue.Type == newIssue.Type && existingIssue.Status != "resolved" {
			return // Similar issue already tracked
		}
	}
	
	newIssue.ID = fmt.Sprintf("AUTO-%d", time.Now().Unix())
	newIssue.ReportedAt = time.Now()
	newIssue.Status = "open"
	
	deployment.Issues = append(deployment.Issues, newIssue)
}

// sendAlert sends alerts for critical issues
func (dt *DeploymentTracker) sendAlert(deployment *DeploymentInfo, issue Issue) {
	// In a real implementation, this would send to Slack, PagerDuty, etc.
	log.Printf("ALERT: %s deployment has %s issue: %s", 
		deployment.Organization, issue.Severity, issue.Description)
}

// HTTP Handlers

func (dt *DeploymentTracker) handleRegister(w http.ResponseWriter, r *http.Request) {
	var info DeploymentInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := dt.RegisterDeployment(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": info.ID})
}

func (dt *DeploymentTracker) handleUpdateMetrics(w http.ResponseWriter, r *http.Request) {
	deploymentID := r.URL.Query().Get("deployment_id")
	if deploymentID == "" {
		http.Error(w, "deployment_id required", http.StatusBadRequest)
		return
	}
	
	var metrics Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := dt.UpdateMetrics(deploymentID, metrics); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func (dt *DeploymentTracker) handleReportIssue(w http.ResponseWriter, r *http.Request) {
	deploymentID := r.URL.Query().Get("deployment_id")
	if deploymentID == "" {
		http.Error(w, "deployment_id required", http.StatusBadRequest)
		return
	}
	
	var issue Issue
	if err := json.NewDecoder(r.Body).Decode(&issue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := dt.ReportIssue(deploymentID, issue); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
}

func (dt *DeploymentTracker) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dt.GetDeploymentStatus())
}

func (dt *DeploymentTracker) handleGetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dt.GetDeploymentHealth())
}

func main() {
	tracker := NewDeploymentTracker()
	
	// Set up HTTP routes
	http.HandleFunc("/api/v1/deployments/register", tracker.handleRegister)
	http.HandleFunc("/api/v1/deployments/metrics", tracker.handleUpdateMetrics)
	http.HandleFunc("/api/v1/deployments/issues", tracker.handleReportIssue)
	http.HandleFunc("/api/v1/deployments/status", tracker.handleGetStatus)
	http.HandleFunc("/api/v1/deployments/health", tracker.handleGetHealth)
	
	log.Println("Deployment tracker starting on :8091")
	log.Fatal(http.ListenAndServe(":8091", nil))
}