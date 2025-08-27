package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"time"
)

// ExportFormat represents the format for key export
type ExportFormat string

const (
	FormatPEM    ExportFormat = "pem"
	FormatJSON   ExportFormat = "json"
	FormatBinary ExportFormat = "binary"
)

// KeyExportOptions contains options for key export
type KeyExportOptions struct {
	Format     ExportFormat `json:"format"`
	Password   string       `json:"password,omitempty"`
	Encrypted  bool         `json:"encrypted"`
	Metadata   bool         `json:"metadata"`
	Expiration *time.Time   `json:"expiration,omitempty"`
}

// KeyImportOptions contains options for key import
type KeyImportOptions struct {
	Format      ExportFormat `json:"format"`
	Password    string       `json:"password,omitempty"`
	ValidateKey bool         `json:"validate_key"`
	Overwrite   bool         `json:"overwrite"`
}

// ExportedKey represents an exported key with metadata
type ExportedKey struct {
	ID          string            `json:"id"`
	Type        KeyType           `json:"type"`
	Algorithm   string            `json:"algorithm"`
	Data        string            `json:"data"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExportedAt  time.Time         `json:"exported_at"`
	Expiration  *time.Time        `json:"expiration,omitempty"`
	Fingerprint string            `json:"fingerprint"`
}

// KeyExporter handles key export operations
type KeyExporter struct {
	keystore *Keystore
}

// KeyImporter handles key import operations
type KeyImporter struct {
	keystore *Keystore
}

// NewKeyExporter creates a new key exporter
func NewKeyExporter(ks *Keystore) *KeyExporter {
	return &KeyExporter{keystore: ks}
}

// NewKeyImporter creates a new key importer
func NewKeyImporter(ks *Keystore) *KeyImporter {
	return &KeyImporter{keystore: ks}
}

// ExportKey exports a key with the specified options
func (ke *KeyExporter) ExportKey(keyID string, opts KeyExportOptions) (*ExportedKey, error) {
	if keyID == "" {
		return nil, errors.New("key ID cannot be empty")
	}

	// Retrieve the key from keystore
	keyEntry, exists := ke.keystore.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key with ID %s not found", keyID)
	}

	// Create exported key structure
	exported := &ExportedKey{
		ID:         keyID,
		Type:       keyEntry.Type,
		Algorithm:  keyEntry.Algorithm,
		CreatedAt:  keyEntry.CreatedAt,
		ExportedAt: time.Now(),
		Expiration: opts.Expiration,
	}

	// Add metadata if requested
	if opts.Metadata {
		exported.Metadata = make(map[string]string)
		for k, v := range keyEntry.Metadata {
			exported.Metadata[k] = v
		}
	}

	// Export key data based on format
	var err error
	switch opts.Format {
	case FormatPEM:
		exported.Data, err = ke.exportToPEM(keyEntry, opts.Password, opts.Encrypted)
	case FormatJSON:
		exported.Data, err = ke.exportToJSON(keyEntry, opts.Password, opts.Encrypted)
	case FormatBinary:
		exported.Data, err = ke.exportToBinary(keyEntry, opts.Password, opts.Encrypted)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", opts.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to export key: %w", err)
	}

	// Generate fingerprint
	exported.Fingerprint = ke.generateFingerprint(exported.Data)

	return exported, nil
}

// exportToPEM exports key to PEM format
func (ke *KeyExporter) exportToPEM(keyEntry *KeyEntry, password string, encrypted bool) (string, error) {
	var keyData []byte
	var blockType string

	switch keyEntry.Type {
	case KeyTypeRSA:
		if rsaKey, ok := keyEntry.Key.(*rsa.PrivateKey); ok {
			keyData = x509.MarshalPKCS1PrivateKey(rsaKey)
			blockType = "RSA PRIVATE KEY"
		} else {
			return "", errors.New("invalid RSA key")
		}
	case KeyTypeAES:
		if aesKey, ok := keyEntry.Key.([]byte); ok {
			keyData = aesKey
			blockType = "AES KEY"
		} else {
			return "", errors.New("invalid AES key")
		}
	default:
		return "", fmt.Errorf("unsupported key type for PEM export: %s", keyEntry.Type)
	}

	block := &pem.Block{
		Type:  blockType,
		Bytes: keyData,
	}

	if encrypted && password != "" {
		var err error
		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, keyData, []byte(password), x509.PEMCipherAES256)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt PEM block: %w", err)
		}
	}

	return string(pem.EncodeToMemory(block)), nil
}

// exportToJSON exports key to JSON format
func (ke *KeyExporter) exportToJSON(keyEntry *KeyEntry, password string, encrypted bool) (string, error) {
	data := map[string]interface{}{
		"type":       keyEntry.Type,
		"algorithm":  keyEntry.Algorithm,
		"created_at": keyEntry.CreatedAt,
	}

	var keyData []byte
	switch keyEntry.Type {
	case KeyTypeRSA:
		if rsaKey, ok := keyEntry.Key.(*rsa.PrivateKey); ok {
			keyData = x509.MarshalPKCS1PrivateKey(rsaKey)
		} else {
			return "", errors.New("invalid RSA key")
		}
	case KeyTypeAES:
		if aesKey, ok := keyEntry.Key.([]byte); ok {
			keyData = aesKey
		} else {
			return "", errors.New("invalid AES key")
		}
	default:
		return "", fmt.Errorf("unsupported key type: %s", keyEntry.Type)
	}

	if encrypted && password != "" {
		encryptedData, err := ke.encryptKeyData(keyData, password)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt key data: %w", err)
		}
		data["key"] = base64.StdEncoding.EncodeToString(encryptedData)
		data["encrypted"] = true
	} else {
		data["key"] = base64.StdEncoding.EncodeToString(keyData)
		data["encrypted"] = false
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}

// exportToBinary exports key to binary format
func (ke *KeyExporter) exportToBinary(keyEntry *KeyEntry, password string, encrypted bool) (string, error) {
	var keyData []byte

	switch keyEntry.Type {
	case KeyTypeRSA:
		if rsaKey, ok := keyEntry.Key.(*rsa.PrivateKey); ok {
			keyData = x509.MarshalPKCS1PrivateKey(rsaKey)
		} else {
			return "", errors.New("invalid RSA key")
		}
	case KeyTypeAES:
		if aesKey, ok := keyEntry.Key.([]byte); ok {
			keyData = aesKey
		} else {
			return "", errors.New("invalid AES key")
		}
	default:
		return "", fmt.Errorf("unsupported key type: %s", keyEntry.Type)
	}

	if encrypted && password != "" {
		encryptedData, err := ke.encryptKeyData(keyData, password)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt key data: %w", err)
		}
		keyData = encryptedData
	}

	return base64.StdEncoding.EncodeToString(keyData), nil
}

// encryptKeyData encrypts key data using AES-GCM
func (ke *KeyExporter) encryptKeyData(keyData []byte, password string) ([]byte, error) {
	// Derive key from password (simplified - in production use proper KDF)
	key := make([]byte, 32)
	copy(key, []byte(password))

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, keyData, nil)
	return ciphertext, nil
}

// generateFingerprint generates a fingerprint for the exported key
func (ke *KeyExporter) generateFingerprint(data string) string {
	// Simplified fingerprint generation
	hash := fmt.Sprintf("%x", []byte(data)[:10])
	return hash
}

// ImportKey imports a key from exported data
func (ki *KeyImporter) ImportKey(exportedKey *ExportedKey, opts KeyImportOptions) error {
	if exportedKey == nil {
		return errors.New("exported key cannot be nil")
	}

	if exportedKey.ID == "" {
		return errors.New("key ID cannot be empty")
	}

	// Check if key already exists
	if _, exists := ki.keystore.keys[exportedKey.ID]; exists && !opts.Overwrite {
		return fmt.Errorf("key with ID %s already exists", exportedKey.ID)
	}

	// Parse key data based on format
	var keyData interface{}
	var err error

	switch opts.Format {
	case FormatPEM:
		keyData, err = ki.importFromPEM(exportedKey.Data, opts.Password)
	case FormatJSON:
		keyData, err = ki.importFromJSON(exportedKey.Data, opts.Password)
	case FormatBinary:
		keyData, err = ki.importFromBinary(exportedKey.Data, opts.Password)
	default:
		return fmt.Errorf("unsupported import format: %s", opts.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to import key: %w", err)
	}

	// Validate key if requested
	if opts.ValidateKey {
		if err := ki.validateKey(keyData, exportedKey.Type); err != nil {
			return fmt.Errorf("key validation failed: %w", err)
		}
	}

	// Create key entry
	keyEntry := &KeyEntry{
		ID:        exportedKey.ID,
		Type:      exportedKey.Type,
		Algorithm: exportedKey.Algorithm,
		Key:       keyData,
		Metadata:  exportedKey.Metadata,
		CreatedAt: exportedKey.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Store in keystore
	ki.keystore.keys[exportedKey.ID] = keyEntry

	return nil
}

// importFromPEM imports key from PEM format
func (ki *KeyImporter) importFromPEM(data string, password string) (interface{}, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	var keyData []byte
	var err error

	if x509.IsEncryptedPEMBlock(block) {
		if password == "" {
			return nil, errors.New("password required for encrypted PEM block")
		}
		keyData, err = x509.DecryptPEMBlock(block, []byte(password))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt PEM block: %w", err)
		}
	} else {
		keyData = block.Bytes
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyData)
	case "AES KEY":
		return keyData, nil
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

// importFromJSON imports key from JSON format
func (ki *KeyImporter) importFromJSON(data string, password string) (interface{}, error) {
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	keyDataStr, ok := jsonData["key"].(string)
	if !ok {
		return nil, errors.New("invalid key data in JSON")
	}

	keyData, err := base64.StdEncoding.DecodeString(keyDataStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key data: %w", err)
	}

	if encrypted, ok := jsonData["encrypted"].(bool); ok && encrypted {
		if password == "" {
			return nil, errors.New("password required for encrypted key")
		}
		keyData, err = ki.decryptKeyData(keyData, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key data: %w", err)
		}
	}

	keyType, ok := jsonData["type"].(string)
	if !ok {
		return nil, errors.New("missing key type in JSON")
	}

	switch KeyType(keyType) {
	case KeyTypeRSA:
		return x509.ParsePKCS1PrivateKey(keyData)
	case KeyTypeAES:
		return keyData, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// importFromBinary imports key from binary format
func (ki *KeyImporter) importFromBinary(data string, password string) (interface{}, error) {
	keyData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	if password != "" {
		keyData, err = ki.decryptKeyData(keyData, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key data: %w", err)
		}
	}

	return keyData, nil
}

// decryptKeyData decrypts key data using AES-GCM
func (ki *KeyImporter) decryptKeyData(encryptedData []byte, password string) ([]byte, error) {
	// Derive key from password (simplified - in production use proper KDF)
	key := make([]byte, 32)
	copy(key, []byte(password))

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// validateKey validates the imported key
func (ki *KeyImporter) validateKey(keyData interface{}, keyType KeyType) error {
	switch keyType {
	case KeyTypeRSA:
		if rsaKey, ok := keyData.(*rsa.PrivateKey); ok {
			return rsaKey.Validate()
		}
		return errors.New("invalid RSA key format")
	case KeyTypeAES:
		if aesKey, ok := keyData.([]byte); ok {
			if len(aesKey) != 16 && len(aesKey) != 24 && len(aesKey) != 32 {
				return errors.New("invalid AES key size")
			}
			return nil
		}
		return errors.New("invalid AES key format")
	default:
		return fmt.Errorf("validation not supported for key type: %s", keyType)
	}
}
