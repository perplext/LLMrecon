// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/perplext/LLMrecon/src/security/access"
)

// AuditLogResponse represents an audit log entry in the response
type AuditLogResponse struct {
	ID         string                 `json:"id"`
	Timestamp  string                 `json:"timestamp"`
	UserID     string                 `json:"user_id,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource,omitempty"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Severity   string                 `json:"severity"`
	Status     string                 `json:"status"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// handleListAuditLogs handles listing audit logs with filtering
func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionAuditView) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	
	// Create filter
	filter := &access.AuditEventFilter{
		UserID:     query.Get("user_id"),
		Username:   query.Get("username"),
		Action:     query.Get("action"),
		Resource:   query.Get("resource"),
		ResourceID: query.Get("resource_id"),
		Severity:   query.Get("severity"),
		Status:     query.Get("status"),
		IPAddress:  query.Get("ip_address"),
	}
	
	// Parse time range
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid start_time format, expected RFC3339")
			return
		}
		filter.StartTime = &startTime
	}
	
	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid end_time format, expected RFC3339")
			return
		}
		filter.EndTime = &endTime
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

	// Get audit logs
	auditLogger := s.accessManager.GetAuditLogger()
	events, err := auditLogger.QueryEvents(r.Context(), filter)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to query audit logs")
		return
	}

	// Get total count
	totalCount, err := auditLogger.CountEvents(r.Context(), filter)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to count audit logs")
		return
	}

	// Convert events to response format
	var eventResponses []AuditLogResponse
	for _, event := range events {
		eventResponses = append(eventResponses, convertAuditEventToResponse(event))
	}

	// Create response
	resp := struct {
		AuditLogs  []AuditLogResponse `json:"audit_logs"`
		TotalCount int64              `json:"total_count"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
		TotalPages int                `json:"total_pages"`
	}{
		AuditLogs:  eventResponses,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: int((totalCount + int64(limit) - 1) / int64(limit)),
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Audit logs retrieved successfully", resp)
}

// handleGetAuditLog handles retrieving a specific audit log entry
func (s *Server) handleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionAuditView) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get audit log ID from URL
	vars := mux.Vars(r)
	auditLogID := vars["id"]
	if auditLogID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Audit log ID is required")
		return
	}

	// Get audit log
	auditLogger := s.accessManager.GetAuditLogger()
	event, err := auditLogger.GetEvent(r.Context(), auditLogID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Audit log not found")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Audit log retrieved successfully", convertAuditEventToResponse(event))
}

// handleExportAuditLogs handles exporting audit logs to a file
func (s *Server) handleExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionAuditExport) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	
	// Create filter
	filter := &access.AuditEventFilter{
		UserID:     query.Get("user_id"),
		Username:   query.Get("username"),
		Action:     query.Get("action"),
		Resource:   query.Get("resource"),
		ResourceID: query.Get("resource_id"),
		Severity:   query.Get("severity"),
		Status:     query.Get("status"),
		IPAddress:  query.Get("ip_address"),
	}
	
	// Parse time range
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid start_time format, expected RFC3339")
			return
		}
		filter.StartTime = &startTime
	}
	
	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid end_time format, expected RFC3339")
			return
		}
		filter.EndTime = &endTime
	}
	
	// Get export format
	format := query.Get("format")
	if format == "" {
		format = "csv" // Default format
	}
	
	if format != "csv" && format != "json" {
		WriteErrorResponse(w, http.StatusBadRequest, "Unsupported export format, supported formats: csv, json")
		return
	}

	// Get audit logs (no pagination for export)
	auditLogger := s.accessManager.GetAuditLogger()
	events, err := auditLogger.QueryEvents(r.Context(), filter)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to query audit logs")
		return
	}

	// Set response headers
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("audit_logs_%s.%s", timestamp, format)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	
	// Export based on format
	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		exportAuditLogsToCSV(w, events)
	case "json":
		w.Header().Set("Content-Type", "application/json")
		exportAuditLogsToJSON(w, events)
	}
}

// exportAuditLogsToCSV exports audit logs to CSV format
func exportAuditLogsToCSV(w http.ResponseWriter, events []*access.AuditEvent) {
	// Create CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "Timestamp", "UserID", "Username", "Action", "Resource", "ResourceID",
		"Severity", "Status", "IPAddress", "UserAgent", "Details",
	}
	writer.Write(header)

	// Write data
	for _, event := range events {
		// Convert details to JSON string
		detailsJSON, _ := json.Marshal(event.Details)
		
		row := []string{
			event.ID,
			event.Timestamp.Format(time.RFC3339),
			event.UserID,
			event.Username,
			event.Action,
			event.Resource,
			event.ResourceID,
			event.Severity,
			event.Status,
			event.IPAddress,
			event.UserAgent,
			string(detailsJSON),
		}
		writer.Write(row)
	}
}

// exportAuditLogsToJSON exports audit logs to JSON format
func exportAuditLogsToJSON(w http.ResponseWriter, events []*access.AuditEvent) {
	// Convert events to response format
	var eventResponses []AuditLogResponse
	for _, event := range events {
		eventResponses = append(eventResponses, convertAuditEventToResponse(event))
	}

	// Write JSON
	json.NewEncoder(w).Encode(eventResponses)
}

// convertAuditEventToResponse converts an audit event to a response format
func convertAuditEventToResponse(event *access.AuditEvent) AuditLogResponse {
	return AuditLogResponse{
		ID:         event.ID,
		Timestamp:  event.Timestamp.Format(time.RFC3339),
		UserID:     event.UserID,
		Username:   event.Username,
		Action:     event.Action,
		Resource:   event.Resource,
		ResourceID: event.ResourceID,
		Severity:   event.Severity,
		Status:     event.Status,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		Details:    event.Details,
		Metadata:   event.Metadata,
	}
}
