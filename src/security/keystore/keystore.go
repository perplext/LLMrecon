// Package keystore provides secure storage for cryptographic keys and sensitive materials.
package keystore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/google/uuid"
	securityAudit "github.com/perplext/LLMrecon/src/security/audit"
	"github.com/perplext/LLMrecon/src/security/vault"
)

// FileKeyStore implements the KeyStore interface with file-based storage
type FileKeyStore struct {
	// storagePath is the path to the key store file
	storagePath string
	
	// vault is the underlying secure vault for storing encrypted keys
	vault *vault.SecureVault
	
	// keys is the in-memory cache of keys
	keys map[string]*Key
	
	// mutex protects the keys map
	mutex sync.RWMutex
	
	// hsmManager is the HSM integration manager
	hsmManager *HSMManager
	
	// auditLogger is used for logging key access
	auditLogger *securityAudit.AuditLoggerAdapter
	
	// autoSave determines whether to automatically save after changes
	autoSave bool
	
	// rotationCheckInterval is how often to check for keys that need rotation
	rotationCheckInterval time.Duration
	
	// rotationChecker is a ticker for checking rotation
	rotationChecker *time.Ticker
	
	// alertCallback is called when keys need rotation
	alertCallback func(key *KeyMetadata, daysUntilExpiration int)
}

// NewFileKeyStore creates a new file-based key store
func NewFileKeyStore(options KeyStoreOptions) (*FileKeyStore, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(options.StoragePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create vault options
	vaultOptions := vault.VaultOptions{
		Passphrase:           options.Passphrase,
		AutoSave:             options.AutoSave,
		RotationCheckInterval: options.RotationCheckInterval,
	}

	// Create vault file path
	vaultPath := options.StoragePath + ".vault"

	// Create vault
	v, err := vault.NewSecureVault(vaultPath, vaultOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure vault: %w", err)
	}

	// Create HSM manager if configured
	var hsmManager *HSMManager
	if options.HSMConfig != nil && options.HSMConfig.Enabled {
		hsmManager, err = NewHSMManager(*options.HSMConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize HSM: %w", err)
		}
	}

	keyStore := &FileKeyStore{
		storagePath:          options.StoragePath,
		vault:                v,
		keys:                 make(map[string]*Key),
		hsmManager:           hsmManager,
		auditLogger:          nil, // TODO: Add audit logger
		autoSave:             options.AutoSave,
		rotationCheckInterval: options.RotationCheckInterval,
		alertCallback:        options.AlertCallback,
	}

	// Load keys from file if it exists
	if _, err := os.Stat(options.StoragePath); err == nil {
		if err := keyStore.load(); err != nil {
			return nil, fmt.Errorf("failed to load keys: %w", err)
		}
	}

	// Start rotation checker if interval is specified
	if options.RotationCheckInterval > 0 {
		keyStore.startRotationChecker()
	}

	return keyStore, nil
}

// load loads keys from the file
func (ks *FileKeyStore) load() error {
	// Read file
	data, err := ioutil.ReadFile(ks.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read key store file: %w", err)
	}

	// Parse JSON
	var keyMetadataList []KeyMetadata
	if err := json.Unmarshal(data, &keyMetadataList); err != nil {
		return fmt.Errorf("failed to parse key store file: %w", err)
	}

	// Load keys
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	for _, metadata := range keyMetadataList {
		// Skip keys that are stored in HSM
		if metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
			// Create key with metadata only
			ks.keys[metadata.ID] = &Key{
				Metadata: metadata,
			}
			continue
		}

		// Get key material from vault
		cred, err := ks.vault.GetCredential(metadata.ID)
		if err != nil {
			return fmt.Errorf("failed to get key material for %s: %w", metadata.ID, err)
		}

		// Parse key material
		var material KeyMaterial
		if err := json.Unmarshal([]byte(cred.Value), &material); err != nil {
			return fmt.Errorf("failed to parse key material for %s: %w", metadata.ID, err)
		}

		// Create key
		ks.keys[metadata.ID] = &Key{
			Metadata: metadata,
			Material: material,
		}
	}

	return nil
}

