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
	access "github.com/perplext/LLMrecon/src/security/access"
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
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionAuditView)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Build filter map for the AuditLogger interface
	filter := make(map[string]interface{})
	if v := query.Get("user_id"); v != "" {
		filter["user_id"] = v
	}
	if v := query.Get("username"); v != "" {
		filter["username"] = v
	}
	if v := query.Get("action"); v != "" {
		filter["action"] = v
	}
	if v := query.Get("resource"); v != "" {
		filter["resource"] = v
	}
	if v := query.Get("resource_id"); v != "" {
		filter["resource_id"] = v
	}
	if v := query.Get("severity"); v != "" {
		filter["severity"] = v
	}
	if v := query.Get("status"); v != "" {
		filter["status"] = v
	}
	if v := query.Get("ip_address"); v != "" {
		filter["ip_address"] = v
	}

	// Parse time range
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid start_time format, expected RFC3339")
			return
		}
		filter["start_time"] = startTime
	}

	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid end_time format, expected RFC3339")
			return
		}
		filter["end_time"] = endTime
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

	offset := (page - 1) * limit

	// Get audit logs
	auditLogger := s.accessManager.GetAuditLogger()
	logs, totalCount, err := auditLogger.GetAuditLogs(r.Context(), filter, offset, limit)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to query audit logs")
		return
	}

	// Convert logs to response format
	var logResponses []AuditLogResponse
	for _, log := range logs {
		logResponses = append(logResponses, convertAuditLogToResponse(log))
	}

	// Create response
	resp := struct {
		AuditLogs  []AuditLogResponse `json:"audit_logs"`
		TotalCount int                `json:"total_count"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
		TotalPages int                `json:"total_pages"`
	}{
		AuditLogs:  logResponses,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: (totalCount + limit - 1) / limit,
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
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionAuditView)
	if err != nil || !hasPermission {
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
	log, err := auditLogger.GetAuditLogByID(r.Context(), auditLogID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Audit log not found")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Audit log retrieved successfully", convertAuditLogToResponse(log))
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
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionAuditExport)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Build filter map for the AuditLogger interface
	filter := make(map[string]interface{})
	if v := query.Get("user_id"); v != "" {
		filter["user_id"] = v
	}
	if v := query.Get("username"); v != "" {
		filter["username"] = v
	}
	if v := query.Get("action"); v != "" {
		filter["action"] = v
	}
	if v := query.Get("resource"); v != "" {
		filter["resource"] = v
	}
	if v := query.Get("resource_id"); v != "" {
		filter["resource_id"] = v
	}
	if v := query.Get("severity"); v != "" {
		filter["severity"] = v
	}
	if v := query.Get("status"); v != "" {
		filter["status"] = v
	}
	if v := query.Get("ip_address"); v != "" {
		filter["ip_address"] = v
	}

	// Parse time range
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid start_time format, expected RFC3339")
			return
		}
		filter["start_time"] = startTime
	}

	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid end_time format, expected RFC3339")
			return
		}
		filter["end_time"] = endTime
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

	// Get audit logs (large limit for export)
	auditLogger := s.accessManager.GetAuditLogger()
	logs, _, err := auditLogger.GetAuditLogs(r.Context(), filter, 0, 10000)
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
		exportAuditLogsToCSV(w, logs)
	case "json":
		w.Header().Set("Content-Type", "application/json")
		exportAuditLogsToJSON(w, logs)
	}
}

// exportAuditLogsToCSV exports audit logs to CSV format
func exportAuditLogsToCSV(w http.ResponseWriter, logs []*access.AuditLog) {
	// Create CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "Timestamp", "UserID", "Username", "Action", "Resource", "ResourceID",
		"Severity", "Status", "IPAddress", "UserAgent", "Metadata",
	}
	_ = writer.Write(header) // #nosec G104 -- error writing HTTP response is not recoverable

	// Write data
	for _, log := range logs {
		// Convert metadata to JSON string
		metadataJSON, _ := json.Marshal(log.Metadata)

		row := []string{
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			log.UserID,
			log.Username,
			string(log.Action),
			log.Resource,
			log.ResourceID,
			string(log.Severity),
			log.Status,
			log.IPAddress,
			log.UserAgent,
			string(metadataJSON),
		}
		_ = writer.Write(row) // #nosec G104 -- error writing HTTP response is not recoverable
	}
}

// exportAuditLogsToJSON exports audit logs to JSON format
func exportAuditLogsToJSON(w http.ResponseWriter, logs []*access.AuditLog) {
	// Convert logs to response format
	var logResponses []AuditLogResponse
	for _, log := range logs {
		logResponses = append(logResponses, convertAuditLogToResponse(log))
	}

	// Write JSON
	_ = json.NewEncoder(w).Encode(logResponses) // Best effort, headers already sent
}

// convertAuditLogToResponse converts an audit log to a response format
func convertAuditLogToResponse(log *access.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:         log.ID,
		Timestamp:  log.Timestamp.Format(time.RFC3339),
		UserID:     log.UserID,
		Username:   log.Username,
		Action:     string(log.Action),
		Resource:   log.Resource,
		ResourceID: log.ResourceID,
		Severity:   string(log.Severity),
		Status:     log.Status,
		IPAddress:  log.IPAddress,
		UserAgent:  log.UserAgent,
		Details:    log.Metadata,
		Metadata:   log.Metadata,
	}
}
