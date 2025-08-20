package mfa

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

const (
	// DefaultBackupCodeLength is the default length of each backup code
	DefaultBackupCodeLength = 8
	
	// DefaultBackupCodeCount is the default number of backup codes to generate
	DefaultBackupCodeCount = 10
)

// MFABackupCode represents a backup code for MFA
type MFABackupCode struct {
	// Code is the backup code
	Code string
	
	// Used indicates whether the code has been used
	Used bool
	
	// UsedAt is the time when the code was used
	UsedAt time.Time
}

// BackupCodeConfig represents the configuration for backup codes
type BackupCodeConfig struct {
	// CodeLength is the length of each backup code
	CodeLength int
	
	// CodeCount is the number of backup codes to generate
	CodeCount int
}

// DefaultBackupCodeConfig returns the default backup code configuration
func DefaultBackupCodeConfig() *BackupCodeConfig {
	return &BackupCodeConfig{
		CodeLength: DefaultBackupCodeLength,
		CodeCount:  DefaultBackupCodeCount,
	}
}

// GenerateBackupCodes generates a set of backup codes
func GenerateBackupCodes(config *BackupCodeConfig) ([]MFABackupCode, error) {
	if config == nil {
		config = DefaultBackupCodeConfig()
	}
	
	codes := make([]MFABackupCode, config.CodeCount)
	
	for i := 0; i < config.CodeCount; i++ {
		// Generate random bytes
		bytes := make([]byte, (config.CodeLength+1)/2) // +1 to handle odd lengths
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		
		// Encode as hex
		code := hex.EncodeToString(bytes)
		
		// Truncate to desired length
		if len(code) > config.CodeLength {
			code = code[:config.CodeLength]
		}
		
		// Format code with hyphen in the middle for readability
		if config.CodeLength >= 6 {
			midpoint := config.CodeLength / 2
			code = code[:midpoint] + "-" + code[midpoint:]
		}
		
		// Store code
		codes[i] = MFABackupCode{
			Code: strings.ToUpper(code),
			Used: false,
		}
	}
	
	return codes, nil
}

// VerifyBackupCode verifies a backup code
func VerifyBackupCode(providedCode string, storedCodes []MFABackupCode) (bool, int, error) {
	// Normalize provided code
	normalizedCode := strings.ToUpper(strings.ReplaceAll(providedCode, "-", ""))
	
	// Check each code
	for i, code := range storedCodes {
		// Skip used codes
		if code.Used {
			continue
		}
		
		// Normalize stored code
		normalizedStoredCode := strings.ToUpper(strings.ReplaceAll(code.Code, "-", ""))
		
		// Compare codes
		if normalizedCode == normalizedStoredCode {
			return true, i, nil
		}
	}
	
	return false, -1, errors.New("invalid backup code")
}

// MarkBackupCodeAsUsed marks a backup code as used
func MarkBackupCodeAsUsed(codes []MFABackupCode, index int) error {
	if index < 0 || index >= len(codes) {
		return errors.New("invalid backup code index")
	}
	
	// Mark code as used
	codes[index].Used = true
	codes[index].UsedAt = time.Now()
	
	return nil
}

// GetRemainingBackupCodes returns the number of remaining backup codes
func GetRemainingBackupCodes(codes []MFABackupCode) int {
	remaining := 0
	
	for _, code := range codes {
		if !code.Used {
			remaining++
		}
	}
	
	return remaining
}
