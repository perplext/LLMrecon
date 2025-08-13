package mfa

import (
	"context"
)

// MFAManager defines the interface for managing multi-factor authentication
type MFAManager interface {
	// Settings management
	GetMFASettings(ctx context.Context, userID string) (*MFASettings, error)
	EnableMFA(ctx context.Context, userID string, method MFAMethod) (*MFASettings, error)
	DisableMFA(ctx context.Context, userID string) error
	
	// TOTP methods
	SetupTOTP(ctx context.Context, userID string, username string) (*TOTPConfig, error)
	VerifyTOTPSetup(ctx context.Context, userID string, code string) error
	
	// Backup code methods
	GenerateBackupCodes(ctx context.Context, userID string) ([]MFABackupCode, error)
	
	// WebAuthn methods
	SetupWebAuthn(ctx context.Context, userID string, username string, displayName string) (map[string]interface{}, error)
	VerifyWebAuthnSetup(ctx context.Context, userID string, attestationResponse string) error
	InitiateWebAuthnVerification(ctx context.Context, userID string) (map[string]interface{}, error)
	VerifyWebAuthnAssertion(ctx context.Context, userID string, assertionResponse string) (bool, error)
	
	// SMS methods
	SetupSMS(ctx context.Context, userID string, phoneNumber string) error
	VerifySMSSetup(ctx context.Context, userID string, code string) error
	InitiateSMSVerification(ctx context.Context, userID string) error
	VerifySMSCode(ctx context.Context, userID string, code string) (bool, error)
	
	// General MFA methods
	VerifyMFA(ctx context.Context, userID string, method MFAMethod, code string) (bool, error)
	ValidateMFASettings(ctx context.Context, userID string) (bool, []string, error)
}


// WebAuthnDevice represents a WebAuthn device
type WebAuthnDevice struct {
	ID        string
	Name      string
	CreatedAt time.Time
	LastUsed  time.Time
}

// TOTPConfig represents TOTP configuration
type TOTPConfig struct {
	Secret    string // Secret key for TOTP
	QRCodeURL string // QR code URL for easy setup
	Digits    int    // Number of digits in the TOTP code (default: 6)
	Period    int    // Period in seconds for TOTP (default: 30)
	Algorithm string // Algorithm for TOTP (default: SHA1)
	Issuer    string // Issuer name for TOTP (default: LLMrecon)
}

// SMSConfig represents the configuration for SMS verification
type SMSConfig struct {
	CodeLength        int               // Length of the SMS verification code
	CodeExpiration    int               // Expiration time for SMS codes in minutes
	SMSProvider       string            // SMS provider to use
	SMSProviderConfig map[string]string // Configuration for the SMS provider
}

// SMSVerification represents SMS verification
type SMSVerification struct {
	PhoneNumber string    // Phone number to send the verification code to
	Code        string    // Verification code
	CreatedAt   time.Time // Time when the verification was created
	ExpiresAt   time.Time // Time when the verification expires
	VerifiedAt  time.Time // Time when the verification was verified
	Verified    bool      // Whether the verification has been verified
	Attempts    int       // Number of verification attempts
	MaxAttempts int       // Maximum number of verification attempts
}

// MFAMethodSettings represents settings for an MFA method
type MFAMethodSettings struct {
	Method MFAMethod
	Status MFAStatus
	// Additional method-specific settings can be added here
}
