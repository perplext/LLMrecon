// Package vault provides a secure credential management system for the LLMreconing Tool.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	securityAudit "github.com/perplext/LLMrecon/src/security/audit"
	"golang.org/x/crypto/scrypt"
)

// CredentialType represents the type of credential
type CredentialType string

const (
	// APIKeyCredential represents an API key credential
	APIKeyCredential CredentialType = "api_key"
	// TokenCredential represents a token credential
	TokenCredential CredentialType = "token"
	// UsernamePasswordCredential represents a username/password credential
	UsernamePasswordCredential CredentialType = "username_password"
	// CertificateCredential represents a certificate credential
	CertificateCredential CredentialType = "certificate"
)

// RotationPolicy defines when credentials should be rotated
type RotationPolicy struct {
	// Enabled indicates whether automatic rotation is enabled
	Enabled bool `json:"enabled"`
	// IntervalDays is the number of days between rotations
	IntervalDays int `json:"interval_days"`
	// LastRotation is the timestamp of the last rotation
	LastRotation time.Time `json:"last_rotation"`
	// WarningDays is the number of days before expiration to start showing warnings
	WarningDays int `json:"warning_days"`
}

// Credential represents a secure credential
type Credential struct {
	// ID is a unique identifier for the credential
	ID string `json:"id"`
	// Name is a human-readable name for the credential
	Name string `json:"name"`
	// Type is the type of credential
	Type CredentialType `json:"type"`
	// Service is the service this credential is for (e.g., "openai", "anthropic")
	Service string `json:"service"`
	// Value is the actual credential value (encrypted when stored)
	Value string `json:"value"`
	// Description is an optional description of the credential
	Description string `json:"description,omitempty"`
	// Tags are optional tags for organizing credentials
	Tags []string `json:"tags,omitempty"`
	// Metadata is additional metadata for the credential
	Metadata map[string]string `json:"metadata,omitempty"`
	// RotationPolicy is the rotation policy for this credential
	RotationPolicy *RotationPolicy `json:"rotation_policy,omitempty"`
	// CreatedAt is when the credential was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the credential was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt is when the credential expires (if applicable)
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	// LastUsedAt is when the credential was last used
	LastUsedAt time.Time `json:"last_used_at,omitempty"`
}

// SecureVault manages secure storage of credentials
type SecureVault struct {
	// filePath is the path to the vault file
	filePath string
	// encryptionKey is the key used for encryption
	encryptionKey []byte
	// credentials is the in-memory cache of credentials
	credentials map[string]*Credential
	// mutex protects the credentials map
	mutex sync.RWMutex
	// auditLogger is used for logging credential access
	auditLogger *securityAudit.AuditLoggerAdapter
	// autoSave determines whether to automatically save after changes
	autoSave bool
	// rotationCheckInterval is how often to check for credentials that need rotation
	rotationCheckInterval time.Duration
	// rotationChecker is a ticker for checking rotation
	rotationChecker *time.Ticker
	// alertCallback is called when credentials need rotation
	alertCallback func(credential *Credential, daysUntilExpiration int)
}

// VaultOptions contains options for creating a new vault
type VaultOptions struct {
	// Passphrase is used to derive the encryption key
	Passphrase string
	// AuditLogger is used for logging credential access
	AuditLogger *securityAudit.AuditLoggerAdapter
	// AutoSave determines whether to automatically save after changes
	AutoSave bool
	// RotationCheckInterval is how often to check for credentials that need rotation
	RotationCheckInterval time.Duration
	// AlertCallback is called when a credential needs rotation
	AlertCallback func(credential *Credential, daysUntilExpiration int)
}

// NewSecureVault creates a new secure vault
func NewSecureVault(filePath string, options VaultOptions) (*SecureVault, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Derive encryption key from passphrase
	salt := []byte("LLMrecon-secure-vault-salt") // In a production system, this should be unique and stored securely
	key, err := deriveKey(options.Passphrase, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Set default rotation check interval if not specified
	rotationCheckInterval := options.RotationCheckInterval
	if rotationCheckInterval == 0 {
		rotationCheckInterval = 24 * time.Hour // Default to daily checks
	}

	vault := &SecureVault{
		filePath:              filePath,
		encryptionKey:         key,
		credentials:           make(map[string]*Credential),
		auditLogger:           options.AuditLogger,
		autoSave:              options.AutoSave,
		rotationCheckInterval: rotationCheckInterval,
		alertCallback:         options.AlertCallback,
	}

	// Load credentials from file if it exists
	if _, err := os.Stat(filePath); err == nil {
		if err := vault.load(); err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}
	}

	// Start rotation checker if interval is specified
	if rotationCheckInterval > 0 {
		vault.startRotationChecker()
	}

	return vault, nil
}

