// Package mfa provides multi-factor authentication functionality
package mfa

import (
	"context"
	"errors"
	"sync"
)

// MockMFAManager is a mock implementation of the MFAManager interface for testing
type MockMFAManager struct {
	totpSecrets     map[string]string
	backupCodes     map[string][]MFABackupCode
	smsCodes        map[string]string
	webAuthnDevices map[string][]WebAuthnDevice
	mfaMethods      map[string][]MFAMethod
	mu              sync.RWMutex
}

// NewMockMFAManager creates a new mock MFA manager
func NewMockMFAManager() *MockMFAManager {
	return &MockMFAManager{
		totpSecrets:     make(map[string]string),
		backupCodes:     make(map[string][]MFABackupCode),
		smsCodes:        make(map[string]string),
		webAuthnDevices: make(map[string][]WebAuthnDevice),
		mfaMethods:      make(map[string][]MFAMethod),
	}
}

// GetMFASettings gets the MFA settings for a user
func (m *MockMFAManager) GetMFASettings(ctx context.Context, userID string) (*MFASettings, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	methods := m.mfaMethods[userID]
	if len(methods) == 0 {
		return &MFASettings{
			UserID:        userID,
			Enabled:       false,
			DefaultMethod: "",
			Methods:       make(map[MFAMethod]MFAStatus),
		}, nil
	}

	methodSettings := make(map[MFAMethod]MFAStatus)
	for _, method := range methods {
		methodSettings[method] = MFAStatusEnabled
	}

	return &MFASettings{
		UserID:        userID,
		Enabled:       true,
		DefaultMethod: methods[0],
		Methods:       methodSettings,
	}, nil
}

// EnableMFA enables MFA for a user
func (m *MockMFAManager) EnableMFA(ctx context.Context, userID string, method MFAMethod) (*MFASettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	methods := m.mfaMethods[userID]
	if !containsMFAMethod(methods, method) {
		methods = append(methods, method)
	}
	m.mfaMethods[userID] = methods

	methodSettings := make(map[MFAMethod]MFAStatus)
	for _, method := range methods {
		methodSettings[method] = MFAStatusEnabled
	}

	return &MFASettings{
		UserID:        userID,
		Enabled:       true,
		DefaultMethod: methods[0],
		Methods:       methodSettings,
	}, nil
}

// DisableMFA disables MFA for a user
func (m *MockMFAManager) DisableMFA(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.mfaMethods, userID)
	delete(m.totpSecrets, userID)
	delete(m.backupCodes, userID)
	delete(m.smsCodes, userID)
	delete(m.webAuthnDevices, userID)

	return nil
}

// SetupTOTP sets up TOTP for a user
func (m *MockMFAManager) SetupTOTP(ctx context.Context, userID string, username string) (*TOTPConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	secret, err := GenerateRandomSecret(20)
	if err != nil {
		return nil, err
	}

	m.totpSecrets[userID] = secret
	qrCodeURL := GenerateTOTPQRCodeURL("LLMrecon", username, secret)

	return &TOTPConfig{
		Secret:    secret,
		QRCodeURL: qrCodeURL,
	}, nil
}

// VerifyTOTPSetup verifies TOTP setup
func (m *MockMFAManager) VerifyTOTPSetup(ctx context.Context, userID string, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.totpSecrets[userID]
	if !exists {
		return errors.New("TOTP not set up for user")
	}

	// For mock purposes, we'll accept any 6-digit code
	if len(code) != 6 {
		return errors.New("invalid TOTP code format")
	}

	// In a real implementation, we would validate the TOTP code
	// For mock purposes, we'll just accept "123456"
	if code != "123456" {
		return errors.New("invalid TOTP code")
	}

	// Add TOTP to user's MFA methods
	methods := m.mfaMethods[userID]
	if !containsMFAMethod(methods, MFAMethodTOTP) {
		methods = append(methods, MFAMethodTOTP)
		m.mfaMethods[userID] = methods
	}

	return nil
}

// GenerateBackupCodes generates backup codes for a user
func (m *MockMFAManager) GenerateBackupCodes(ctx context.Context, userID string) ([]MFABackupCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate 10 backup codes
	codes := make([]MFABackupCode, 10)
	for i := 0; i < 10; i++ {
		code, err := GenerateBackupCode()
		if err != nil {
			return nil, err
		}
		codes[i] = MFABackupCode{
			Code: code,
			Used: false,
		}
	}

	m.backupCodes[userID] = codes

	// Add backup codes to user's MFA methods
	methods := m.mfaMethods[userID]
	if !containsMFAMethod(methods, MFAMethodBackupCode) {
		methods = append(methods, MFAMethodBackupCode)
		m.mfaMethods[userID] = methods
	}

	return codes, nil
}

// SetupWebAuthn initiates WebAuthn setup
func (m *MockMFAManager) SetupWebAuthn(ctx context.Context, userID string, username string, displayName string) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For mock purposes, return a simple options object
	options := map[string]interface{}{
		"challenge": "random-challenge-string",
		"rp": map[string]string{
			"name": "LLMrecon",
			"id":   "LLMrecon.example.com",
		},
		"user": map[string]interface{}{
			"id":          userID,
			"name":        username,
			"displayName": displayName,
		},
	}

	return options, nil
}

