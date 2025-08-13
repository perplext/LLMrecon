package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/scrypt"
)

// CredentialStore manages secure storage of credentials
type CredentialStore struct {
	// filePath is the path to the credential store file
	filePath string
	
	// encryptionKey is the key used for encryption
	encryptionKey []byte
	
	// credentials is the in-memory cache of credentials
	credentials map[string]*Credentials
	
	// mutex protects the credentials map
	mutex sync.RWMutex
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(filePath string, passphrase string) (*CredentialStore, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Derive encryption key from passphrase
	salt := []byte("LLMrecon-salt") // In a production system, this should be unique and stored securely
	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}
	
	store := &CredentialStore{
		filePath:      filePath,
		encryptionKey: key,
		credentials:   make(map[string]*Credentials),
	}
	
	// Load credentials from file if it exists
	if _, err := os.Stat(filePath); err == nil {
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}
	}
	
	return store, nil
}

// deriveKey derives an encryption key from a passphrase
func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
}

// encrypt encrypts data using AES-GCM
func (s *CredentialStore) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
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
func (s *CredentialStore) decrypt(encryptedData string) ([]byte, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	
	block, err := aes.NewCipher(s.encryptionKey)
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
func (s *CredentialStore) load() error {
	// Read file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	
	// Decrypt data
	decryptedData, err := s.decrypt(string(data))
	if err != nil {
		return err
	}
	
	// Parse JSON
	var storedCreds []*Credentials
	if err := json.Unmarshal(decryptedData, &storedCreds); err != nil {
		return err
	}
	
	// Update in-memory cache
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.credentials = make(map[string]*Credentials)
	for _, cred := range storedCreds {
		s.credentials[cred.ID] = cred
	}
	
	return nil
}

// save saves credentials to the file
func (s *CredentialStore) save() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Convert map to slice
	var storedCreds []*Credentials
	for _, cred := range s.credentials {
		storedCreds = append(storedCreds, cred)
	}
	
	// Convert to JSON
	data, err := json.Marshal(storedCreds)
	if err != nil {
		return err
	}
	
	// Encrypt data
	encryptedData, err := s.encrypt(data)
	if err != nil {
		return err
	}
	
	// Write to file with secure permissions
	return os.WriteFile(s.filePath, []byte(encryptedData), 0600)
}

// GetCredentials gets credentials by ID
func (s *CredentialStore) GetCredentials(id string) (*Credentials, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	cred, exists := s.credentials[id]
	if !exists {
		return nil, fmt.Errorf("credentials with ID '%s' not found", id)
	}
	
	return cred, nil
}

// SaveCredentials saves credentials
func (s *CredentialStore) SaveCredentials(creds *Credentials) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Set timestamps
	now := time.Now()
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = now
	}
	creds.UpdatedAt = now
	
	// Save to in-memory cache
	s.credentials[creds.ID] = creds
	
	// Save to file
	return s.save()
}

// DeleteCredentials deletes credentials by ID
func (s *CredentialStore) DeleteCredentials(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if credentials exist
	if _, exists := s.credentials[id]; !exists {
		return fmt.Errorf("credentials with ID '%s' not found", id)
	}
	
	// Delete from in-memory cache
	delete(s.credentials, id)
	
	// Save to file
	return s.save()
}

// ListCredentials lists all credentials
func (s *CredentialStore) ListCredentials() ([]*Credentials, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Convert map to slice
	var creds []*Credentials
	for _, cred := range s.credentials {
		creds = append(creds, cred)
	}
	
	return creds, nil
}

// UpdateLastUsed updates the last used timestamp for credentials
func (s *CredentialStore) UpdateLastUsed(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if credentials exist
	cred, exists := s.credentials[id]
	if !exists {
		return fmt.Errorf("credentials with ID '%s' not found", id)
	}
	
	// Update last used timestamp
	cred.LastUsedAt = time.Now()
	
	// Save to file
	return s.save()
}
