// Package mfa provides multi-factor authentication functionality
package mfa

import (
	"encoding/json"
	"net/http"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// MFAHandler handles MFA-related API requests
type MFAHandler struct {
	mfaManager MFAManager
	sessionStore interfaces.SessionStore

}
// NewMFAHandler creates a new MFA handler
func NewMFAHandler(mfaManager MFAManager, sessionStore interfaces.SessionStore) *MFAHandler {
	return &MFAHandler{
		mfaManager:   mfaManager,
		sessionStore: sessionStore,
	}

// SetupTOTPHandler handles TOTP setup requests
}
func (h *MFAHandler) SetupTOTPHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	session, err := getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Setup TOTP
	totpConfig, err := h.mfaManager.SetupTOTP(r.Context(), session.UserID, session.UserID)
	if err != nil {
		http.Error(w, "Failed to setup TOTP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return TOTP configuration
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"secret":   totpConfig.Secret,
		"qr_code":  totpConfig.QRCodeURL,
	})

// VerifyTOTPSetupHandler handles TOTP verification requests
}
func (h *MFAHandler) VerifyTOTPSetupHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	session, err := getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify TOTP setup
	err = h.mfaManager.VerifyTOTPSetup(r.Context(), session.UserID, req.Code)
	if err != nil {
		http.Error(w, "Failed to verify TOTP: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update session to mark MFA as completed
	session.MFACompleted = true
	if err := h.sessionStore.UpdateSession(r.Context(), session); err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "TOTP setup verified successfully",
	})

// GenerateBackupCodesHandler handles backup code generation requests
}
func (h *MFAHandler) GenerateBackupCodesHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	session, err := getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate backup codes
	backupCodes, err := h.mfaManager.GenerateBackupCodes(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "Failed to generate backup codes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract code strings
	codes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		codes[i] = code.Code
	}

	// Return backup codes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backup_codes": codes,
	})

// VerifyMFAHandler handles MFA verification requests
}
func (h *MFAHandler) VerifyMFAHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	session, err := getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Method MFAMethod `json:"method"`
		Code   string    `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify MFA
	valid, err := h.mfaManager.VerifyMFA(r.Context(), session.UserID, req.Method, req.Code)
	if err != nil {
		http.Error(w, "Failed to verify MFA: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !valid {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	// Update session to mark MFA as completed
	session.MFACompleted = true
	if err := h.sessionStore.UpdateSession(r.Context(), session); err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MFA verified successfully",
	})

// DisableMFAHandler handles MFA disabling requests
}
func (h *MFAHandler) DisableMFAHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	session, err := getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Disable MFA
	err = h.mfaManager.DisableMFA(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "Failed to disable MFA: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MFA disabled successfully",
	})

// GetMFASettingsHandler handles MFA settings retrieval requests
}
func (h *MFAHandler) GetMFASettingsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	session, err := getSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get MFA settings
	settings, err := h.mfaManager.GetMFASettings(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "Failed to get MFA settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return MFA settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)

// Helper function to get session from request
}
func getSessionFromRequest(r *http.Request) (*interfaces.Session, error) {
	// This is a placeholder - implement based on your session management
	// For example:
	// sessionID := getSessionIDFromCookie(r)
	// return sessionStore.GetSessionByID(r.Context(), sessionID)
