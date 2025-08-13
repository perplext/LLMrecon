package mfa

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// MFAMethod represents the type of MFA method
type MFAMethod string

const (
	// MFAMethodTOTP represents Time-based One-Time Password authentication
	MFAMethodTOTP MFAMethod = "totp"
	
	// MFAMethodBackupCode represents backup code authentication
	MFAMethodBackupCode MFAMethod = "backup_code"
	
	// MFAMethodWebAuthn represents WebAuthn/FIDO2 authentication
	MFAMethodWebAuthn MFAMethod = "webauthn"
	
	// MFAMethodSMS represents SMS-based authentication
	MFAMethodSMS MFAMethod = "sms"
)

// MFAStatus represents the status of an MFA method
type MFAStatus string

const (
	// MFAStatusEnabled indicates that the MFA method is enabled
	MFAStatusEnabled MFAStatus = "enabled"
	
	// MFAStatusDisabled indicates that the MFA method is disabled
	MFAStatusDisabled MFAStatus = "disabled"
	
	// MFAStatusPending indicates that the MFA method is pending setup
	MFAStatusPending MFAStatus = "pending"
)

// MFASettings represents the MFA settings for a user
type MFASettings struct {
	// UserID is the ID of the user
	UserID string
	
	// Enabled indicates whether MFA is enabled for the user
	Enabled bool
	
	// DefaultMethod is the default MFA method for the user
	DefaultMethod MFAMethod
	
	// Methods is a map of MFA methods to their status
	Methods map[MFAMethod]MFAStatus
	
	// TOTPConfig is the TOTP configuration for the user
	TOTPConfig *TOTPConfig
	
	// BackupCodes are the backup codes for the user
	BackupCodes []MFABackupCode
	
	// WebAuthnCredentials are the WebAuthn credentials for the user
	WebAuthnCredentials []WebAuthnCredential
	
	// PhoneNumber is the phone number for SMS verification
	PhoneNumber string
	
	// LastUpdated is the time when the settings were last updated
	LastUpdated time.Time
}

// MFAVerification represents an MFA verification
type MFAVerification struct {
	// UserID is the ID of the user
	UserID string
	
	// Method is the MFA method used for verification
	Method MFAMethod
	
	// Challenge is the challenge for WebAuthn verification
	Challenge string
	
	// SMSVerification is the SMS verification
	SMSVerification *SMSVerification
	
	// CreatedAt is the time when the verification was created
	CreatedAt time.Time
	
	// ExpiresAt is the time when the verification expires
	ExpiresAt time.Time
}

// MFAStore defines the interface for storing MFA settings
type MFAStore interface {
	// GetMFASettings gets the MFA settings for a user
	GetMFASettings(ctx context.Context, userID string) (*MFASettings, error)
	
	// SaveMFASettings saves the MFA settings for a user
	SaveMFASettings(ctx context.Context, settings *MFASettings) error
	
	// DeleteMFASettings deletes the MFA settings for a user
	DeleteMFASettings(ctx context.Context, userID string) error
	
	// CreateVerification creates a new MFA verification
	CreateVerification(ctx context.Context, verification *MFAVerification) error
	
	// GetVerification gets an MFA verification
	GetVerification(ctx context.Context, userID string) (*MFAVerification, error)
	
	// DeleteVerification deletes an MFA verification
	DeleteVerification(ctx context.Context, userID string) error
}

// InMemoryMFAStore is an in-memory implementation of MFAStore
type InMemoryMFAStore struct {
	settings      map[string]*MFASettings
	verifications map[string]*MFAVerification
	mu            sync.RWMutex
}

// NewInMemoryMFAStore creates a new in-memory MFA store
func NewInMemoryMFAStore() *InMemoryMFAStore {
	return &InMemoryMFAStore{
		settings:      make(map[string]*MFASettings),
		verifications: make(map[string]*MFAVerification),
	}
}

// GetMFASettings gets the MFA settings for a user
func (s *InMemoryMFAStore) GetMFASettings(ctx context.Context, userID string) (*MFASettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	settings, ok := s.settings[userID]
	if !ok {
		return nil, errors.New("MFA settings not found")
	}
	
	return settings, nil
}

// SaveMFASettings saves the MFA settings for a user
func (s *InMemoryMFAStore) SaveMFASettings(ctx context.Context, settings *MFASettings) error {
	if settings == nil {
		return errors.New("settings cannot be nil")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	settings.LastUpdated = time.Now()
	s.settings[settings.UserID] = settings
	
	return nil
}

// DeleteMFASettings deletes the MFA settings for a user
func (s *InMemoryMFAStore) DeleteMFASettings(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.settings, userID)
	
	return nil
}