// VerifyWebAuthnSetup verifies WebAuthn setup
func (m *MockMFAManager) VerifyWebAuthnSetup(ctx context.Context, userID string, attestationResponse string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For mock purposes, we'll accept any non-empty string
	if attestationResponse == "" {
		return errors.New("invalid attestation response")
	}

	// Add a mock WebAuthn device
	devices := m.webAuthnDevices[userID]
	devices = append(devices, WebAuthnDevice{
		ID:        "mock-device-id",
		Name:      "Mock Device",
		CreatedAt: time.Now(),
	})
	m.webAuthnDevices[userID] = devices

	// Add WebAuthn to user's MFA methods
	methods := m.mfaMethods[userID]
	if !containsMFAMethod(methods, MFAMethodWebAuthn) {
		methods = append(methods, MFAMethodWebAuthn)
		m.mfaMethods[userID] = methods
	}

	return nil
}

// SetupSMS initiates SMS setup
func (m *MockMFAManager) SetupSMS(ctx context.Context, userID string, phoneNumber string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a code
	code, err := GenerateRandomCode(6)
	if err != nil {
		return err
	}

	m.smsCodes[userID] = code

	// In a real implementation, we would send the SMS
	// For mock purposes, we'll just store the code

	return nil
}

// VerifySMSSetup verifies SMS setup
func (m *MockMFAManager) VerifySMSSetup(ctx context.Context, userID string, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	storedCode, exists := m.smsCodes[userID]
	if !exists {
		return errors.New("SMS verification not initiated")
	}

	if code != storedCode {
		return errors.New("invalid SMS code")
	}

	// Add SMS to user's MFA methods
	methods := m.mfaMethods[userID]
	if !containsMFAMethod(methods, MFAMethodSMS) {
		methods = append(methods, MFAMethodSMS)
		m.mfaMethods[userID] = methods
	}

	// Clear the code
	delete(m.smsCodes, userID)

	return nil
}

// VerifyMFA verifies an MFA code
func (m *MockMFAManager) VerifyMFA(ctx context.Context, userID string, method MFAMethod, code string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	methods := m.mfaMethods[userID]
	if !containsMFAMethod(methods, method) {
		return false, errors.New("MFA method not enabled for user")
	}

	switch method {
	case MFAMethodTOTP:
		// For mock purposes, we'll accept "123456"
		return code == "123456", nil
	case MFAMethodBackupCode:
		codes := m.backupCodes[userID]
		valid, index := IsBackupCodeValid(code, codes)
		if valid {
			// Mark the code as used
			m.mu.RUnlock()
			m.mu.Lock()
			defer m.mu.Unlock()
			
			codes := m.backupCodes[userID]
			codes[index].Used = true
			m.backupCodes[userID] = codes
			
			return true, nil
		}
		return false, nil
	case MFAMethodWebAuthn:
		// For mock purposes, we'll accept "mock-assertion-response"
		return code == "mock-assertion-response", nil
	case MFAMethodSMS:
		storedCode := m.smsCodes[userID]
		return code == storedCode, nil
	default:
		return false, errors.New("unsupported MFA method")
	}
}

// InitiateSMSVerification initiates SMS verification
func (m *MockMFAManager) InitiateSMSVerification(ctx context.Context, userID string) error {
	return m.SetupSMS(ctx, userID, "")
}

// VerifySMSCode verifies an SMS code
func (m *MockMFAManager) VerifySMSCode(ctx context.Context, userID string, code string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storedCode, exists := m.smsCodes[userID]
	if !exists {
		return false, errors.New("SMS verification not initiated")
	}

	if code != storedCode {
		return false, nil
	}

	// Clear the code
	m.mu.RUnlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.smsCodes, userID)
	
	return true, nil
}

// InitiateWebAuthnVerification initiates WebAuthn verification
func (m *MockMFAManager) InitiateWebAuthnVerification(ctx context.Context, userID string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For mock purposes, return a simple options object
	options := map[string]interface{}{
		"challenge": "random-challenge-string",
		"rpId":      "LLMrecon.example.com",
		"timeout":   60000,
	}

	return options, nil
}

// VerifyWebAuthnAssertion verifies a WebAuthn assertion
func (m *MockMFAManager) VerifyWebAuthnAssertion(ctx context.Context, userID string, assertionResponse string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For mock purposes, we'll accept "mock-assertion-response"
	return assertionResponse == "mock-assertion-response", nil
}

