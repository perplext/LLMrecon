package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Global server instance - TODO: This should be passed through context or dependency injection
var serverInstance *ServerImpl

// Handler placeholder functions - these will be implemented in subsequent tasks

// Health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   APIVersion,
	}
	writeSuccess(w, health)

// Version endpoint
func handleVersion(w http.ResponseWriter, r *http.Request) {
	if serverInstance == nil || serverInstance.services.UpdateManager == nil {
		writeError(w, http.StatusServiceUnavailable, 
			NewAPIError(ErrCodeServiceUnavailable, "Update service not available"))
		return
	}
	
	versionInfo, err := serverInstance.services.UpdateManager.CheckForUpdates()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to get version info", err.Error()))
		return
	}
	
	writeSuccess(w, versionInfo)

// Authentication handlers are implemented in auth_handlers.go

// handleRefreshToken is implemented in auth_handlers.go

// Scan handlers
func handleCreateScan(w http.ResponseWriter, r *http.Request) {
	var req CreateScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest,
			NewAPIError(ErrCodeInvalidRequest, "Invalid request body"))
		return
	}
	
	if serverInstance == nil || serverInstance.services.ScanEngine == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Scan service not available"))
		return
	}
	
	scan, err := serverInstance.services.ScanEngine.CreateScan(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to create scan", err.Error()))
		return
	}
	
	writeSuccess(w, scan)

func handleListScans(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	status := query.Get("status")
	
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	
	if serverInstance == nil || serverInstance.services.ScanEngine == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Scan service not available"))
		return
	}
	
	// Create filter
	filter := ScanFilter{
		Limit: perPage,
	}
	if status != "" {
		filter.Status = ScanStatus(status)
	}
	
	scans, err := serverInstance.services.ScanEngine.ListScans(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to list scans", err.Error()))
		return
	}
	
	// Create pagination metadata
	total := len(scans)
	meta := &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: calculateTotalPages(total, perPage),
	}
	
	// Apply pagination
	offset, limit := paginate(page, perPage, total)
	if limit > 0 && offset < len(scans) {
		end := offset + limit
		if end > len(scans) {
			end = len(scans)
		}
		scans = scans[offset:end]
	} else {
		scans = []Scan{}
	}
	
	writeSuccessWithMeta(w, scans, meta)

func handleGetScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["id"]
	
	if serverInstance == nil || serverInstance.services.ScanEngine == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Scan service not available"))
		return
	}
	
	scan, err := serverInstance.services.ScanEngine.GetScan(scanID)
	if err != nil {
		writeError(w, http.StatusNotFound,
			NewAPIError(ErrCodeNotFound, "Scan not found"))
		return
	}
	
	writeSuccess(w, scan)

func handleCancelScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["id"]
	
	if serverInstance == nil || serverInstance.services.ScanEngine == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Scan service not available"))
		return
	}
	
	err := serverInstance.services.ScanEngine.CancelScan(scanID)
	if err != nil {
		writeError(w, http.StatusNotFound,
			NewAPIError(ErrCodeNotFound, "Scan not found"))
		return
	}
	
	writeSuccess(w, map[string]string{"status": "cancelled"})

func handleGetScanResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["id"]
	
	if serverInstance == nil || serverInstance.services.ScanEngine == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Scan service not available"))
		return
	}
	
	results, err := serverInstance.services.ScanEngine.GetScanResults(scanID)
	if err != nil {
		writeError(w, http.StatusNotFound,
			NewAPIError(ErrCodeNotFound, "Scan results not found"))
		return
	}
	
	writeSuccess(w, results)

// Template handlers
func handleListTemplates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	filter := TemplateFilter{
		Category: query.Get("category"),
		Severity: query.Get("severity"),
		Search:   query.Get("search"),
	}
	
	if tags := query["tags"]; len(tags) > 0 {
		filter.Tags = tags
	}
	
	if serverInstance == nil || serverInstance.services.TemplateManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Template service not available"))
		return
	}
	
	templates, err := serverInstance.services.TemplateManager.ListTemplates(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to list templates", err.Error()))
		return
	}
	
	writeSuccess(w, templates)

func handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]
	
	if serverInstance == nil || serverInstance.services.TemplateManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Template service not available"))
		return
	}
	
	template, err := serverInstance.services.TemplateManager.GetTemplate(templateID)
	if err != nil {
		writeError(w, http.StatusNotFound,
			NewAPIError(ErrCodeNotFound, "Template not found"))
		return
	}
	
	writeSuccess(w, template)

func handleListCategories(w http.ResponseWriter, r *http.Request) {
	if serverInstance == nil || serverInstance.services.TemplateManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Template service not available"))
		return
	}
	
	categories, err := serverInstance.services.TemplateManager.GetCategories()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to get categories", err.Error()))
		return
	}
	
	writeSuccess(w, categories)

