package bundle

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
)

// SignatureVersion defines the current signature format version
const SignatureVersion = "1.0"

// BundleSignature represents a digital signature for a bundle
type BundleSignature struct {
	Version     string            `json:"version"`
	Algorithm   string            `json:"algorithm"`
	KeyID       string            `json:"keyId"`
	Timestamp   time.Time         `json:"timestamp"`
	ContentHash string            `json:"contentHash"`
	Signature   string            `json:"signature"`
	Metadata    SignatureMetadata `json:"metadata"`
}

// SignatureMetadata contains additional signature information
type SignatureMetadata struct {
	Signer      string   `json:"signer"`
	Environment string   `json:"environment"`
	BuildID     string   `json:"buildId"`
	Tags        []string `json:"tags"`

// FileHash represents a file and its hash
type FileHash struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
	Mode uint32 `json:"mode"`

// ContentManifest represents the content to be signed
type ContentManifest struct {
	Files []FileHash `json:"files"`
}

// SigningKey represents a key used for signing
type SigningKey struct {
	KeyID      string    `json:"keyId"`
	Algorithm  string    `json:"algorithm"`
	PublicKey  string    `json:"publicKey"`
	PrivateKey string    `json:"privateKey,omitempty"`
	Created    time.Time `json:"created"`
	Expires    time.Time `json:"expires"`
	Usage      []string  `json:"usage"`
}

// VerificationResult contains the result of signature verification
type VerificationResult struct {
	Valid      bool                   `json:"valid"`
	Timestamp  time.Time              `json:"timestamp"`
	KeyID      string                 `json:"keyId"`
	Signer     string                 `json:"signer"`
	Errors     []string               `json:"errors,omitempty"`
	Warnings   []string               `json:"warnings,omitempty"`
	Details    map[string]interface{} `json:"details"`
}

// Signer handles bundle signing operations
type Signer struct {
	privateKey ed25519.PrivateKey
	keyID      string
	metadata   SignatureMetadata
}

// NewSigner creates a new bundle signer
func NewSigner(privateKey ed25519.PrivateKey, keyID string, metadata SignatureMetadata) *Signer {
	return &Signer{
		privateKey: privateKey,
		keyID:      keyID,
		metadata:   metadata,
	}

// SignBundle creates a digital signature for the bundle
func (s *Signer) SignBundle(bundlePath string) (*BundleSignature, error) {
	// Calculate content hash
	contentHash, manifest, err := s.calculateBundleHash(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate bundle hash: %w", err)
	}

	// Create signature object
	sig := &BundleSignature{
		Version:     SignatureVersion,
		Algorithm:   "Ed25519",
		KeyID:       s.keyID,
		Timestamp:   time.Now().UTC(),
		ContentHash: contentHash,
		Metadata:    s.metadata,
	}

	// Sign the content
	message, err := sig.getSigningMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to create signing message: %w", err)
	}

	signature := ed25519.Sign(s.privateKey, message)
	sig.Signature = base64.URLEncoding.EncodeToString(signature)

	// Save manifest alongside signature
	if err := s.saveManifest(bundlePath, manifest); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	return sig, nil

// calculateBundleHash computes the hash of bundle contents
func (s *Signer) calculateBundleHash(bundlePath string) (string, *ContentManifest, error) {
	manifest := &ContentManifest{
		Files: []FileHash{},
	}

	// Walk through bundle files
	err := filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and signature files
		if info.IsDir() || filepath.Base(path) == "bundle.sig" {
			return nil
		}

		// Calculate file hash
		relPath, err := filepath.Rel(bundlePath, path)
		if err != nil {
			return err
		}

		hash, err := s.hashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", relPath, err)
		}

		manifest.Files = append(manifest.Files, FileHash{
			Path: relPath,
			Hash: hash,
			Size: info.Size(),
			Mode: uint32(info.Mode()),
		})

		return nil
	})

	if err != nil {
		return "", nil, err
	}
	// Sort files for deterministic ordering
	sort.Slice(manifest.Files, func(i, j int) bool {
		return manifest.Files[i].Path < manifest.Files[j].Path
	})

	// Create canonical JSON
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return "", nil, err
	}

	// Hash the manifest
	h := sha256.Sum256(manifestJSON)
	contentHash := base64.URLEncoding.EncodeToString(h[:])
	return contentHash, manifest, nil

