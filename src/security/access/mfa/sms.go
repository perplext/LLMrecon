package mfa

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	// DefaultSMSCodeLength is the default length of SMS verification codes
	DefaultSMSCodeLength = 6
	
	// DefaultSMSCodeExpiration is the default expiration time for SMS codes in minutes
	DefaultSMSCodeExpiration = 10
)

// DefaultSMSConfig returns the default SMS configuration
func DefaultSMSConfig() *SMSConfig {
	return &SMSConfig{
		CodeLength:     DefaultSMSCodeLength,
		CodeExpiration: DefaultSMSCodeExpiration,
		SMSProvider:    "mock", // Use a mock provider by default
		SMSProviderConfig: map[string]string{
			"from": "LLMrecon",
		},
	}
}

// GenerateSMSCode generates a random SMS verification code
func GenerateSMSCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("code length must be positive")
	}
	
	// Calculate the maximum value (10^length - 1)
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	max.Sub(max, big.NewInt(1))
	
	// Generate a random number between 0 and max
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	
	// Format the number with leading zeros
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

// CreateSMSVerification creates a new SMS verification
func CreateSMSVerification(config *SMSConfig, phoneNumber string, maxAttempts int) (*SMSVerification, error) {
	if config == nil {
		config = DefaultSMSConfig()
	}
	
	// Generate verification code
	code, err := GenerateSMSCode(config.CodeLength)
	if err != nil {
		return nil, err
	}
	
	// Create verification
	now := time.Now()
	verification := &SMSVerification{
		PhoneNumber: phoneNumber,
		Code:        code,
		CreatedAt:   now,
		ExpiresAt:   now.Add(time.Duration(config.CodeExpiration) * time.Minute),
		Verified:    false,
		Attempts:    0,
		MaxAttempts: maxAttempts,
	}
	
	return verification, nil
}

// SendSMSCode sends an SMS verification code
func SendSMSCode(config *SMSConfig, verification *SMSVerification) error {
	if config == nil {
		config = DefaultSMSConfig()
	}
	
	// Check if verification is valid
	if verification == nil {
		return errors.New("verification is nil")
	}
	
	// Check if verification has expired
	if time.Now().After(verification.ExpiresAt) {
		return errors.New("verification has expired")
	}
	
	// In a real implementation, you would use an SMS provider to send the code
	// For this placeholder, we'll just log the code
	fmt.Printf("[SMS] Sending code %s to %s\n", verification.Code, verification.PhoneNumber)
	
	return nil
}

// VerifySMSCode verifies an SMS verification code
func VerifySMSCode(verification *SMSVerification, code string) (bool, error) {
	// Check if verification is valid
	if verification == nil {
		return false, errors.New("verification is nil")
	}
	
	// Check if verification has expired
	if time.Now().After(verification.ExpiresAt) {
		return false, errors.New("verification has expired")
	}
	
	// Check if verification has been verified
	if verification.Verified {
		return false, errors.New("verification has already been verified")
	}
	
	// Check if maximum attempts have been reached
	if verification.Attempts >= verification.MaxAttempts {
		return false, errors.New("maximum verification attempts reached")
	}
	
	// Increment attempts
	verification.Attempts++
	
	// Check if code matches
	if verification.Code != code {
		return false, errors.New("invalid verification code")
	}
	
	// Mark as verified
	verification.Verified = true
	verification.VerifiedAt = time.Now()
	
	return true, nil
}

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	// SendSMS sends an SMS message
	SendSMS(to, from, message string) error
}

// MockSMSProvider is a mock implementation of SMSProvider for testing
type MockSMSProvider struct{}

// SendSMS sends an SMS message (mock implementation)
func (p *MockSMSProvider) SendSMS(to, from, message string) error {
	fmt.Printf("[MockSMS] From: %s, To: %s, Message: %s\n", from, to, message)
	return nil
}

// GetSMSProvider returns an SMS provider based on the configuration
func GetSMSProvider(config *SMSConfig) (SMSProvider, error) {
	if config == nil {
		config = DefaultSMSConfig()
	}
	
	switch config.SMSProvider {
	case "mock":
		return &MockSMSProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", config.SMSProvider)
	}
}
