// Package scan provides API endpoints for managing red-team scans
package scan

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Handler handles HTTP requests for scan management
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new scan handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// RegisterRoutes registers the scan API routes with the given router
func (h *Handler) RegisterRoutes(mux *http.ServeMux, middleware func(http.Handler) http.Handler) {
	// Scan config endpoints
	mux.Handle("/api/v1/scan-configs", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.listScanConfigs(w, r)
		case http.MethodPost:
			h.createScanConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/scan-configs/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) != 3 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getScanConfig(w, r)
		case http.MethodPut:
			h.updateScanConfig(w, r)
		case http.MethodDelete:
			h.deleteScanConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Scan execution endpoints
	mux.Handle("/api/v1/scans", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.listScans(w, r)
		case http.MethodPost:
			h.createScan(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/scans/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Handle different endpoints
		if len(pathParts) == 3 {
			// /api/v1/scans/{id}
			if r.Method == http.MethodGet {
				h.getScan(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(pathParts) == 4 {
			// /api/v1/scans/{id}/cancel or /api/v1/scans/{id}/results
			if pathParts[3] == "cancel" && r.Method == http.MethodPost {
				h.cancelScan(w, r)
			} else if pathParts[3] == "results" && r.Method == http.MethodGet {
				h.getScanResults(w, r)
			} else {
				http.Error(w, "Invalid path or method", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	})))
}

// listScanConfigs handles requests to list scan configurations
func (h *Handler) listScanConfigs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, pageSize := getPaginationParams(r)
	filters := getFilterParams(r)

	// Get scan configs
	configs, pagination, err := h.service.ListScanConfigs(r.Context(), page, pageSize, filters)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list scan configurations", err)
		return
	}

	// Respond with the list
	h.respondWithJSON(w, http.StatusOK, ListResponse{
		Pagination: *pagination,
		Data:       configs,
	})
}

// createScanConfig handles requests to create a new scan configuration
func (h *Handler) createScanConfig(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateScanConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request parameters", err)
		return
	}

	// Get user ID from context (in a real implementation, this would come from authentication)
	userID := getUserIDFromContext(r.Context())

	// Create scan config
	config, err := h.service.CreateScanConfig(r.Context(), req, userID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create scan configuration", err)
		return
	}

	// Respond with the created config
	h.respondWithJSON(w, http.StatusCreated, config)
}

// getScanConfig handles requests to get a scan configuration by ID
func (h *Handler) getScanConfig(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/scan-configs/")

	// Get scan config
	config, err := h.service.GetScanConfig(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Scan configuration not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get scan configuration", err)
		}
		return
	}

	// Respond with the config
	h.respondWithJSON(w, http.StatusOK, config)
}

// updateScanConfig handles requests to update a scan configuration
func (h *Handler) updateScanConfig(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/scan-configs/")

	// Parse request body
	var req UpdateScanConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Update scan config
	config, err := h.service.UpdateScanConfig(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Scan configuration not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to update scan configuration", err)
		}
		return
	}

	// Respond with the updated config
	h.respondWithJSON(w, http.StatusOK, config)
}

// deleteScanConfig handles requests to delete a scan configuration
func (h *Handler) deleteScanConfig(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/scan-configs/")

	// Delete scan config
	if err := h.service.DeleteScanConfig(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Scan configuration not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to delete scan configuration", err)
		}
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusNoContent)
}

// listScans handles requests to list scans
func (h *Handler) listScans(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, pageSize := getPaginationParams(r)
	filters := getFilterParams(r)

	// Get scans
	scans, pagination, err := h.service.ListScans(r.Context(), page, pageSize, filters)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list scans", err)
		return
	}

	// Respond with the list
	h.respondWithJSON(w, http.StatusOK, ListResponse{
		Pagination: *pagination,
		Data:       scans,
	})
}

// createScan handles requests to create a new scan
func (h *Handler) createScan(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request parameters", err)
		return
	}

	// Create scan
	scan, err := h.service.CreateScan(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondWithError(w, http.StatusNotFound, "Scan configuration not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to create scan", err)
		}
		return
	}

	// Respond with the created scan
	h.respondWithJSON(w, http.StatusAccepted, scan)
}

// getScan handles requests to get a scan by ID
func (h *Handler) getScan(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/scans/")

	// Get scan
	scan, err := h.service.GetScan(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Scan not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get scan", err)
		}
		return
	}

	// Respond with the scan
	h.respondWithJSON(w, http.StatusOK, scan)
}

// cancelScan handles requests to cancel a running scan
func (h *Handler) cancelScan(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	path := r.URL.Path
	id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/scans/"), "/cancel")

	// Cancel scan
	scan, err := h.service.CancelScan(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Scan not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to cancel scan", err)
		}
		return
	}

	// Respond with the updated scan
	h.respondWithJSON(w, http.StatusOK, scan)
}

// getScanResults handles requests to get the results for a scan
func (h *Handler) getScanResults(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	path := r.URL.Path
	id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/scans/"), "/results")

	// Parse query parameters
	page, pageSize := getPaginationParams(r)
	filters := getFilterParams(r)

	// Get scan results
	results, pagination, err := h.service.GetScanResults(r.Context(), id, page, pageSize, filters)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "Scan not found", err)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get scan results", err)
		}
		return
	}

	// Respond with the results
	h.respondWithJSON(w, http.StatusOK, ListResponse{
		Pagination: *pagination,
		Data:       results,
	})
}

// respondWithJSON writes a JSON response
func (h *Handler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error but don't try to write a response
		// since we've already written the status code
		// In a real implementation, we would use a proper logger
		// instead of fmt.Printf
		// fmt.Printf("Error encoding JSON response: %v\n", err)
	}
}

// respondWithError writes an error response
func (h *Handler) respondWithError(w http.ResponseWriter, status int, message string, err error) {
	// In a real implementation, we would log the error
	// fmt.Printf("Error: %s: %v\n", message, err)

	// Create error response
	response := ErrorResponse{
		Error: message,
		Code:  status,
	}

	// Write response
	h.respondWithJSON(w, status, response)
}

// Helper functions

// getPaginationParams extracts pagination parameters from the request
func getPaginationParams(r *http.Request) (int, int) {
	// Get page parameter
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Get page size parameter
	pageSizeStr := r.URL.Query().Get("page_size")
	pageSize := 10
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	return page, pageSize
}

// getFilterParams extracts filter parameters from the request
func getFilterParams(r *http.Request) FilterParams {
	return FilterParams{
		Status:    r.URL.Query().Get("status"),
		Severity:  r.URL.Query().Get("severity"),
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Search:    r.URL.Query().Get("search"),
	}
}

// getUserIDFromContext gets the user ID from the request context
// In a real implementation, this would come from authentication middleware
func getUserIDFromContext(ctx interface{}) string {
	// Placeholder implementation
	return "user-1"
}