// hashFile computes SHA-256 hash of a file
func (s *Signer) hashFile(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil

// saveManifest saves the content manifest
func (s *Signer) saveManifest(bundlePath string, manifest *ContentManifest) error {
	manifestPath := filepath.Join(bundlePath, "signatures", "manifest.json")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(manifestPath, data, 0600))

// getSigningMessage creates the message to be signed
func (b *BundleSignature) getSigningMessage() ([]byte, error) {
	// Create a deterministic representation
	msg := map[string]interface{}{
		"version":     b.Version,
		"algorithm":   b.Algorithm,
		"keyId":       b.KeyID,
		"timestamp":   b.Timestamp.Format(time.RFC3339),
		"contentHash": b.ContentHash,
		"metadata":    b.Metadata,
	}

	return json.Marshal(msg)

// Verifier handles bundle verification operations
type Verifier struct {
	publicKeys map[string]ed25519.PublicKey
}

// NewVerifier creates a new bundle verifier
func NewVerifier() *Verifier {
	return &Verifier{
		publicKeys: make(map[string]ed25519.PublicKey),
	}

// AddPublicKey adds a public key for verification
func (v *Verifier) AddPublicKey(keyID string, publicKey ed25519.PublicKey) {
	v.publicKeys[keyID] = publicKey

// VerifyBundle verifies a bundle's signature
func (v *Verifier) VerifyBundle(bundlePath string) (*VerificationResult, error) {
	result := &VerificationResult{
		Valid:     false,
		Timestamp: time.Now().UTC(),
		Details:   make(map[string]interface{}),
		Errors:    []string{},
		Warnings:  []string{},
	}

	// Load signature
	sigPath := filepath.Join(bundlePath, "signatures", "bundle.sig")
	sigData, err := os.ReadFile(filepath.Clean(sigPath))
	if err != nil {
		result.Errors = append(result.Errors, "signature file not found")
		return result, fmt.Errorf("failed to read signature: %w", err)
	}

	var sig BundleSignature
	if err := json.Unmarshal(sigData, &sig); err != nil {
		result.Errors = append(result.Errors, "invalid signature format")
		return result, fmt.Errorf("failed to parse signature: %w", err)
	}

	result.KeyID = sig.KeyID
	result.Signer = sig.Metadata.Signer

	// Get public key
	publicKey, exists := v.publicKeys[sig.KeyID]
	if !exists {
		result.Errors = append(result.Errors, fmt.Sprintf("unknown key ID: %s", sig.KeyID))
		return result, nil
	}

	// Verify timestamp
	age := time.Since(sig.Timestamp)
	if age > 90*24*time.Hour {
		result.Warnings = append(result.Warnings, fmt.Sprintf("signature is %d days old", int(age.Hours()/24)))
	}

	// Calculate current hash
	signer := &Signer{} // Just for hash calculation
	currentHash, _, err := signer.calculateBundleHash(bundlePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to calculate bundle hash: %v", err))
		return result, nil
	}

	// Verify content hasn't changed
	if currentHash != sig.ContentHash {
		result.Errors = append(result.Errors, "bundle content has been modified")
		return result, nil
	}

	// Verify signature
	sigBytes, err := base64.URLEncoding.DecodeString(sig.Signature)
	if err != nil {
		result.Errors = append(result.Errors, "invalid signature encoding")
		return result, nil
	}

	message, err := sig.getSigningMessage()
	if err != nil {
		result.Errors = append(result.Errors, "failed to recreate signing message")
		return result, nil
	}

	if !ed25519.Verify(publicKey, message, sigBytes) {
		result.Errors = append(result.Errors, "signature verification failed")
		return result, nil
	}

	// All checks passed
	result.Valid = true
	result.Details["signatureAge"] = age.String()
	result.Details["algorithm"] = sig.Algorithm
	result.Details["environment"] = sig.Metadata.Environment

	return result, nil
// GenerateSigningKeyPair generates a new Ed25519 key pair for signing
func GenerateSigningKeyPair(keyID string) (*SigningKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	now := time.Now().UTC()
	key := &SigningKey{
		KeyID:      keyID,
		Algorithm:  "Ed25519",
		PublicKey:  base64.URLEncoding.EncodeToString(publicKey),
		PrivateKey: base64.URLEncoding.EncodeToString(privateKey),
		Created:    now,
		Expires:    now.AddDate(1, 0, 0), // 1 year validity
		Usage:      []string{"bundle-signing"},
	}

	return key, nil

// SaveSignature saves a signature to the bundle
func SaveSignature(bundlePath string, sig *BundleSignature) error {
	sigPath := filepath.Join(bundlePath, "signatures", "bundle.sig")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(sigPath), 0700); err != nil {
		return fmt.Errorf("failed to create signatures directory: %w", err)
	}

	data, err := json.MarshalIndent(sig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal signature: %w", err)
	}

	return os.WriteFile(filepath.Clean(sigPath, data, 0600))

// LoadSignature loads a signature from the bundle
func LoadSignature(bundlePath string) (*BundleSignature, error) {
	sigPath := filepath.Join(bundlePath, "signatures", "bundle.sig")
	data, err := os.ReadFile(filepath.Clean(sigPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	var sig BundleSignature
	if err := json.Unmarshal(data, &sig); err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}

}
}
}
}
}
}