// CreateVerification creates a new MFA verification
func (s *InMemoryMFAStore) CreateVerification(ctx context.Context, verification *MFAVerification) error {
	if verification == nil {
		return errors.New("verification cannot be nil")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.verifications[verification.UserID] = verification
	
	return nil
}

// GetVerification gets an MFA verification
func (s *InMemoryMFAStore) GetVerification(ctx context.Context, userID string) (*MFAVerification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	verification, ok := s.verifications[userID]
	if !ok {
		return nil, errors.New("verification not found")
	}
	
	// Check if verification has expired
	if time.Now().After(verification.ExpiresAt) {
		return nil, errors.New("verification has expired")
	}
	
	return verification, nil
}

// DeleteVerification deletes an MFA verification
func (s *InMemoryMFAStore) DeleteVerification(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.verifications, userID)
	
	return nil
}

// MFAManager interface is defined in interface.go

// DefaultMFAManager manages MFA operations
type DefaultMFAManager struct {
	// store is the MFA store
	store MFAStore
	
	// totpConfig is the default TOTP configuration
	totpConfig *TOTPConfig
	
	// backupConfig is the default backup code configuration
	backupConfig *BackupCodeConfig
	
	// webAuthnConfig is the default WebAuthn configuration
	webAuthnConfig *WebAuthnConfig
	
	// smsConfig is the default SMS configuration
	smsConfig *SMSConfig
	
	// verificationExpiration is the expiration time for verifications in minutes
	verificationExpiration int
}

// MFAManagerConfig represents the configuration for MFAManager
type MFAManagerConfig struct {
	// TOTPConfig is the default TOTP configuration
	TOTPConfig *TOTPConfig
	
	// BackupConfig is the default backup code configuration
	BackupConfig *BackupCodeConfig
	
	// WebAuthnConfig is the default WebAuthn configuration
	WebAuthnConfig *WebAuthnConfig
	
	// SMSConfig is the default SMS configuration
	SMSConfig *SMSConfig
	
	// VerificationExpiration is the expiration time for verifications in minutes
	VerificationExpiration int
}

// DefaultMFAManagerConfig returns the default MFA manager configuration
func DefaultMFAManagerConfig() *MFAManagerConfig {
	return &MFAManagerConfig{
		TOTPConfig:             DefaultTOTPConfig(),
		BackupConfig:           DefaultBackupCodeConfig(),
		WebAuthnConfig:         DefaultWebAuthnConfig(),
		SMSConfig:              DefaultSMSConfig(),
		VerificationExpiration: 10,
	}
}

// NewDefaultMFAManager creates a new MFA manager
func NewDefaultMFAManager(store MFAStore, config *MFAManagerConfig) *DefaultMFAManager {
	if config == nil {
		config = DefaultMFAManagerConfig()
	}
	
	return &DefaultMFAManager{
		store:                  store,
		totpConfig:             config.TOTPConfig,
		backupConfig:           config.BackupConfig,
		webAuthnConfig:         config.WebAuthnConfig,
		smsConfig:              config.SMSConfig,
		verificationExpiration: config.VerificationExpiration,
	}
}

// GetMFASettings gets the MFA settings for a user
func (m *DefaultMFAManager) GetMFASettings(ctx context.Context, userID string) (*MFASettings, error) {
	return m.store.GetMFASettings(ctx, userID)
}

// EnableMFA enables MFA for a user
func (m *DefaultMFAManager) EnableMFA(ctx context.Context, userID string, method MFAMethod) (*MFASettings, error) {
	// Get existing settings or create new ones
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		// Create new settings
		settings = &MFASettings{
			UserID:       userID,
			Enabled:      false,
			DefaultMethod: method,
			Methods:      make(map[MFAMethod]MFAStatus),
		}
	}
	
	// Set method status to pending
	settings.Methods[method] = MFAStatusPending
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return nil, err
	}
	
	return settings, nil
}

// DisableMFA disables MFA for a user
func (m *DefaultMFAManager) DisableMFA(ctx context.Context, userID string) error {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return err
	}
	
	// Disable MFA
	settings.Enabled = false
	
	// Save settings
	return m.store.SaveMFASettings(ctx, settings)
}