// ValidateMFASettings validates MFA settings
func (m *MockMFAManager) ValidateMFASettings(ctx context.Context, userID string) (bool, []string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	methods := m.mfaMethods[userID]
	if len(methods) == 0 {
		return false, nil, nil
	}

	var issues []string
	valid := true

	// Check TOTP
	if containsMFAMethod(methods, MFAMethodTOTP) {
		if _, exists := m.totpSecrets[userID]; !exists {
			valid = false
			issues = append(issues, "TOTP secret not set")
		}
	}

	// Check backup codes
	if containsMFAMethod(methods, MFAMethodBackupCode) {
		codes := m.backupCodes[userID]
		if len(codes) == 0 {
			valid = false
			issues = append(issues, "No backup codes generated")
		}
	}

	// Check WebAuthn
	if containsMFAMethod(methods, MFAMethodWebAuthn) {
		devices := m.webAuthnDevices[userID]
		if len(devices) == 0 {
			valid = false
			issues = append(issues, "No WebAuthn devices registered")
		}
	}

	return valid, issues, nil
}

// Helper methods for testing

// GenerateTOTPSecret generates a TOTP secret for testing
func (m *MockMFAManager) GenerateTOTPSecret(userID string) (string, error) {
	secret, err := GenerateRandomSecret(20)
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.totpSecrets[userID] = secret

	return secret, nil
}

// GenerateTOTPQRCodeURL generates a TOTP QR code URL for testing
func (m *MockMFAManager) GenerateTOTPQRCodeURL(username, secret string) string {
	return GenerateTOTPQRCodeURL("LLMrecon", username, secret)
}

// VerifyTOTPCode verifies a TOTP code for testing
func (m *MockMFAManager) VerifyTOTPCode(userID, secret, code string) (bool, error) {
	// For mock purposes, we'll accept "123456"
	return code == "123456", nil
}

// VerifyBackupCode verifies a backup code for testing
func (m *MockMFAManager) VerifyBackupCode(userID, code string) (bool, error) {
	m.mu.RLock()
	
	codes := m.backupCodes[userID]
	valid, index := IsBackupCodeValid(code, codes)
	
	m.mu.RUnlock()
	
	if valid {
		m.mu.Lock()
		defer m.mu.Unlock()
		
		codes := m.backupCodes[userID]
		codes[index].Used = true
		m.backupCodes[userID] = codes
	}
	
	return valid, nil
}

// GenerateSMSCode generates an SMS code for testing
func (m *MockMFAManager) GenerateSMSCode(userID string) (string, error) {
	code, err := GenerateRandomCode(6)
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.smsCodes[userID] = code

	return code, nil
}

// SendSMS sends an SMS for testing
func (m *MockMFAManager) SendSMS(userID, phoneNumber, message string) error {
	// In a real implementation, we would send the SMS
	// For mock purposes, this is a no-op
	return nil
}

// GenerateWebAuthnRegistrationOptions generates WebAuthn registration options for testing
func (m *MockMFAManager) GenerateWebAuthnRegistrationOptions(userID string) (map[string]interface{}, error) {
	return m.SetupWebAuthn(context.Background(), userID, userID, userID)
}

// VerifyWebAuthnRegistration verifies WebAuthn registration for testing
func (m *MockMFAManager) VerifyWebAuthnRegistration(userID, attestationResponse string) error {
	return m.VerifyWebAuthnSetup(context.Background(), userID, attestationResponse)
}

// GenerateWebAuthnAuthenticationOptions generates WebAuthn authentication options for testing
func (m *MockMFAManager) GenerateWebAuthnAuthenticationOptions(userID string) (map[string]interface{}, error) {
	return m.InitiateWebAuthnVerification(context.Background(), userID)
}

// VerifyWebAuthnAuthentication verifies WebAuthn authentication for testing
func (m *MockMFAManager) VerifyWebAuthnAuthentication(userID, assertionResponse string) error {
	valid, err := m.VerifyWebAuthnAssertion(context.Background(), userID, assertionResponse)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid WebAuthn assertion")
	}
	return nil
}

// GetMFAMethods gets the MFA methods for a user for testing
func (m *MockMFAManager) GetMFAMethods(userID string) ([]MFAMethod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mfaMethods[userID], nil
}

// IsMFAEnabled checks if MFA is enabled for a user for testing
func (m *MockMFAManager) IsMFAEnabled(userID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mfaMethods[userID]) > 0, nil
}

// DisableMFAMethod disables an MFA method for a user for testing
func (m *MockMFAManager) DisableMFAMethod(userID string, method MFAMethod) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	methods := m.mfaMethods[userID]
	newMethods := make([]MFAMethod, 0, len(methods))
	for _, m := range methods {
		if m != method {
			newMethods = append(newMethods, m)
		}
	}
	m.mfaMethods[userID] = newMethods

	return nil
}

// VerifyMFACode verifies an MFA code for testing
func (m *MockMFAManager) VerifyMFACode(ctx context.Context, userID, methodStr, code string) (bool, error) {
	method := MFAMethod(methodStr)
	return m.VerifyMFA(ctx, userID, method, code)
}

// CleanupExpiredCodes cleans up expired codes for testing
func (m *MockMFAManager) CleanupExpiredCodes() {
	// In a real implementation, we would clean up expired codes
	// For mock purposes, this is a no-op
}

// Close closes the MFA manager for testing
func (m *MockMFAManager) Close() error {
	// In a real implementation, we would close connections
	// For mock purposes, this is a no-op
	return nil
}

// Helper function to check if a slice contains a value
func containsMFAMethod(slice []MFAMethod, item MFAMethod) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
