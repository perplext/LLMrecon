// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/perplext/LLMrecon/src/security/access"
)

// CreateIncidentRequest represents a request to create a new security incident
type CreateIncidentRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	AuditLogIDs []string               `json:"audit_log_ids,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateIncidentRequest represents a request to update a security incident
type UpdateIncidentRequest struct {
	Status     string `json:"status,omitempty"`
	AssignedTo string `json:"assigned_to,omitempty"`
}

// IncidentResponse represents a security incident response
type IncidentResponse struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	ReportedBy  string                 `json:"reported_by,omitempty"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	ResolvedAt  string                 `json:"resolved_at,omitempty"`
	AuditLogIDs []string               `json:"audit_log_ids,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// handleListIncidents handles listing security incidents with filtering
func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentView) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Create filter
	filter := &access.IncidentFilter{
		Severity:   query.Get("severity"),
		Status:     query.Get("status"),
		AssigneeID: query.Get("assigned_to"),
	}

	// Parse time range
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		_, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD")
			return
		}
		filter.ReportedAfter = startDateStr
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter.Offset = (page - 1) * limit
	filter.Limit = limit

	// Get security incidents
	incidents, err := s.accessManager.ListIncidents(r.Context(), filter)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list security incidents")
		return
	}

	// Use len for total count
	totalCount := int64(len(incidents))

	// Convert incidents to response format
	var incidentResponses []IncidentResponse
	for _, incident := range incidents {
		incidentResponses = append(incidentResponses, convertIncidentToResponse(incident))
	}

	// Create response
	resp := struct {
		Incidents  []IncidentResponse `json:"incidents"`
		TotalCount int64              `json:"total_count"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
		TotalPages int                `json:"total_pages"`
	}{
		Incidents:  incidentResponses,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: int((totalCount + int64(limit) - 1) / int64(limit)),
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incidents retrieved successfully", resp)
}

// handleCreateIncident handles creating a new security incident
func (s *Server) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentCreate) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse request
	var req CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Title == "" || req.Description == "" || req.Severity == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Title, description, and severity are required")
		return
	}

	// Validate severity
	validSeverities := []string{
		access.SeverityCritical,
		access.SeverityHigh,
		access.SeverityMedium,
		access.SeverityLow,
		access.SeverityInfo,
	}

	validSeverity := false
	for _, severity := range validSeverities {
		if req.Severity == severity {
			validSeverity = true
			break
		}
	}

	if !validSeverity {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid severity value")
		return
	}

	// Create incident using AccessControlManager
	incident, err := s.accessManager.CreateIncident(
		r.Context(),
		req.Title,
		req.Description,
		access.AuditSeverity(req.Severity),
		currentUser.ID,
		req.AuditLogIDs,
		req.Metadata,
	)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create security incident")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusCreated, "Security incident created successfully", convertIncidentToResponse(incident))
}

// handleGetIncident handles retrieving a security incident
func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentView) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get incident ID from URL
	vars := mux.Vars(r)
	incidentID := vars["id"]
	if incidentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Incident ID is required")
		return
	}

	// Get incident
	incident, err := s.accessManager.GetIncident(r.Context(), incidentID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Security incident not found")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incident retrieved successfully", convertIncidentToResponse(incident))
}

// handleUpdateIncident handles updating a security incident
func (s *Server) handleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentUpdate) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get incident ID from URL
	vars := mux.Vars(r)
	incidentID := vars["id"]
	if incidentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Incident ID is required")
		return
	}

	// Parse request
	var req UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate status if provided
	if req.Status != "" {
		validStatuses := []string{
			string(access.IncidentStatusNew),
			string(access.IncidentStatusInProgress),
			string(access.IncidentStatusResolved),
			string(access.IncidentStatusClosed),
			string(access.IncidentStatusDuplicate),
		}

		validStatus := false
		for _, status := range validStatuses {
			if req.Status == status {
				validStatus = true
				break
			}
		}

		if !validStatus {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid status value")
			return
		}
	}

	// Update incident status
	if err := s.accessManager.UpdateIncidentStatus(
		r.Context(),
		incidentID,
		access.IncidentStatus(req.Status),
		req.AssignedTo,
		currentUser.ID,
	); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update security incident")
		return
	}

	// Get updated incident to return
	incident, err := s.accessManager.GetIncident(r.Context(), incidentID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve updated incident")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incident updated successfully", convertIncidentToResponse(incident))
}

// handleDeleteIncident handles deleting a security incident
func (s *Server) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentDelete) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get incident ID from URL
	vars := mux.Vars(r)
	incidentID := vars["id"]
	if incidentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Incident ID is required")
		return
	}

	// Close the incident instead of deleting (no delete method available)
	if err := s.accessManager.UpdateIncidentStatus(
		r.Context(),
		incidentID,
		access.IncidentStatusClosed,
		"",
		currentUser.ID,
	); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to close security incident")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incident closed successfully", nil)
}

// convertIncidentToResponse converts a security incident to a response format
func convertIncidentToResponse(incident *access.SecurityIncident) IncidentResponse {
	var resolvedAt string
	if !incident.ResolvedAt.IsZero() {
		resolvedAt = incident.ResolvedAt.Format(time.RFC3339)
	}

	return IncidentResponse{
		ID:          incident.ID,
		Title:       incident.Title,
		Description: incident.Description,
		Severity:    string(incident.Severity),
		Status:      string(incident.Status),
		ReportedBy:  incident.ReportedBy,
		AssignedTo:  incident.AssignedTo,
		CreatedAt:   incident.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   incident.UpdatedAt.Format(time.RFC3339),
		ResolvedAt:  resolvedAt,
		AuditLogIDs: incident.AuditLogIDs,
		Metadata:    incident.Metadata,
	}
}