// SetupTOTP sets up TOTP for a user
func (m *DefaultMFAManager) SetupTOTP(ctx context.Context, userID string, username string) (*TOTPConfig, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Check if TOTP is pending
	if settings.Methods[MFAMethodTOTP] != MFAStatusPending {
		return nil, errors.New("TOTP setup not initiated")
	}
	
	// Generate secret
	secret, err := GenerateSecret(20)
	if err != nil {
		return nil, err
	}
	
	// Create TOTP config
	totpConfig := &TOTPConfig{
		Secret:    secret,
		Digits:    m.totpConfig.Digits,
		Period:    m.totpConfig.Period,
		Algorithm: m.totpConfig.Algorithm,
		Issuer:    m.totpConfig.Issuer,
	}
	
	// Save config
	settings.TOTPConfig = totpConfig
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return nil, err
	}
	
	return totpConfig, nil
}

// VerifyTOTPSetup verifies TOTP setup
func (m *DefaultMFAManager) VerifyTOTPSetup(ctx context.Context, userID string, code string) error {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return err
	}
	
	// Check if TOTP is pending
	if settings.Methods[MFAMethodTOTP] != MFAStatusPending {
		return errors.New("TOTP setup not initiated")
	}
	
	// Verify code
	if !VerifyTOTPCode(settings.TOTPConfig, code, time.Now(), 1) {
		return errors.New("invalid TOTP code")
	}
	
	// Enable TOTP
	settings.Methods[MFAMethodTOTP] = MFAStatusEnabled
	settings.Enabled = true
	
	// Save settings
	return m.store.SaveMFASettings(ctx, settings)
}

// GenerateBackupCodes generates backup codes for a user
func (m *DefaultMFAManager) GenerateBackupCodes(ctx context.Context, userID string) ([]MFABackupCode, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Generate backup codes
	codes, err := GenerateBackupCodes(m.backupConfig)
	if err != nil {
		return nil, err
	}
	
	// Save backup codes
	settings.BackupCodes = codes
	settings.Methods[MFAMethodBackupCode] = MFAStatusEnabled
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return nil, err
	}
	
	return codes, nil
}

// SetupWebAuthn initiates WebAuthn setup
func (m *DefaultMFAManager) SetupWebAuthn(ctx context.Context, userID string, username string, displayName string) (map[string]interface{}, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Generate registration options
	options, challenge, err := RegistrationOptions(m.webAuthnConfig, userID, username, displayName)
	if err != nil {
		return nil, err
	}
	
	// Create verification
	verification := &MFAVerification{
		UserID:    userID,
		Method:    MFAMethodWebAuthn,
		Challenge: challenge,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(m.verificationExpiration) * time.Minute),
	}
	
	// Save verification
	if err := m.store.CreateVerification(ctx, verification); err != nil {
		return nil, err
	}
	
	// Set method status to pending
	settings.Methods[MFAMethodWebAuthn] = MFAStatusPending
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return nil, err
	}
	
	return options, nil
}

// VerifyWebAuthnSetup verifies WebAuthn setup
func (m *DefaultMFAManager) VerifyWebAuthnSetup(ctx context.Context, userID string, attestationResponse string) error {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return err
	}
	
	// Check if WebAuthn is pending
	if settings.Methods[MFAMethodWebAuthn] != MFAStatusPending {
		return errors.New("WebAuthn setup not initiated")
	}
	
	// Get verification
	verification, err := m.store.GetVerification(ctx, userID)
	if err != nil {
		return err
	}
	
	// Verify attestation
	credential, err := VerifyWebAuthnRegistration(m.webAuthnConfig, verification.Challenge, attestationResponse)
	if err != nil {
		return err
	}
	
	// Save credential
	if settings.WebAuthnCredentials == nil {
		settings.WebAuthnCredentials = make([]WebAuthnCredential, 0)
	}
	settings.WebAuthnCredentials = append(settings.WebAuthnCredentials, *credential)
	
	// Enable WebAuthn
	settings.Methods[MFAMethodWebAuthn] = MFAStatusEnabled
	settings.Enabled = true
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return err
	}
	
	// Delete verification
	return m.store.DeleteVerification(ctx, userID)
}