// startRotationChecker starts the rotation checker
func (v *SecureVault) startRotationChecker() {
	v.rotationChecker = time.NewTicker(v.rotationCheckInterval)
	go func() {
		for range v.rotationChecker.C {
			v.checkCredentialRotation()
		}
	}()
}

// checkCredentialRotation checks for credentials that need rotation
func (v *SecureVault) checkCredentialRotation() {
	// Make a copy of credentials under read lock to avoid long lock times
	v.mutex.RLock()
	credentialsCopy := make([]*Credential, 0, len(v.credentials))
	for _, cred := range v.credentials {
		// Only copy credentials with rotation policy to minimize memory usage
		if cred.RotationPolicy != nil && cred.RotationPolicy.Enabled {
			credCopy := *cred
			credentialsCopy = append(credentialsCopy, &credCopy)
		}
	}
	v.mutex.RUnlock() // Release lock before processing

	now := time.Now()
	for _, cred := range credentialsCopy {
		// Skip credentials without rotation policy
		if cred.RotationPolicy == nil || !cred.RotationPolicy.Enabled {
			continue
		}

		// Check if credential needs rotation
		var nextRotation time.Time
		if !cred.RotationPolicy.LastRotation.IsZero() {
			nextRotation = cred.RotationPolicy.LastRotation.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
		} else {
			nextRotation = cred.CreatedAt.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
		}

		// Calculate days until rotation
		daysUntilRotation := int(nextRotation.Sub(now).Hours() / 24)

		// Check if we need to alert
		if daysUntilRotation <= cred.RotationPolicy.WarningDays && v.alertCallback != nil {
			v.alertCallback(cred, daysUntilRotation)
		}
	}
}

// deriveKey derives an encryption key from a passphrase
func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
}

// encrypt encrypts data using AES-GCM
func (v *SecureVault) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(v.encryptionKey)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Encode as base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-GCM
func (v *SecureVault) decrypt(encryptedData string) ([]byte, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(v.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract the nonce
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	// Decrypt the data
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// load loads credentials from the file
func (v *SecureVault) load() error {
	// Read file
	data, err := os.ReadFile(v.filePath)
	if err != nil {
		return err
	}

	// Decrypt data
	decryptedData, err := v.decrypt(string(data))
	if err != nil {
		return err
	}

	// Parse JSON
	var storedCreds []*Credential
	if err := json.Unmarshal(decryptedData, &storedCreds); err != nil {
		return err
	}

	// Update in-memory cache
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.credentials = make(map[string]*Credential)
	for _, cred := range storedCreds {
		v.credentials[cred.ID] = cred
	}

	return nil
}

// Save saves credentials to the file
func (v *SecureVault) Save() error {
	// Make a copy of credentials under read lock
	v.mutex.RLock()
	var storedCreds []*Credential
	for _, cred := range v.credentials {
		// Create a deep copy to avoid race conditions
		credCopy := *cred
		storedCreds = append(storedCreds, &credCopy)
	}
	v.mutex.RUnlock() // Release lock before file operations

	// Convert to JSON
	data, err := json.Marshal(storedCreds)
	if err != nil {
		return err
	}

	// Encrypt data
	encryptedData, err := v.encrypt(data)
	if err != nil {
		return err
	}

	// Write to file with secure permissions
	return os.WriteFile(v.filePath, []byte(encryptedData), 0600)
}

// GetCredential gets a credential by ID
func (v *SecureVault) GetCredential(id string) (*Credential, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	cred, exists := v.credentials[id]
	if !exists {
		return nil, fmt.Errorf("credential with ID '%s' not found", id)
	}

	// Log access
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess(id, cred.Service, "get")
	}

	// Update last used timestamp
	cred.LastUsedAt = time.Now()
	if v.autoSave {
		go v.Save() // Save asynchronously to avoid blocking
	}

	return cred, nil
}

// StoreCredential stores a credential
func (v *SecureVault) StoreCredential(cred *Credential) error {
	if cred.ID == "" {
		return errors.New("credential ID cannot be empty")
	}

	// Variables to track state outside the lock
	var needToSave bool
	var isNew bool
	var credID, credService string

	// Critical section - keep it as short as possible
	{
		v.mutex.Lock()

		// Set timestamps
		now := time.Now()

		// Check if this is a new credential
		_, exists := v.credentials[cred.ID]
		if !exists {
			cred.CreatedAt = now
			isNew = true
		}
		cred.UpdatedAt = now

		// Store credential
		v.credentials[cred.ID] = cred

		// Save credential info for logging outside the lock
		credID = cred.ID
		credService = cred.Service
		needToSave = v.autoSave

		v.mutex.Unlock()
	}

	// Log operation outside the lock
	if v.auditLogger != nil {
		operation := "update"
		if isNew {
			operation = "create"
		}
		v.auditLogger.LogCredentialAccess(credID, credService, operation)
	}

	// Save to file if auto-save is enabled - outside the lock
	if needToSave {
		return v.Save()
	}

	return nil
}

// DeleteCredential deletes a credential by ID
func (v *SecureVault) DeleteCredential(id string) error {
	// Variables to track state outside the lock
	var needToSave bool
	var credFound bool
	var credService string

	// Critical section - keep it as short as possible
	{
		v.mutex.Lock()

		// Check if credential exists
		cred, exists := v.credentials[id]
		if !exists {
			v.mutex.Unlock()
			return fmt.Errorf("credential with ID '%s' not found", id)
		}

		// Save credential info for logging outside the lock
		credService = cred.Service
		credFound = true

		// Delete credential
		delete(v.credentials, id)

		// Check if we need to save
		needToSave = v.autoSave

		v.mutex.Unlock()
	}

	// Only proceed if credential was found
	if credFound {
		// Log operation outside the lock
		if v.auditLogger != nil {
			v.auditLogger.LogCredentialAccess(id, credService, "delete")
		}

		// Save to file if auto-save is enabled - outside the lock
		if needToSave {
			return v.Save()
		}
	}

	return nil
}

// ListCredentials lists all credentials
func (v *SecureVault) ListCredentials() ([]*Credential, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var credentials []*Credential
	for _, cred := range v.credentials {
		credentials = append(credentials, cred)
	}

	// Log operation
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess("all", "", "list")
	}

	return credentials, nil
}

