// Package config provides functionality for managing provider configurations.
package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// EncryptData encrypts data using AES-GCM
func EncryptData(plaintext []byte, passphrase string) ([]byte, error) {
	// Create a new AES cipher using the key
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode as base64
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
	base64.StdEncoding.Encode(encoded, ciphertext)

	return encoded, nil
}

// DecryptData decrypts data using AES-GCM
func DecryptData(ciphertext []byte, passphrase string) ([]byte, error) {
	// Decode from base64
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(ciphertext)))
	n, err := base64.StdEncoding.Decode(decoded, ciphertext)
	if err != nil {
		return nil, err
	}
	decoded = decoded[:n]

	// Create a new AES cipher using the key
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract the nonce and ciphertext
	nonce, ciphertextBytes := decoded[:nonceSize], decoded[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// deriveKey derives a 32-byte key from a passphrase using SHA-256
func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// GenerateEncryptionKey generates a random encryption key
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptConfig encrypts a configuration file
func EncryptConfig(inputFile, outputFile, passphrase string) error {
	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Encrypt data
	encryptedData, err := EncryptData(data, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputFile, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// DecryptConfig decrypts a configuration file
func DecryptConfig(inputFile, outputFile, passphrase string) error {
	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Decrypt data
	decryptedData, err := DecryptData(data, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputFile, decryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// UpdateEncrypt updates the encrypt function in the ConfigManager
func (m *ConfigManager) UpdateEncrypt() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// encryptData encrypts data using the manager's encryption key
	m.encryptData = func(data []byte) ([]byte, error) {
		// If no encryption key is provided, return the data as is
		if len(m.encryptionKey) == 0 {
			return data, nil
		}

		// Create a new AES cipher block
		block, err := aes.NewCipher(m.encryptionKey)
		if err != nil {
			return nil, err
		}

		// Create a new GCM cipher
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		// Create a nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		// Encrypt the data
		return gcm.Seal(nonce, nonce, data, nil), nil
	}

	// decryptData decrypts data using the manager's encryption key
	m.decryptData = func(data []byte) ([]byte, error) {
		// If no encryption key is provided, return the data as is
		if len(m.encryptionKey) == 0 {
			return data, nil
		}

		// Create a new AES cipher block
		block, err := aes.NewCipher(m.encryptionKey)
		if err != nil {
			return nil, err
		}

		// Create a new GCM cipher
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		// Check if the data is long enough
		if len(data) < gcm.NonceSize() {
			return nil, fmt.Errorf("ciphertext too short")
		}

		// Get the nonce and ciphertext
		nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

		// Decrypt the data
		return gcm.Open(nil, nonce, ciphertext, nil)
	}
}

// EncryptSensitiveData encrypts sensitive data in a provider configuration
func (m *ConfigManager) EncryptSensitiveData(config *core.ProviderConfig) error {
	// Encrypt API key
	if config.APIKey != "" {
		encryptedAPIKey, err := m.encryptData([]byte(config.APIKey))
		if err != nil {
			return fmt.Errorf("failed to encrypt API key: %w", err)
		}
		config.APIKey = string(encryptedAPIKey)
	}

	// Encrypt org ID
	if config.OrgID != "" {
		encryptedOrgID, err := m.encryptData([]byte(config.OrgID))
		if err != nil {
			return fmt.Errorf("failed to encrypt org ID: %w", err)
		}
		config.OrgID = string(encryptedOrgID)
	}

	return nil
}

// DecryptSensitiveData decrypts sensitive data in a provider configuration
func (m *ConfigManager) DecryptSensitiveData(config *core.ProviderConfig) error {
	// Decrypt API key
	if config.APIKey != "" {
		decryptedAPIKey, err := m.decryptData([]byte(config.APIKey))
		if err != nil {
			return fmt.Errorf("failed to decrypt API key: %w", err)
		}
		config.APIKey = string(decryptedAPIKey)
	}

	// Decrypt org ID
	if config.OrgID != "" {
		decryptedOrgID, err := m.decryptData([]byte(config.OrgID))
		if err != nil {
			return fmt.Errorf("failed to decrypt org ID: %w", err)
		}
		config.OrgID = string(decryptedOrgID)
	}

	return nil
}