// save saves keys to the file
func (ks *FileKeyStore) save() error {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	// Extract metadata
	metadataList := make([]KeyMetadata, 0, len(ks.keys))
	for _, key := range ks.keys {
		metadataList = append(metadataList, key.Metadata)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(metadataList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key metadata: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(ks.storagePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write key store file: %w", err)
	}

	return nil
}

// startRotationChecker starts the rotation checker
func (ks *FileKeyStore) startRotationChecker() {
	ks.rotationChecker = time.NewTicker(ks.rotationCheckInterval)
	go func() {
		for range ks.rotationChecker.C {
			ks.checkKeyRotation()
		}
	}()
}

// checkKeyRotation checks for keys that need rotation
func (ks *FileKeyStore) checkKeyRotation() {
	// Make a copy of keys under read lock to avoid long lock times
	ks.mutex.RLock()
	keysCopy := make([]*KeyMetadata, 0, len(ks.keys))
	for _, key := range ks.keys {
		// Only copy keys with rotation period to minimize memory usage
		if key.Metadata.RotationPeriod > 0 {
			metadataCopy := key.Metadata
			keysCopy = append(keysCopy, &metadataCopy)
		}
	}
	ks.mutex.RUnlock() // Release lock before processing

	now := time.Now()
	for _, metadata := range keysCopy {
		// Skip keys without rotation period
		if metadata.RotationPeriod <= 0 {
			continue
		}

		// Check if key needs rotation
		var nextRotation time.Time
		if !metadata.LastRotatedAt.IsZero() {
			nextRotation = metadata.LastRotatedAt.AddDate(0, 0, metadata.RotationPeriod)
		} else {
			nextRotation = metadata.CreatedAt.AddDate(0, 0, metadata.RotationPeriod)
		}

		// Calculate days until rotation
		daysUntilRotation := int(nextRotation.Sub(now).Hours() / 24)

		// Alert if rotation is needed soon or overdue
		if daysUntilRotation <= 0 {
			// Key is due for rotation
			if ks.alertCallback != nil {
				ks.alertCallback(metadata, daysUntilRotation)
			}
		}
	}
}

// StoreKey stores a key in the key store
func (ks *FileKeyStore) StoreKey(key *Key) error {
	if key == nil {
		return errors.New("key cannot be nil")
	}

	// Generate ID if not provided
	if key.Metadata.ID == "" {
		key.Metadata.ID = generateKeyID(key.Metadata.Name, key.Metadata.Type)
	}

	// Set timestamps
	now := time.Now()
	if key.Metadata.CreatedAt.IsZero() {
		key.Metadata.CreatedAt = now
	}
	key.Metadata.UpdatedAt = now

	// Calculate fingerprint if not provided
	if key.Metadata.Fingerprint == "" && len(key.Material.Public) > 0 {
		fingerprint, err := calculateKeyFingerprint(key.Material.Public)
		if err == nil {
			key.Metadata.Fingerprint = fingerprint
		}
	}

	// Store key based on protection level
	if key.Metadata.ProtectionLevel == HSMProtection {
		if ks.hsmManager == nil {
			return errors.New("HSM protection requested but HSM is not configured")
		}

		// Store key in HSM
		if err := ks.hsmManager.StoreKey(key); err != nil {
			return fmt.Errorf("failed to store key in HSM: %w", err)
		}

		// Store metadata only in memory
		ks.mutex.Lock()
		ks.keys[key.Metadata.ID] = &Key{
			Metadata: key.Metadata,
		}
		ks.mutex.Unlock()
	} else {
		// Store key material in vault
		materialJSON, err := json.Marshal(key.Material)
		if err != nil {
			return fmt.Errorf("failed to marshal key material: %w", err)
		}

		cred := &vault.Credential{
			ID:        key.Metadata.ID,
			Name:      key.Metadata.Name,
			Type:      vault.CredentialType("key_" + string(key.Metadata.Type)),
			Service:   "keystore",
			Value:     string(materialJSON),
			Tags:      key.Metadata.Tags,
			CreatedAt: key.Metadata.CreatedAt,
			UpdatedAt: key.Metadata.UpdatedAt,
		}

		if err := ks.vault.StoreCredential(cred); err != nil {
			return fmt.Errorf("failed to store key material in vault: %w", err)
		}

		// Store key in memory
		ks.mutex.Lock()
		ks.keys[key.Metadata.ID] = key
		ks.mutex.Unlock()
	}

	// Save to file if auto-save is enabled
	if ks.autoSave {
		if err := ks.save(); err != nil {
			return fmt.Errorf("failed to save key store: %w", err)
		}
	}

	return nil
}

// GetKey retrieves a key by ID
func (ks *FileKeyStore) GetKey(id string) (*Key, error) {
	ks.mutex.RLock()
	key, exists := ks.keys[id]
	ks.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key not found: %s", id)
	}

	// If key is stored in HSM, retrieve it
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetKey(id)
	}

	// Update last used timestamp
	ks.mutex.Lock()
	ks.keys[id].Metadata.LastUsedAt = time.Now()
	ks.mutex.Unlock()

	// Return a copy of the key to prevent modification
	keyCopy := *key
	return &keyCopy, nil
}

