package mfa

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// WebAuthnConfig represents the configuration for WebAuthn
type WebAuthnConfig struct {
	// RelyingPartyID is the ID of the relying party (typically the domain name)
	RelyingPartyID string
	
	// RelyingPartyName is the name of the relying party
	RelyingPartyName string
	
	// Origin is the origin of the relying party
	Origin string
	
	// UserVerification specifies the user verification requirement
	// Can be "required", "preferred", or "discouraged"
	UserVerification string
	
	// AttestationPreference specifies the attestation conveyance preference
	// Can be "none", "indirect", "direct", or "enterprise"
	AttestationPreference string
	
	// Timeout is the timeout for WebAuthn operations in milliseconds
	Timeout int
	
	// ChallengeLength is the length of the challenge in bytes
	ChallengeLength int
}

// WebAuthnCredential represents a WebAuthn credential
type WebAuthnCredential struct {
	// ID is the credential ID
	ID string
	
	// PublicKey is the public key of the credential
	PublicKey string
	
	// AAGUID is the Authenticator Attestation GUID
	AAGUID string
	
	// SignCount is the signature counter
	SignCount uint32
	
	// CreatedAt is the time when the credential was created
	CreatedAt time.Time
	
	// LastUsedAt is the time when the credential was last used
	LastUsedAt time.Time
	
	// DeviceType is the type of device (e.g., "security key", "platform")
	DeviceType string
	
	// DeviceName is a user-friendly name for the device
	DeviceName string
}

// DefaultWebAuthnConfig returns the default WebAuthn configuration
func DefaultWebAuthnConfig() *WebAuthnConfig {
	return &WebAuthnConfig{
		RelyingPartyID:        "LLMrecon.example.com",
		RelyingPartyName:      "LLMrecon",
		Origin:                "https://LLMrecon.example.com",
		UserVerification:      "preferred",
		AttestationPreference: "direct",
		Timeout:               60000,
		ChallengeLength:       32,
	}
}

// GenerateChallenge generates a random challenge for WebAuthn
func GenerateChallenge(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// RegistrationOptions generates the options for WebAuthn registration
func RegistrationOptions(config *WebAuthnConfig, userID, username, displayName string) (map[string]interface{}, string, error) {
	if config == nil {
		config = DefaultWebAuthnConfig()
	}
	
	// Generate challenge
	challenge, err := GenerateChallenge(config.ChallengeLength)
	if err != nil {
		return nil, "", err
	}
	
	// Create registration options
	options := map[string]interface{}{
		"rp": map[string]interface{}{
			"id":   config.RelyingPartyID,
			"name": config.RelyingPartyName,
		},
		"user": map[string]interface{}{
			"id":          base64.RawURLEncoding.EncodeToString([]byte(userID)),
			"name":        username,
			"displayName": displayName,
		},
		"challenge": challenge,
		"pubKeyCredParams": []map[string]interface{}{
			{
				"type": "public-key",
				"alg":  -7, // ES256
			},
			{
				"type": "public-key",
				"alg":  -257, // RS256
			},
		},
		"timeout":              config.Timeout,
		"attestation":          config.AttestationPreference,
		"authenticatorSelection": map[string]interface{}{
			"userVerification": config.UserVerification,
		},
	}
	
	return options, challenge, nil
}

// AuthenticationOptions generates the options for WebAuthn authentication
func AuthenticationOptions(config *WebAuthnConfig, credentials []WebAuthnCredential) (map[string]interface{}, string, error) {
	if config == nil {
		config = DefaultWebAuthnConfig()
	}
	
	// Generate challenge
	challenge, err := GenerateChallenge(config.ChallengeLength)
	if err != nil {
		return nil, "", err
	}
	
	// Create allowCredentials list
	allowCredentials := make([]map[string]interface{}, len(credentials))
	for i, cred := range credentials {
		allowCredentials[i] = map[string]interface{}{
			"type": "public-key",
			"id":   cred.ID,
		}
	}
	
	// Create authentication options
	options := map[string]interface{}{
		"challenge":        challenge,
		"timeout":          config.Timeout,
		"rpId":             config.RelyingPartyID,
		"allowCredentials": allowCredentials,
		"userVerification": config.UserVerification,
	}
	
	return options, challenge, nil
}

// This is a placeholder for a real WebAuthn verification implementation
// In a real implementation, you would use a WebAuthn library to verify the attestation
// and assertion responses from the client
func VerifyWebAuthnRegistration(config *WebAuthnConfig, challenge string, attestationResponse string) (*WebAuthnCredential, error) {
	if config == nil {
		config = DefaultWebAuthnConfig()
	}
	
	// Parse attestation response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(attestationResponse), &response); err != nil {
		return nil, err
	}
	
	// Verify challenge
	responseChallenge, ok := response["challenge"].(string)
	if !ok || responseChallenge != challenge {
		return nil, errors.New("invalid challenge")
	}
	
	// In a real implementation, you would verify the attestation statement
	// and extract the credential ID and public key
	
	// For this placeholder, we'll just create a dummy credential
	credential := &WebAuthnCredential{
		ID:         fmt.Sprintf("credential-%d", time.Now().Unix()),
		PublicKey:  "dummy-public-key",
		AAGUID:     "00000000-0000-0000-0000-000000000000",
		SignCount:  0,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		DeviceType: "security key",
		DeviceName: "Security Key",
	}
	
	return credential, nil
}

// This is a placeholder for a real WebAuthn verification implementation
// In a real implementation, you would use a WebAuthn library to verify the assertion
func VerifyWebAuthnAuthentication(config *WebAuthnConfig, challenge string, assertionResponse string, credential *WebAuthnCredential) error {
	if config == nil {
		config = DefaultWebAuthnConfig()
	}
	
	// Parse assertion response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(assertionResponse), &response); err != nil {
		return err
	}
	
	// Verify challenge
	responseChallenge, ok := response["challenge"].(string)
	if !ok || responseChallenge != challenge {
		return errors.New("invalid challenge")
	}
	
	// In a real implementation, you would verify the signature
	// and update the signature counter
	
	// Update last used time
	credential.LastUsedAt = time.Now()
	
	return nil
}