// Module handlers
func handleListModules(w http.ResponseWriter, r *http.Request) {
	if serverInstance == nil || serverInstance.services.ModuleManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Module service not available"))
		return
	}
	
	modules, err := serverInstance.services.ModuleManager.ListModules()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to list modules", err.Error()))
		return
	}
	
	writeSuccess(w, modules)

func handleGetModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	moduleID := vars["id"]
	
	if serverInstance == nil || serverInstance.services.ModuleManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Module service not available"))
		return
	}
	
	module, err := serverInstance.services.ModuleManager.GetModule(moduleID)
	if err != nil {
		writeError(w, http.StatusNotFound,
			NewAPIError(ErrCodeNotFound, "Module not found"))
		return
	}
	
	// Filter out sensitive information
	module.Config.Credentials = nil
	
	writeSuccess(w, module)

func handleUpdateModuleConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	moduleID := vars["id"]
	
	var config ModuleConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest,
			NewAPIError(ErrCodeInvalidRequest, "Invalid request body"))
		return
	}
	
	if serverInstance == nil || serverInstance.services.ModuleManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Module service not available"))
		return
	}
	
	err := serverInstance.services.ModuleManager.UpdateModuleConfig(moduleID, config)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to update module config", err.Error()))
		return
	}
	
	writeSuccess(w, map[string]string{"status": "updated"})

// Update handlers
func handleCheckUpdate(w http.ResponseWriter, r *http.Request) {
	if serverInstance == nil || serverInstance.services.UpdateManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Update service not available"))
		return
	}
	
	versionInfo, err := serverInstance.services.UpdateManager.CheckForUpdates()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to check for updates", err.Error()))
		return
	}
	
	writeSuccess(w, versionInfo)

func handlePerformUpdate(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest,
			NewAPIError(ErrCodeInvalidRequest, "Invalid request body"))
		return
	}
	
	if serverInstance == nil || serverInstance.services.UpdateManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Update service not available"))
		return
	}
	
	response, err := serverInstance.services.UpdateManager.PerformUpdate(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to perform update", err.Error()))
		return
	}
	
	writeSuccess(w, response)
// Bundle handlers
func handleListBundles(w http.ResponseWriter, r *http.Request) {
	if serverInstance == nil || serverInstance.services.BundleManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Bundle service not available"))
		return
	}
	
	bundles, err := serverInstance.services.BundleManager.ListBundles()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to list bundles", err.Error()))
		return
	}
	
	writeSuccess(w, bundles)

func handleExportBundle(w http.ResponseWriter, r *http.Request) {
	var req ExportBundleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest,
			NewAPIError(ErrCodeInvalidRequest, "Invalid request body"))
		return
	}
	
	if serverInstance == nil || serverInstance.services.BundleManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Bundle service not available"))
		return
	}
	
	result, err := serverInstance.services.BundleManager.ExportBundle(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to export bundle", err.Error()))
		return
	}
	
	writeSuccess(w, result)

func handleImportBundle(w http.ResponseWriter, r *http.Request) {
	var req ImportBundleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest,
			NewAPIError(ErrCodeInvalidRequest, "Invalid request body"))
		return
	}
	
	if serverInstance == nil || serverInstance.services.BundleManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Bundle service not available"))
		return
	}
	
	result, err := serverInstance.services.BundleManager.ImportBundle(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to import bundle", err.Error()))
		return
	}
	
	writeSuccess(w, result)

// Compliance handlers
func handleGenerateComplianceReport(w http.ResponseWriter, r *http.Request) {
	var req ComplianceReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest,
			NewAPIError(ErrCodeInvalidRequest, "Invalid request body"))
		return
	}
	
	if serverInstance == nil || serverInstance.services.ComplianceManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Compliance service not available"))
		return
	}
	
	report, err := serverInstance.services.ComplianceManager.GenerateReport(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to generate report", err.Error()))
		return
	}
	
	writeSuccess(w, report)

func handleCheckCompliance(w http.ResponseWriter, r *http.Request) {
	framework := r.URL.Query().Get("framework")
	if framework == "" {
		framework = "owasp"
	}
	
	if serverInstance == nil || serverInstance.services.ComplianceManager == nil {
		writeError(w, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "Compliance service not available"))
		return
	}
	
	status, err := serverInstance.services.ComplianceManager.CheckCompliance(framework)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			NewAPIErrorWithDetails(ErrCodeInternalError, "Failed to check compliance", err.Error()))
		return
	}
	
	writeSuccess(w, status)

// OpenAPI documentation handler is implemented in openapi.go

// Helper functions
func generateMockToken() string {
	return uuid.New().String()