// GetKeyMetadata retrieves key metadata by ID
func (ks *FileKeyStore) GetKeyMetadata(id string) (*KeyMetadata, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	key, exists := ks.keys[id]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", id)
	}

	// Return a copy of the metadata to prevent modification
	metadataCopy := key.Metadata
	return &metadataCopy, nil
}

// DeleteKey deletes a key by ID
func (ks *FileKeyStore) DeleteKey(id string) error {
	ks.mutex.Lock()
	key, exists := ks.keys[id]
	if !exists {
		ks.mutex.Unlock()
		return fmt.Errorf("key not found: %s", id)
	}

	// Delete key based on protection level
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		// Delete key from HSM
		if err := ks.hsmManager.DeleteKey(id); err != nil {
			ks.mutex.Unlock()
			return fmt.Errorf("failed to delete key from HSM: %w", err)
		}
	} else {
		// Delete key material from vault
		if err := ks.vault.DeleteCredential(id); err != nil {
			ks.mutex.Unlock()
			return fmt.Errorf("failed to delete key material from vault: %w", err)
		}
	}

	// Delete key from memory
	delete(ks.keys, id)
	ks.mutex.Unlock()

	// Save to file if auto-save is enabled
	if ks.autoSave {
		if err := ks.save(); err != nil {
			return fmt.Errorf("failed to save key store: %w", err)
		}
	}

	return nil
}

// ListKeys lists all keys in the key store
func (ks *FileKeyStore) ListKeys() ([]*KeyMetadata, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	keys := make([]*KeyMetadata, 0, len(ks.keys))
	for _, key := range ks.keys {
		// Return a copy of the metadata to prevent modification
		metadataCopy := key.Metadata
		keys = append(keys, &metadataCopy)
	}

	return keys, nil
}

// ListKeysByType lists keys of a specific type
func (ks *FileKeyStore) ListKeysByType(keyType KeyType) ([]*KeyMetadata, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	keys := make([]*KeyMetadata, 0)
	for _, key := range ks.keys {
		if key.Metadata.Type == keyType {
			// Return a copy of the metadata to prevent modification
			metadataCopy := key.Metadata
			keys = append(keys, &metadataCopy)
		}
	}

	return keys, nil
}

// ListKeysByUsage lists keys with a specific usage
func (ks *FileKeyStore) ListKeysByUsage(usage KeyUsage) ([]*KeyMetadata, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	keys := make([]*KeyMetadata, 0)
	for _, key := range ks.keys {
		if key.Metadata.Usage == usage {
			// Return a copy of the metadata to prevent modification
			metadataCopy := key.Metadata
			keys = append(keys, &metadataCopy)
		}
	}

	return keys, nil
}

// ListKeysByTag lists keys with a specific tag
func (ks *FileKeyStore) ListKeysByTag(tag string) ([]*KeyMetadata, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	keys := make([]*KeyMetadata, 0)
	for _, key := range ks.keys {
		for _, t := range key.Metadata.Tags {
			if t == tag {
				// Return a copy of the metadata to prevent modification
				metadataCopy := key.Metadata
				keys = append(keys, &metadataCopy)
				break
			}
		}
	}

	return keys, nil
}

// Close closes the key store
func (ks *FileKeyStore) Close() error {
	// Stop rotation checker
	if ks.rotationChecker != nil {
		ks.rotationChecker.Stop()
	}

	// Close HSM manager
	if ks.hsmManager != nil {
		if err := ks.hsmManager.Close(); err != nil {
			return fmt.Errorf("failed to close HSM manager: %w", err)
		}
	}

	// Save to file
	if err := ks.save(); err != nil {
		return fmt.Errorf("failed to save key store: %w", err)
	}

	// Close vault
	if err := ks.vault.Close(); err != nil {
		return fmt.Errorf("failed to close vault: %w", err)
	}

	return nil
}

// generateKeyID generates a unique ID for a key
func generateKeyID(name string, keyType KeyType) string {
	// Generate a UUID
	id := uuid.New().String()

	// Create a prefix from the name and type
	prefix := ""
	if name != "" {
		// Use first 8 characters of name (or less if name is shorter)
		if len(name) > 8 {
			prefix = name[:8]
		} else {
			prefix = name
		}
		prefix = strings.ToLower(strings.ReplaceAll(prefix, " ", "-"))
		prefix += "-"
	}

	// Add key type
	prefix += string(keyType) + "-"

	// Return prefix + first 8 characters of UUID
	return prefix + id[:8]
}

// calculateKeyFingerprint calculates a fingerprint for a key
func calculateKeyFingerprint(keyData []byte) (string, error) {
	// Calculate SHA-256 hash of key data
	hash := sha256.Sum256(keyData)

	// Return base64-encoded hash
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}
