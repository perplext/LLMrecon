// Package mfa provides multi-factor authentication functionality
package mfa

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
)

// Additional MFA errors specific to this file
var (
	ErrInvalidTOTP         = errors.New("invalid TOTP code")
	ErrInvalidBackupCode   = errors.New("invalid backup code")
	ErrNoBackupCodesLeft   = errors.New("no backup codes left")
	ErrBackupCodeUsed      = errors.New("backup code already used")
	ErrInvalidMFAMethod    = errors.New("invalid MFA method")
	ErrMFAVerificationFail = errors.New("MFA verification failed")
)

// Note: MFAMethod, MFAStatus, BackupCode, TOTPConfig, and MFAManager types are defined in other files to avoid duplicates

// BasicMFAService provides basic MFA functionality
type BasicMFAService struct {
	// Secret is the TOTP secret key
	Secret string `json:"secret"`
	
	// Algorithm is the TOTP algorithm (default: SHA1)
	Algorithm string `json:"algorithm"`
	
	// Digits is the number of digits in the TOTP code (default: 6)
	Digits int `json:"digits"`
	
	// Period is the TOTP period in seconds (default: 30)
	Period int `json:"period"`
	
	// Issuer is the name of the issuer for the TOTP
	Issuer string `json:"issuer"`
	
	// AccountName is the account name for the TOTP
	AccountName string `json:"account_name"`

// Note: MFAManager, MFAStore, DefaultMFAManager types and their implementations are defined in other files to avoid duplicates
// This file contains only utility functions and basic service definitions

// generateTOTPSecret generates a random TOTP secret
func generateTOTPSecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
