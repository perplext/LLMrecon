// Package mfa provides multi-factor authentication functionality
package mfa

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// GenerateRandomSecret generates a random secret for TOTP
func GenerateRandomSecret(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode as base32 (RFC 4648)
	secret := base32.StdEncoding.EncodeToString(bytes)
	// Remove padding
	secret = strings.TrimRight(secret, "=")
	return secret, nil
}

// GenerateRandomCode generates a random numeric code of specified length
func GenerateRandomCode(length int) (string, error) {
	// Define the maximum value for each digit (10 for digits 0-9)
	max := big.NewInt(10)
	
	// Build the code digit by digit
	var codeBuilder strings.Builder
	for i := 0; i < length; i++ {
		// Generate a random digit (0-9)
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		
		// Append the digit to the code
		codeBuilder.WriteString(fmt.Sprintf("%d", n.Int64()))
	}
	
	return codeBuilder.String(), nil
}

// GenerateBackupCode generates a random backup code
func GenerateBackupCode() (string, error) {
	// Generate 10 random bytes
	bytes := make([]byte, 10)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	// Encode as base64
	code := base64.StdEncoding.EncodeToString(bytes)
	// Take first 10 characters and remove special characters
	code = strings.ReplaceAll(code[:10], "+", "A")
	code = strings.ReplaceAll(code, "/", "B")
	code = strings.ReplaceAll(code, "=", "C")
	
	// Format as XXXX-XXXX-XX
	return fmt.Sprintf("%s-%s-%s", code[:4], code[4:8], code[8:10]), nil
}

// GenerateTOTPQRCodeURL generates a URL for a TOTP QR code
func GenerateTOTPQRCodeURL(issuer, username, secret string) string {
	// Format according to the otpauth URL format
	// otpauth://totp/<issuer>:<username>?secret=<secret>&issuer=<issuer>&algorithm=SHA1&digits=6&period=30
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, username, secret, issuer)
}

// IsBackupCodeValid checks if a backup code is valid
func IsBackupCodeValid(inputCode string, storedCodes []MFABackupCode) (bool, int) {
	// Normalize input code
	normalizedInput := strings.ReplaceAll(inputCode, "-", "")
	normalizedInput = strings.ToUpper(normalizedInput)
	
	// Check against stored codes
	for i, code := range storedCodes {
		if !code.Used {
			normalizedStored := strings.ReplaceAll(code.Code, "-", "")
			normalizedStored = strings.ToUpper(normalizedStored)
			
			if normalizedInput == normalizedStored {
				return true, i
			}
		}
	}
	
	return false, -1
}

// IsVerificationExpired checks if a verification is expired
func IsVerificationExpired(verification *MFAVerification) bool {
	return verification == nil || time.Now().After(verification.ExpiresAt)
}

// FormatPhoneNumber formats a phone number for display
func FormatPhoneNumber(phoneNumber string) string {
	// Remove non-digit characters
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phoneNumber)
	
	// Format based on length
	switch len(digits) {
	case 10: // US number without country code
		return fmt.Sprintf("+1 (%s) %s-%s", digits[:3], digits[3:6], digits[6:])
	case 11: // US number with country code
		if digits[0] == '1' {
			return fmt.Sprintf("+%s (%s) %s-%s", digits[:1], digits[1:4], digits[4:7], digits[7:])
		}
		fallthrough
	default:
		// For other formats, just add + if it doesn't have one
		if !strings.HasPrefix(phoneNumber, "+") {
			return "+" + phoneNumber
		}
		return phoneNumber
	}
}

// MaskPhoneNumber masks a phone number for privacy
func MaskPhoneNumber(phoneNumber string) string {
	// Remove non-digit characters
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phoneNumber)
	
	// Mask based on length
	if len(digits) >= 10 {
		// Keep country code and last 4 digits
		lastFour := digits[len(digits)-4:]
		return fmt.Sprintf("********%s", lastFour)
	}
	
	// For shorter numbers, mask all but the last 2
	if len(digits) > 2 {
		lastTwo := digits[len(digits)-2:]
		return fmt.Sprintf("%s**", lastTwo)
	}
	
	// For very short numbers, just return ***
	return "***"
}
