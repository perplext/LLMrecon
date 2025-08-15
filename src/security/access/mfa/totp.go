package mfa

import (
	"os"
	"time"
	"crypto/sha1"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
)

const (
	// DefaultDigits is the default number of digits in a TOTP code
	DefaultDigits = 6
	
	// DefaultPeriod is the default period in seconds for TOTP
	DefaultPeriod = 30
	
	// DefaultAlgorithm is the default algorithm for TOTP
	DefaultAlgorithm = "SHA1"
	
	// DefaultIssuer is the default issuer for TOTP
	DefaultIssuer = "LLMrecon"
)

// DefaultTOTPConfig returns the default TOTP configuration
func DefaultTOTPConfig() *TOTPConfig {
	return &TOTPConfig{
		Digits:    DefaultDigits,
		Period:    DefaultPeriod,
		Algorithm: DefaultAlgorithm,
		Issuer:    DefaultIssuer,
	}

// GenerateSecret generates a new random secret for TOTP
func GenerateSecret(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {

		return "", err
	}
	
	// Encode as base32
	secret := base32.StdEncoding.EncodeToString(bytes)
	
	// Remove padding
	secret = strings.TrimRight(secret, "=")
	
	return secret, nil

// GenerateQRCodeURL generates a URL for a QR code for TOTP
func GenerateQRCodeURL(config *TOTPConfig, accountName string) string {
	// Create URL
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   fmt.Sprintf("%s:%s", url.PathEscape(config.Issuer), url.PathEscape(accountName)),
	}
	
	// Add query parameters
	q := u.Query()
	q.Set("secret", config.Secret)
	q.Set("issuer", config.Issuer)
	q.Set("algorithm", config.Algorithm)
	q.Set("digits", fmt.Sprintf("%d", config.Digits))
	q.Set("period", fmt.Sprintf("%d", config.Period))
	u.RawQuery = q.Encode()
	
	return u.String()

// GenerateTOTPCode generates a TOTP code for the given time
func GenerateTOTPCode(config *TOTPConfig, t time.Time) (string, error) {
	// Calculate counter
	counter := uint64(t.Unix() / int64(config.Period))
	
	// Generate HOTP code
	return generateHOTP(config, counter)

// VerifyTOTPCode verifies a TOTP code
func VerifyTOTPCode(config *TOTPConfig, code string, t time.Time, window int) bool {
	// Calculate counter
	counter := uint64(t.Unix() / int64(config.Period))
	
	// Check codes within window
	for i := -window; i <= window; i++ {
		c, err := generateHOTP(config, counter+uint64(i))
		if err != nil {
			continue
		}
		
		if c == code {
			return true
		}
	}
	
	return false

// generateHOTP generates an HOTP code
func generateHOTP(config *TOTPConfig, counter uint64) (string, error) {
	// Decode secret
	secret := strings.TrimRight(config.Secret, "=")
	missingPadding := len(secret) % 8
	if missingPadding > 0 {
		secret = secret + strings.Repeat("=", 8-missingPadding)
	}
	
	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}
	
	// Convert counter to bytes
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)
	
	// Calculate HMAC
	h := hmac.New(sha1.New, secretBytes)
	h.Write(counterBytes)
	hash := h.Sum(nil)
	
	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0F
	binary := ((int(hash[offset]) & 0x7F) << 24) |
		((int(hash[offset+1]) & 0xFF) << 16) |
		((int(hash[offset+2]) & 0xFF) << 8) |
		(int(hash[offset+3]) & 0xFF)
	
	// Generate code
	otp := binary % pow10(config.Digits)
	format := fmt.Sprintf("%%0%dd", config.Digits)
	return fmt.Sprintf(format, otp), nil

// pow10 returns 10^n
func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