// ListCredentialsByService lists credentials for a specific service
func (v *SecureVault) ListCredentialsByService(service string) ([]*Credential, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var credentials []*Credential
	for _, cred := range v.credentials {
		if cred.Service == service {
			credentials = append(credentials, cred)
		}
	}

	// Log operation
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess("service:"+service, service, "list")
	}

	return credentials, nil
}

// ListCredentialsByType lists credentials of a specific type
func (v *SecureVault) ListCredentialsByType(credType CredentialType) ([]*Credential, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var credentials []*Credential
	for _, cred := range v.credentials {
		if cred.Type == credType {
			credentials = append(credentials, cred)
		}
	}

	// Log operation
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess("type:"+string(credType), "", "list")
	}

	return credentials, nil
}

// ListCredentialsByTag lists credentials with a specific tag
func (v *SecureVault) ListCredentialsByTag(tag string) ([]*Credential, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var credentials []*Credential
	for _, cred := range v.credentials {
		for _, t := range cred.Tags {
			if t == tag {
				credentials = append(credentials, cred)
				break
			}
		}
	}

	// Log operation
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess("tag:"+tag, "", "list")
	}

	return credentials, nil
}

// RotateCredential marks a credential as rotated
func (v *SecureVault) RotateCredential(id string, newValue string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	// Check if credential exists
	cred, exists := v.credentials[id]
	if !exists {
		return fmt.Errorf("credential with ID '%s' not found", id)
	}

	// Update credential
	cred.Value = newValue
	cred.UpdatedAt = time.Now()
	if cred.RotationPolicy != nil {
		cred.RotationPolicy.LastRotation = time.Now()
	}

	// Log operation
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess(id, cred.Service, "rotate")
	}

	// Save to file if auto-save is enabled
	if v.autoSave {
		return v.Save()
	}

	return nil
}

// GetCredentialsNeedingRotation returns credentials that need rotation
func (v *SecureVault) GetCredentialsNeedingRotation() ([]*Credential, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	var credentials []*Credential
	now := time.Now()

	for _, cred := range v.credentials {
		// Skip credentials without rotation policy
		if cred.RotationPolicy == nil || !cred.RotationPolicy.Enabled {
			continue
		}

		// Calculate next rotation time
		var nextRotation time.Time
		if !cred.RotationPolicy.LastRotation.IsZero() {
			nextRotation = cred.RotationPolicy.LastRotation.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
		} else {
			nextRotation = cred.CreatedAt.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
		}

		// Check if credential needs rotation
		if now.After(nextRotation) {
			credentials = append(credentials, cred)
		}
	}

	// Log operation
	if v.auditLogger != nil {
		v.auditLogger.LogCredentialAccess("rotation-check", "", "list")
	}

	return credentials, nil
}

// Close closes the vault and stops any background processes
func (v *SecureVault) Close() error {
	if v.rotationChecker != nil {
		v.rotationChecker.Stop()
	}
	return v.Save()
}

// GenerateCredentialID generates a unique ID for a credential
func GenerateCredentialID(service string, name string) string {
	// Generate a hash of the service and name
	h := sha256.New()
	h.Write([]byte(service))
	h.Write([]byte(name))
	h.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	hash := h.Sum(nil)
	
	// Use first 8 bytes of hash as ID
	return fmt.Sprintf("%s-%s-%x", service, name, hash[:8])
}