// SetupSMS initiates SMS setup
func (m *DefaultMFAManager) SetupSMS(ctx context.Context, userID string, phoneNumber string) error {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return err
	}
	
	// Create SMS verification
	smsVerification, err := CreateSMSVerification(m.smsConfig, phoneNumber, 3)
	if err != nil {
		return err
	}
	
	// Create verification
	verification := &MFAVerification{
		UserID:          userID,
		Method:          MFAMethodSMS,
		SMSVerification: smsVerification,
		CreatedAt:       time.Now(),
		ExpiresAt:       smsVerification.ExpiresAt,
	}
	
	// Save verification
	if err := m.store.CreateVerification(ctx, verification); err != nil {
		return err
	}
	
	// Set method status to pending
	settings.Methods[MFAMethodSMS] = MFAStatusPending
	settings.PhoneNumber = phoneNumber
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return err
	}
	
	// Send SMS code
	return SendSMSCode(m.smsConfig, smsVerification)
}

// VerifySMSSetup verifies SMS setup
func (m *DefaultMFAManager) VerifySMSSetup(ctx context.Context, userID string, code string) error {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return err
	}
	
	// Check if SMS is pending
	if settings.Methods[MFAMethodSMS] != MFAStatusPending {
		return errors.New("SMS setup not initiated")
	}
	
	// Get verification
	verification, err := m.store.GetVerification(ctx, userID)
	if err != nil {
		return err
	}
	
	// Verify SMS code
	verified, err := VerifySMSCode(verification.SMSVerification, code)
	if err != nil {
		return err
	}
	
	if !verified {
		return errors.New("SMS verification failed")
	}
	
	// Enable SMS
	settings.Methods[MFAMethodSMS] = MFAStatusEnabled
	settings.Enabled = true
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return err
	}
	
	// Delete verification
	return m.store.DeleteVerification(ctx, userID)
}

// VerifyMFA verifies an MFA code
func (m *DefaultMFAManager) VerifyMFA(ctx context.Context, userID string, method MFAMethod, code string) (bool, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return false, err
	}
	
	// Check if MFA is enabled
	if !settings.Enabled {
		return false, errors.New("MFA not enabled")
	}
	
	// Check if method is enabled
	if settings.Methods[method] != MFAStatusEnabled {
		return false, fmt.Errorf("MFA method %s not enabled", method)
	}
	
	// Verify based on method
	switch method {
	case MFAMethodTOTP:
		return VerifyTOTPCode(settings.TOTPConfig, code, time.Now(), 1), nil
	case MFAMethodBackupCode:
		verified, index, err := VerifyBackupCode(code, settings.BackupCodes)
		if err != nil {
			return false, err
		}
		
		// Mark backup code as used
		if err := MarkBackupCodeAsUsed(settings.BackupCodes, index); err != nil {
			return false, err
		}
		
		// Save settings
		if err := m.store.SaveMFASettings(ctx, settings); err != nil {
			return false, err
		}
		
		return verified, nil
	case MFAMethodSMS:
		// For SMS, we need to initiate a new verification
		return false, errors.New("SMS verification must be initiated first")
	case MFAMethodWebAuthn:
		// For WebAuthn, we need to initiate a new verification
		return false, errors.New("WebAuthn verification must be initiated first")
	default:
		return false, fmt.Errorf("unsupported MFA method: %s", method)
	}
}

// InitiateSMSVerification initiates SMS verification
func (m *DefaultMFAManager) InitiateSMSVerification(ctx context.Context, userID string) error {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return err
	}
	
	// Check if SMS is enabled
	if settings.Methods[MFAMethodSMS] != MFAStatusEnabled {
		return errors.New("SMS not enabled")
	}
	
	// Create SMS verification
	smsVerification, err := CreateSMSVerification(m.smsConfig, settings.PhoneNumber, 3)
	if err != nil {
		return err
	}
	
	// Create verification
	verification := &MFAVerification{
		UserID:          userID,
		Method:          MFAMethodSMS,
		SMSVerification: smsVerification,
		CreatedAt:       time.Now(),
		ExpiresAt:       smsVerification.ExpiresAt,
	}
	
	// Save verification
	if err := m.store.CreateVerification(ctx, verification); err != nil {
		return err
	}
	
	// Send SMS code
	return SendSMSCode(m.smsConfig, smsVerification)
}

// VerifySMSCode verifies an SMS code
func (m *DefaultMFAManager) VerifySMSCode(ctx context.Context, userID string, code string) (bool, error) {
	// Get verification
	verification, err := m.store.GetVerification(ctx, userID)
	if err != nil {
		return false, err
	}
	
	// Check if method is SMS
	if verification.Method != MFAMethodSMS {
		return false, errors.New("no SMS verification in progress")
	}
	
	// Verify SMS code
	verified, err := VerifySMSCode(verification.SMSVerification, code)
	if err != nil {
		return false, err
	}
	
	// Delete verification
	if err := m.store.DeleteVerification(ctx, userID); err != nil {
		return false, err
	}
	
	return verified, nil
}

