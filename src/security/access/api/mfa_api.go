package api

import (
	"encoding/json"
	"net/http"

	"github.com/perplext/LLMrecon/src/security/access"
	"github.com/perplext/LLMrecon/src/security/access/common"
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
	Enabled      bool                 `json:"enabled"`
	Methods      []common.AuthMethod  `json:"methods"`
	DefaultMethod common.AuthMethod   `json:"default_method"`
	LastUpdated  time.Time            `json:"last_updated,omitempty"`
}

// TOTPSetupResponse represents the response for TOTP setup
type TOTPSetupResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// MFAVerifyRequest represents a request to verify an MFA code
type MFAVerifyRequest struct {
	Method common.AuthMethod `json:"method"`
	Code   string            `json:"code"`
}

// MFASetupRequest represents a request to set up MFA
type MFASetupRequest struct {
	Method common.AuthMethod `json:"method"`
}

// BackupCodesResponse represents the response for backup codes
type BackupCodesResponse struct {
	Codes     []string `json:"codes"`
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
	json.NewEncoder(w).Encode(response)
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
	w.Write([]byte(`{"success": true}`))
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
	w.Write([]byte(`{"success": true}`))
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Verify MFA code
	if !h.authManager.VerifyMFACode(user, request.Code) {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}
	
	// Update session to indicate MFA is completed
	session.MFACompleted = true
	if err := h.authManager.UpdateSession(r.Context(), session); err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// handleTOTPSetup handles the TOTP setup endpoint
func (h *MFAHandler) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Generate TOTP secret
	secret, err := h.authManager.GenerateTOTPSecret()
	if err != nil {
		http.Error(w, "Failed to generate TOTP secret", http.StatusInternalServerError)
		return
	}
	
	// Save secret to user
	user.MFASecret = secret
	
	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	// Generate QR code URL
	qrCodeURL := h.authManager.GenerateTOTPQRCodeURL(user.Username, secret)
	
	// Send response
	response := TOTPSetupResponse{
		Secret:    secret,
		QRCodeURL: qrCodeURL,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Verify TOTP code
	if !h.authManager.VerifyTOTPCode(user.MFASecret, request.Code) {
		http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
		return
	}
	
	// Enable TOTP for user
	methodExists := false
	for _, method := range user.MFAMethods {
		if method == common.AuthMethodTOTP {
			methodExists = true
			break
		}
	}
	
	if !methodExists {
		user.MFAMethods = append(user.MFAMethods, common.AuthMethodTOTP)
	}
	
	user.MFAEnabled = true
	
	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// handleGenerateBackupCodes handles the generate backup codes endpoint
func (h *MFAHandler) handleGenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Generate backup codes
	codes, err := h.authManager.GenerateBackupCodes()
	if err != nil {
		http.Error(w, "Failed to generate backup codes", http.StatusInternalServerError)
		return
	}
	
	// Save backup codes to user's metadata
	if user.Metadata == nil {
		user.Metadata = make(map[string]interface{})
	}
	
	// Store backup codes in user metadata
	user.Metadata["backup_codes"] = codes
	user.Metadata["backup_codes_generated"] = time.Now()
	
	// Enable backup codes method
	methodExists := false
	for _, method := range user.MFAMethods {
		if method == common.AuthMethodBackupCode {
			methodExists = true
			break
		}
	}
	
	if !methodExists {
		user.MFAMethods = append(user.MFAMethods, common.AuthMethodBackupCode)
	}
	
	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	// Send response
	response := BackupCodesResponse{
		Codes:     codes,
		Generated: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleWebAuthnRegisterBegin handles the WebAuthn registration begin endpoint
func (h *MFAHandler) handleWebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Generate WebAuthn registration options
	options, err := h.authManager.GenerateWebAuthnRegistrationOptions(user)
	if err != nil {
		http.Error(w, "Failed to generate registration options", http.StatusInternalServerError)
		return
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

// handleWebAuthnRegisterComplete handles the WebAuthn registration complete endpoint
func (h *MFAHandler) handleWebAuthnRegisterComplete(w http.ResponseWriter, r *http.Request) {
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
		AttestationResponse string `json:"attestationResponse"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Verify WebAuthn registration
	if err := h.authManager.VerifyWebAuthnRegistration(user, request.AttestationResponse); err != nil {
		http.Error(w, "Failed to verify WebAuthn registration: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Enable WebAuthn for user
	methodExists := false
	for _, method := range user.MFAMethods {
		if method == common.AuthMethodWebAuthn {
			methodExists = true
			break
		}
	}
	
	if !methodExists {
		user.MFAMethods = append(user.MFAMethods, common.AuthMethodWebAuthn)
	}
	
	user.MFAEnabled = true
	
	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// handleWebAuthnAuthenticateBegin handles the WebAuthn authentication begin endpoint
func (h *MFAHandler) handleWebAuthnAuthenticateBegin(w http.ResponseWriter, r *http.Request) {
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Generate WebAuthn authentication options
	options, err := h.authManager.GenerateWebAuthnAuthenticationOptions(user)
	if err != nil {
		http.Error(w, "Failed to generate authentication options", http.StatusInternalServerError)
		return
	}
	
	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

// handleWebAuthnAuthenticateComplete handles the WebAuthn authentication complete endpoint
func (h *MFAHandler) handleWebAuthnAuthenticateComplete(w http.ResponseWriter, r *http.Request) {
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
		AssertionResponse string `json:"assertionResponse"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Verify WebAuthn authentication
	if err := h.authManager.VerifyWebAuthnAuthentication(user, request.AssertionResponse); err != nil {
		http.Error(w, "Failed to verify WebAuthn authentication: "+err.Error(), http.StatusUnauthorized)
		return
	}
	
	// Update session to indicate MFA is completed
	session.MFACompleted = true
	if err := h.authManager.UpdateSession(r.Context(), session); err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// handleSMSSetup handles the SMS setup endpoint
func (h *MFAHandler) handleSMSSetup(w http.ResponseWriter, r *http.Request) {
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
	var request SMSSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Validate phone number
	if request.PhoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Save phone number to user's metadata
	if user.Metadata == nil {
		user.Metadata = make(map[string]interface{})
	}
	
	user.Metadata["phone_number"] = request.PhoneNumber
	
	// Generate and send verification code
	code, err := h.authManager.GenerateSMSCode()
	if err != nil {
		http.Error(w, "Failed to generate SMS code", http.StatusInternalServerError)
		return
	}
	
	// Store verification code in user's metadata
	user.Metadata["sms_verification_code"] = code
	user.Metadata["sms_verification_expires"] = time.Now().Add(10 * time.Minute)
	
	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	// Send SMS (in a real implementation, this would use an SMS provider)
	// For now, we'll just log the code
	h.authManager.SendSMS(request.PhoneNumber, "Your verification code is: "+code)
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Verification code sent"}`))
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
	
	// Get user
	user, err := h.authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Verify SMS code
	if user.Metadata == nil || 
	   user.Metadata["sms_verification_code"] != request.Code {
		http.Error(w, "Invalid verification code", http.StatusUnauthorized)
		return
	}
	
	// Check if code has expired
	expiresTime, ok := user.Metadata["sms_verification_expires"].(time.Time)
	if !ok || time.Now().After(expiresTime) {
		http.Error(w, "Verification code has expired", http.StatusUnauthorized)
		return
	}
	
	// Enable SMS for user
	methodExists := false
	for _, method := range user.MFAMethods {
		if method == common.AuthMethodSMS {
			methodExists = true
			break
		}
	}
	
	if !methodExists {
		user.MFAMethods = append(user.MFAMethods, common.AuthMethodSMS)
	}
	
	user.MFAEnabled = true
	
	// Clear verification code
	delete(user.Metadata, "sms_verification_code")
	delete(user.Metadata, "sms_verification_expires")
	
	// Save user
	if err := h.authManager.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
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
func isValidMFAMethod(method common.AuthMethod) bool {
	validMethods := []common.AuthMethod{
		common.AuthMethodTOTP,
		common.AuthMethodBackupCode,
		common.AuthMethodWebAuthn,
		common.AuthMethodSMS,
	}
	
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	
	return false
}
