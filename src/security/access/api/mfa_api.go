package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/security/access"
)

// MFAHandler handles MFA-related API endpoints
type MFAHandler struct {
	authManager *access.AuthManager
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(authManager *access.AuthManager) *MFAHandler {
	return &MFAHandler{
		authManager: authManager,
	}
}

// MFAStatusResponse represents the response for MFA status
type MFAStatusResponse struct {
	Enabled       bool     `json:"enabled"`
	Methods       []string `json:"methods"`
	DefaultMethod string   `json:"default_method"`
	LastUpdated   string   `json:"last_updated,omitempty"`
}

// TOTPSetupResponse represents the response for TOTP setup
type TOTPSetupResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// Note: MFAVerifyRequest is defined in auth_handlers.go

// MFASetupRequest represents a request to set up MFA
type MFASetupRequest struct {
	Method string `json:"method"`
}

// BackupCodesResponse represents the response for backup codes
type BackupCodesResponse struct {
	Codes     []string  `json:"codes"`
	Generated time.Time `json:"generated"`
}

// SMSSetupRequest represents a request to set up SMS verification
type SMSSetupRequest struct {
	PhoneNumber string `json:"phone_number"`
}

// RegisterRoutes registers the MFA API routes
func (h *MFAHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/api/mfa/status", h.handleMFAStatus)
	router.HandleFunc("/api/mfa/enable", h.handleEnableMFA)
	router.HandleFunc("/api/mfa/disable", h.handleDisableMFA)
	router.HandleFunc("/api/mfa/verify", h.handleVerifyMFA)

	// TOTP-specific endpoints
	router.HandleFunc("/api/mfa/totp/setup", h.handleTOTPSetup)
	router.HandleFunc("/api/mfa/totp/verify", h.handleTOTPVerify)

	// Backup codes endpoints
	router.HandleFunc("/api/mfa/backup-codes/generate", h.handleGenerateBackupCodes)

	// WebAuthn endpoints
	router.HandleFunc("/api/mfa/webauthn/register-begin", h.handleWebAuthnRegisterBegin)
	router.HandleFunc("/api/mfa/webauthn/register-complete", h.handleWebAuthnRegisterComplete)
	router.HandleFunc("/api/mfa/webauthn/authenticate-begin", h.handleWebAuthnAuthenticateBegin)
	router.HandleFunc("/api/mfa/webauthn/authenticate-complete", h.handleWebAuthnAuthenticateComplete)

	// SMS endpoints
	router.HandleFunc("/api/mfa/sms/setup", h.handleSMSSetup)
	router.HandleFunc("/api/mfa/sms/verify", h.handleSMSVerify)
}

// handleMFAStatus handles the MFA status endpoint
func (h *MFAHandler) handleMFAStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's MFA status
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Create response
	response := MFAStatusResponse{
		Enabled: user.MFAEnabled,
		Methods: user.MFAMethods,
	}

	// Set default method if available
	if len(user.MFAMethods) > 0 {
		response.DefaultMethod = user.MFAMethods[0]
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// handleEnableMFA handles the enable MFA endpoint
func (h *MFAHandler) handleEnableMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request MFASetupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate method
	if !isValidMFAMethod(request.Method) {
		http.Error(w, "Invalid MFA method", http.StatusBadRequest)
		return
	}

	// Enable MFA for user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update user's MFA settings
	user.MFAEnabled = true

	// Add method if not already present
	methodExists := false
	for _, method := range user.MFAMethods {
		if method == request.Method {
			methodExists = true
			break
		}
	}

	if !methodExists {
		user.MFAMethods = append(user.MFAMethods, request.Method)
	}

	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success": true}`))
}

// handleDisableMFA handles the disable MFA endpoint
func (h *MFAHandler) handleDisableMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Disable MFA for user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update user's MFA settings
	user.MFAEnabled = false

	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success": true}`))
}

// handleVerifyMFA handles the verify MFA endpoint
func (h *MFAHandler) handleVerifyMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify MFA code using the AuthManager's VerifyMFA method
	if err := h.authManager.VerifyMFA(r.Context(), session.ID, request.Code); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		} else {
			http.Error(w, "MFA verification failed", http.StatusInternalServerError)
		}
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success": true}`))
}

// handleTOTPSetup handles the TOTP setup endpoint
func (h *MFAHandler) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TOTP setup requires methods not available on the AuthManager interface.
	// Return a not-implemented response.
	http.Error(w, "TOTP setup is not available through this endpoint", http.StatusNotImplemented)
}

// handleTOTPVerify handles the TOTP verification endpoint
func (h *MFAHandler) handleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Use the general VerifyMFA method
	if err := h.authManager.VerifyMFA(r.Context(), session.ID, request.Code); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
		} else {
			http.Error(w, "TOTP verification failed", http.StatusInternalServerError)
		}
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success": true}`))
}

// handleGenerateBackupCodes handles the generate backup codes endpoint
func (h *MFAHandler) handleGenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Backup code generation requires methods not available on the AuthManager interface.
	http.Error(w, "Backup code generation is not available through this endpoint", http.StatusNotImplemented)
}

// handleWebAuthnRegisterBegin handles the WebAuthn registration begin endpoint
func (h *MFAHandler) handleWebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// WebAuthn registration requires methods not available on the AuthManager interface.
	http.Error(w, "WebAuthn registration is not available through this endpoint", http.StatusNotImplemented)
}

// handleWebAuthnRegisterComplete handles the WebAuthn registration complete endpoint
func (h *MFAHandler) handleWebAuthnRegisterComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// WebAuthn registration requires methods not available on the AuthManager interface.
	http.Error(w, "WebAuthn registration is not available through this endpoint", http.StatusNotImplemented)
}

// handleWebAuthnAuthenticateBegin handles the WebAuthn authentication begin endpoint
func (h *MFAHandler) handleWebAuthnAuthenticateBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// WebAuthn authentication requires methods not available on the AuthManager interface.
	http.Error(w, "WebAuthn authentication is not available through this endpoint", http.StatusNotImplemented)
}

// handleWebAuthnAuthenticateComplete handles the WebAuthn authentication complete endpoint
func (h *MFAHandler) handleWebAuthnAuthenticateComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// WebAuthn authentication requires methods not available on the AuthManager interface.
	http.Error(w, "WebAuthn authentication is not available through this endpoint", http.StatusNotImplemented)
}

// handleSMSSetup handles the SMS setup endpoint
func (h *MFAHandler) handleSMSSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	_, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// SMS setup requires methods not available on the AuthManager interface.
	http.Error(w, "SMS setup is not available through this endpoint", http.StatusNotImplemented)
}

// handleSMSVerify handles the SMS verification endpoint
func (h *MFAHandler) handleSMSVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Use the general VerifyMFA method
	if err := h.authManager.VerifyMFA(r.Context(), session.ID, request.Code); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			http.Error(w, "Invalid verification code", http.StatusUnauthorized)
		} else {
			http.Error(w, "SMS verification failed", http.StatusInternalServerError)
		}
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success": true}`))
}

// Helper functions

// getSessionFromRequest gets the session from the request
func (h *MFAHandler) getSessionFromRequest(r *http.Request) (*access.Session, error) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	// Validate session
	session, err := h.authManager.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// isValidMFAMethod checks if the given method is a valid MFA method
func isValidMFAMethod(method string) bool {
	validMethods := []string{
		"totp",
		"backup_code",
		"webauthn",
		"sms",
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}

	return false
}