// InitiateWebAuthnVerification initiates WebAuthn verification
func (m *DefaultMFAManager) InitiateWebAuthnVerification(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Check if WebAuthn is enabled
	if settings.Methods[MFAMethodWebAuthn] != MFAStatusEnabled {
		return nil, errors.New("WebAuthn not enabled")
	}
	
	// Generate authentication options
	options, challenge, err := AuthenticationOptions(m.webAuthnConfig, settings.WebAuthnCredentials)
	if err != nil {
		return nil, err
	}
	
	// Create verification
	verification := &MFAVerification{
		UserID:    userID,
		Method:    MFAMethodWebAuthn,
		Challenge: challenge,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(m.verificationExpiration) * time.Minute),
	}
	
	// Save verification
	if err := m.store.CreateVerification(ctx, verification); err != nil {
		return nil, err
	}
	
	return options, nil
}

// VerifyWebAuthnAssertion verifies a WebAuthn assertion
func (m *DefaultMFAManager) VerifyWebAuthnAssertion(ctx context.Context, userID string, assertionResponse string) (bool, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return false, err
	}
	
	// Get verification
	verification, err := m.store.GetVerification(ctx, userID)
	if err != nil {
		return false, err
	}
	
	// Check if method is WebAuthn
	if verification.Method != MFAMethodWebAuthn {
		return false, errors.New("no WebAuthn verification in progress")
	}
	
	// Find credential
	// In a real implementation, you would extract the credential ID from the assertion
	// and find the matching credential
	if len(settings.WebAuthnCredentials) == 0 {
		return false, errors.New("no WebAuthn credentials found")
	}
	
	// Verify assertion
	err = VerifyWebAuthnAuthentication(m.webAuthnConfig, verification.Challenge, assertionResponse, &settings.WebAuthnCredentials[0])
	if err != nil {
		return false, err
	}
	
	// Save settings
	if err := m.store.SaveMFASettings(ctx, settings); err != nil {
		return false, err
	}
	
	// Delete verification
	if err := m.store.DeleteVerification(ctx, userID); err != nil {
		return false, err
	}
	
	return true, nil
}

// ValidateMFASettings validates MFA settings
func (m *DefaultMFAManager) ValidateMFASettings(ctx context.Context, userID string) (bool, []string, error) {
	// Get settings
	settings, err := m.store.GetMFASettings(ctx, userID)
	if err != nil {
		return false, nil, err
	}
	
	// Check if MFA is enabled
	if !settings.Enabled {
		return true, nil, nil
	}
	
	// Validate settings
	valid := true
	var issues []string
	
	// Check if at least one method is enabled
	methodEnabled := false
	for _, status := range settings.Methods {
		if status == MFAStatusEnabled {
			methodEnabled = true
			break
		}
	}
	
	if !methodEnabled {
		valid = false
		issues = append(issues, "No MFA methods are enabled")
	}
	
	// Check TOTP configuration
	if settings.Methods[MFAMethodTOTP] == MFAStatusEnabled {
		if settings.TOTPConfig == nil || settings.TOTPConfig.Secret == "" {
			valid = false
			issues = append(issues, "TOTP is enabled but not configured properly")
		}
	}
	
	// Check backup codes
	if settings.Methods[MFAMethodBackupCode] == MFAStatusEnabled {
		if settings.BackupCodes == nil || len(settings.BackupCodes) == 0 {
			valid = false
			issues = append(issues, "Backup codes are enabled but not configured")
		} else {
			// Check if all backup codes are used
			allUsed := true
			for _, code := range settings.BackupCodes {
				if !code.Used {
					allUsed = false
					break
				}
			}
			
			if allUsed {
				issues = append(issues, "All backup codes have been used")
			}
		}
	}
	
	// Check WebAuthn credentials
	if settings.Methods[MFAMethodWebAuthn] == MFAStatusEnabled {
		if settings.WebAuthnCredentials == nil || len(settings.WebAuthnCredentials) == 0 {
			valid = false
			issues = append(issues, "WebAuthn is enabled but no credentials are registered")
		}
	}
	
	// Check SMS
	if settings.Methods[MFAMethodSMS] == MFAStatusEnabled {
		if settings.PhoneNumber == "" {
			valid = false
			issues = append(issues, "SMS is enabled but no phone number is configured")
		}
	}
	
	return valid, issues, nil
}
