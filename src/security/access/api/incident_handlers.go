// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/perplext/LLMrecon/src/security/access"
)

// CreateIncidentRequest represents a request to create a new security incident
type CreateIncidentRequest struct {
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Severity          string                 `json:"severity"`
	AffectedResources []string               `json:"affected_resources,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`

// UpdateIncidentRequest represents a request to update a security incident
type UpdateIncidentRequest struct {
	Title             string                 `json:"title,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Severity          string                 `json:"severity,omitempty"`
	Status            string                 `json:"status,omitempty"`
	AssignedTo        string                 `json:"assigned_to,omitempty"`
	ResolutionNotes   string                 `json:"resolution_notes,omitempty"`
	AffectedResources []string               `json:"affected_resources,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// IncidentResponse represents a security incident response
type IncidentResponse struct {
	ID                string                 `json:"id"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Severity          string                 `json:"severity"`
	Status            string                 `json:"status"`
	ReportedBy        string                 `json:"reported_by,omitempty"`
	AssignedTo        string                 `json:"assigned_to,omitempty"`
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
	ResolvedAt        string                 `json:"resolved_at,omitempty"`
	ResolutionNotes   string                 `json:"resolution_notes,omitempty"`
	AffectedResources []string               `json:"affected_resources,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`

// handleListIncidents handles listing security incidents with filtering
func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentView) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	
	// Create filter
	filter := &access.IncidentFilter{
		Severity:   query.Get("severity"),
		Status:     query.Get("status"),
		ReportedBy: query.Get("reported_by"),
		AssignedTo: query.Get("assigned_to"),
	}
	
	// Parse time range
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid start_date format, expected YYYY-MM-DD")
			return
		}
		filter.StartDate = &startDate
	}
	
	if endDateStr := query.Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid end_date format, expected YYYY-MM-DD")
			return
		}
		// Set to end of day
		endDate = endDate.Add(24*time.Hour - time.Second)
		filter.EndDate = &endDate
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
	securityManager := s.accessManager.GetSecurityManager()
	incidents, err := securityManager.ListIncidents(r.Context(), filter)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list security incidents")
		return
	}

	// Get total count
	totalCount, err := securityManager.CountIncidents(r.Context(), filter)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to count security incidents")
		return
	}

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

// handleCreateIncident handles creating a new security incident
func (s *Server) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentCreate) {
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

	// Create incident
	incident := &access.SecurityIncident{
		Title:             req.Title,
		Description:       req.Description,
		Severity:          req.Severity,
		Status:            access.StatusOpen,
		ReportedBy:        currentUser.ID,
		AffectedResources: req.AffectedResources,
		Tags:              req.Tags,
		Metadata:          req.Metadata,
	}

	securityManager := s.accessManager.GetSecurityManager()
	if err := securityManager.CreateIncident(r.Context(), incident); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create security incident")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusCreated, "Security incident created successfully", convertIncidentToResponse(incident))

// handleGetIncident handles retrieving a security incident
func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentView) {
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
	securityManager := s.accessManager.GetSecurityManager()
	incident, err := securityManager.GetIncident(r.Context(), incidentID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Security incident not found")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incident retrieved successfully", convertIncidentToResponse(incident))

// handleUpdateIncident handles updating a security incident
func (s *Server) handleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentUpdate) {
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

	// Get incident
	securityManager := s.accessManager.GetSecurityManager()
	incident, err := securityManager.GetIncident(r.Context(), incidentID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Security incident not found")
		return
	}

	// Update incident fields
	if req.Title != "" {
		incident.Title = req.Title
	}
	
	if req.Description != "" {
		incident.Description = req.Description
	}
	
	if req.Severity != "" {
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
		
		incident.Severity = req.Severity
	}
	
	if req.Status != "" {
		// Validate status
		validStatuses := []string{
			access.StatusOpen,
			access.StatusInProgress,
			access.StatusResolved,
			access.StatusClosed,
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
		
		// If changing to resolved or closed, set resolved time
		if (req.Status == access.StatusResolved || req.Status == access.StatusClosed) && 
		   (incident.Status != access.StatusResolved && incident.Status != access.StatusClosed) {
			incident.ResolvedAt = time.Now()
		}
		
		incident.Status = req.Status
	}
	
	if req.AssignedTo != "" {
		// Validate that the assigned user exists
		userManager := s.accessManager.GetUserManager()
		_, err := userManager.GetUserByID(r.Context(), req.AssignedTo)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Assigned user does not exist")
			return
		}
		
		incident.AssignedTo = req.AssignedTo
	}
	
	if req.ResolutionNotes != "" {
		incident.ResolutionNotes = req.ResolutionNotes
	}
	
	if len(req.AffectedResources) > 0 {
		incident.AffectedResources = req.AffectedResources
	}
	
	if len(req.Tags) > 0 {
		incident.Tags = req.Tags
	}
	
	if req.Metadata != nil {
		incident.Metadata = req.Metadata
	}
	
	// Update timestamp
	incident.UpdatedAt = time.Now()

	// Update incident
	if err := securityManager.UpdateIncident(r.Context(), incident); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update security incident")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incident updated successfully", convertIncidentToResponse(incident))

// handleDeleteIncident handles deleting a security incident
func (s *Server) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionSecurityIncidentDelete) {
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

	// Delete incident
	securityManager := s.accessManager.GetSecurityManager()
	if err := securityManager.DeleteIncident(r.Context(), incidentID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorResponse(w, http.StatusNotFound, "Security incident not found")
		} else {
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete security incident")
		}
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Security incident deleted successfully", nil)

// convertIncidentToResponse converts a security incident to a response format
func convertIncidentToResponse(incident *access.SecurityIncident) IncidentResponse {
	var resolvedAt string
	if !incident.ResolvedAt.IsZero() {
		resolvedAt = incident.ResolvedAt.Format(time.RFC3339)
	}

	return IncidentResponse{
		ID:                incident.ID,
		Title:             incident.Title,
		Description:       incident.Description,
		Severity:          incident.Severity,
		Status:            incident.Status,
		ReportedBy:        incident.ReportedBy,
		AssignedTo:        incident.AssignedTo,
		CreatedAt:         incident.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         incident.UpdatedAt.Format(time.RFC3339),
		ResolvedAt:        resolvedAt,
		ResolutionNotes:   incident.ResolutionNotes,
		AffectedResources: incident.AffectedResources,
		Tags:              incident.Tags,
		Metadata:          incident.Metadata,
	}
}
}
}
}
}
