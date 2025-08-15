// Package update provides functionality for checking and applying updates
package update

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/perplext/LLMrecon/src/security/keystore"
	"github.com/perplext/LLMrecon/src/version"
)

// SecurityPolicy defines the minimum security requirements for cryptographic operations
type SecurityPolicy struct {
	// MinimumTLSVersion is the minimum allowed TLS version
	MinimumTLSVersion uint16 `json:"minimum_tls_version"`
	
	// AllowedCipherSuites is the list of allowed TLS cipher suites
	AllowedCipherSuites []uint16 `json:"allowed_cipher_suites"`
	
	// AllowedSignatureAlgorithms is the list of allowed signature algorithms
	AllowedSignatureAlgorithms []SignatureAlgorithm `json:"allowed_signature_algorithms"`
	
	// MinimumKeySize maps algorithm types to their minimum key sizes in bits
	MinimumKeySize map[string]int `json:"minimum_key_size"`
	
	// RequireCertificatePinning indicates whether certificate pinning is required
	RequireCertificatePinning bool `json:"require_certificate_pinning"`
	
	// RequireRevocationCheck indicates whether certificate revocation checking is required
	RequireRevocationCheck bool `json:"require_revocation_check"`
	
	// MinimumVersions maps component types to their minimum allowed versions
	MinimumVersions map[string]string `json:"minimum_versions"`
	
	// LastUpdateTime is when the policy was last updated
	LastUpdateTime time.Time `json:"last_update_time"`
	
	// PolicyVersion is the version of the security policy
	PolicyVersion string `json:"policy_version"`
	
	// PolicySignature is the cryptographic signature of the policy
	PolicySignature string `json:"policy_signature"`
	
	// RequireSignatureVerification indicates whether signature verification is required
	RequireSignatureVerification bool `json:"require_signature_verification"`

// DowngradeProtection provides protection against cryptographic downgrade attacks
type DowngradeProtection struct {
	// Policy is the current security policy
	Policy *SecurityPolicy
	
	// PolicyPath is the path to the security policy file
	PolicyPath string
	
	// Verifier is used to verify policy signatures
	Verifier *SignatureVerifier
	
	// KeyStore is used to store and retrieve cryptographic keys
	KeyStore *keystore.FileKeyStore

// NewDowngradeProtection creates a new DowngradeProtection instance
func NewDowngradeProtection(policyPath string, keyStore *keystore.FileKeyStore) (*DowngradeProtection, error) {
	dp := &DowngradeProtection{
		PolicyPath: policyPath,
		KeyStore:   keyStore,
	}

	// Load the security policy
	if err := dp.LoadPolicy(); err != nil {
		// If policy doesn't exist, create a default one
		if os.IsNotExist(err) {
			dp.Policy = DefaultSecurityPolicy()
			if err := dp.SavePolicy(); err != nil {
				return nil, fmt.Errorf("failed to save default security policy: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load security policy: %w", err)
		}
	}

	// Initialize the signature verifier for policy verification
	if dp.KeyStore != nil {
		// Try to get the policy verification key from the key store
		policyKey, err := dp.KeyStore.GetKeyMetadata("policy-verification-key")
		if err == nil && policyKey.HasPublicKey {
			// Get the public key
			publicKeyData, err := dp.KeyStore.ExportKey(policyKey.ID, "PEM", false)
			if err == nil {
				// Create the verifier
				dp.Verifier, err = NewSignatureVerifier(string(publicKeyData))
				if err != nil {
					return nil, fmt.Errorf("failed to create signature verifier: %w", err)
				}
			}
		}
	}

	return dp, nil

// DefaultSecurityPolicy returns the default security policy
func DefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		MinimumTLSVersion: tls.VersionTLS12,
		AllowedCipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		AllowedSignatureAlgorithms: []SignatureAlgorithm{
			Ed25519Algorithm,
			ECDSAAlgorithm,
			RSAAlgorithm,
		},
		MinimumKeySize: map[string]int{
			"rsa":      2048,
			"ecdsa":    256,
			"ed25519":  256,
			"symmetric": 256,
		},
		RequireCertificatePinning: true,
		RequireRevocationCheck:    true,
		MinimumVersions: map[string]string{
			"core":      "1.0.0",
			"templates": "1.0.0",
			"modules":   "1.0.0",
		},
		LastUpdateTime:             time.Now(),
		PolicyVersion:              "1.0.0",
		PolicySignature:            "",
		RequireSignatureVerification: true,
	}
// LoadPolicy loads the security policy from disk
func (dp *DowngradeProtection) LoadPolicy() error {
	// Check if policy file exists
	if _, err := os.Stat(dp.PolicyPath); os.IsNotExist(err) {
		return err
	}

	// Read the policy file
	policyData, err := ioutil.ReadFile(filepath.Clean(dp.PolicyPath))
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	// Parse the policy
	var policy SecurityPolicy
	if err := json.Unmarshal(policyData, &policy); err != nil {
		return fmt.Errorf("failed to parse policy file: %w", err)
	}
	// Verify the policy signature if a verifier is available
	if dp.Verifier != nil && policy.PolicySignature != "" {
		// Create a temporary copy of the policy without the signature
		tempPolicy := policy
		tempPolicy.PolicySignature = ""
		
		// Marshal the policy without the signature
		tempPolicyData, err := json.Marshal(tempPolicy)
		if err != nil {
			return fmt.Errorf("failed to marshal policy for signature verification: %w", err)
		}
		
		// Create a temporary file for verification
		tempFile := dp.PolicyPath + ".temp"
		if err := ioutil.WriteFile(tempFile, tempPolicyData, 0600); err != nil {
			return fmt.Errorf("failed to write temporary policy file: %w", err)
		}
		defer os.Remove(tempFile)
		
		// Verify the signature
		if err := dp.Verifier.VerifySignature(tempFile, policy.PolicySignature); err != nil {
			return fmt.Errorf("policy signature verification failed: %w", err)
		}
	}

	dp.Policy = &policy
	return nil

// SavePolicy saves the security policy to disk
func (dp *DowngradeProtection) SavePolicy() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dp.PolicyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update the last update time
	dp.Policy.LastUpdateTime = time.Now()

	// Marshal the policy
	policyData, err := json.MarshalIndent(dp.Policy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	// Write the policy file
	if err := ioutil.WriteFile(dp.PolicyPath, policyData, 0600); err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}

	return nil

// SignPolicy signs the security policy
func (dp *DowngradeProtection) SignPolicy(privateKeyID string) error {
	if dp.KeyStore == nil {
		return errors.New("key store is required for policy signing")
	}

	// Verify the private key exists
	_, err := dp.KeyStore.GetKey(privateKeyID)
	if err != nil {
		return fmt.Errorf("failed to get private key: %w", err)
	}
	// Export the private key
	privateKeyData, err := dp.KeyStore.ExportKey(privateKeyID, "PEM", true)
	if err != nil {
		return fmt.Errorf("failed to export private key: %w", err)
	}

	// Create a signature generator
	generator, err := NewSignatureGenerator(string(privateKeyData))
	if err != nil {
		return fmt.Errorf("failed to create signature generator: %w", err)
	}

	// Create a temporary copy of the policy without the signature
	tempPolicy := *dp.Policy
	tempPolicy.PolicySignature = ""
	
	// Marshal the policy without the signature
	tempPolicyData, err := json.MarshalIndent(tempPolicy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy for signing: %w", err)
	}
	
	// Create a temporary file for signing
	tempFile := dp.PolicyPath + ".temp"
	if err := ioutil.WriteFile(tempFile, tempPolicyData, 0600); err != nil {
		return fmt.Errorf("failed to write temporary policy file: %w", err)
	}
	defer os.Remove(tempFile)
	
	// Generate the signature
	signature, err := generator.GenerateSignature(tempFile)
	if err != nil {
		return fmt.Errorf("failed to generate policy signature: %w", err)
	}
	
	// Update the policy signature
	dp.Policy.PolicySignature = signature
	
	// Save the policy
	return dp.SavePolicy()

// ValidateConnectionSecurity validates that the connection security options meet the policy requirements
func (dp *DowngradeProtection) ValidateConnectionSecurity(options *ConnectionSecurityOptions) error {
	// Check TLS version
	if options.MinTLSVersion < dp.Policy.MinimumTLSVersion {
		return fmt.Errorf("TLS version %x is below the minimum required version %x", 
			options.MinTLSVersion, dp.Policy.MinimumTLSVersion)
	}

	// Check cipher suites if specified
	if options.CipherSuites != nil && len(options.CipherSuites) > 0 {
		// Check if all specified cipher suites are allowed
		for _, suite := range options.CipherSuites {
			allowed := false
			for _, allowedSuite := range dp.Policy.AllowedCipherSuites {
				if suite == allowedSuite {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("cipher suite %x is not allowed by the security policy", suite)
			}
		}
	} else {
		// If no cipher suites are specified, set them to the allowed ones
		options.CipherSuites = dp.Policy.AllowedCipherSuites
	}

	// Check certificate pinning
	if dp.Policy.RequireCertificatePinning && !options.EnableCertificatePinning {
		return errors.New("certificate pinning is required by the security policy")
	}

	// Check revocation checking
	if dp.Policy.RequireRevocationCheck && !options.CheckRevocation {
		return errors.New("certificate revocation checking is required by the security policy")
	}

	return nil

// ValidateSignatureAlgorithm validates that a signature algorithm meets the policy requirements
func (dp *DowngradeProtection) ValidateSignatureAlgorithm(algorithm SignatureAlgorithm) error {
	// Check if the algorithm is allowed
	allowed := false
	for _, allowedAlgorithm := range dp.Policy.AllowedSignatureAlgorithms {
		if algorithm == allowedAlgorithm {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("signature algorithm %s is not allowed by the security policy", algorithm)
	}
	return nil

// ValidateKeySize validates that a key size meets the policy requirements
func (dp *DowngradeProtection) ValidateKeySize(algorithm string, keySize int) error {
	// Check if the algorithm has a minimum key size requirement
	minSize, exists := dp.Policy.MinimumKeySize[strings.ToLower(algorithm)]
	if !exists {
		return fmt.Errorf("no minimum key size defined for algorithm %s", algorithm)
	}

	// Check if the key size meets the minimum requirement
	if keySize < minSize {
		return fmt.Errorf("key size %d is below the minimum required size %d for algorithm %s", 
			keySize, minSize, algorithm)
	}

	return nil

// ValidateVersion validates that a version meets the minimum version requirement
func (dp *DowngradeProtection) ValidateVersion(componentType, versionStr string) error {
	// Check if the component type has a minimum version requirement
	minVersionStr, exists := dp.Policy.MinimumVersions[componentType]
	if !exists {
		return fmt.Errorf("no minimum version defined for component type %s", componentType)
	}

	// Parse the versions
	ver, err := version.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("failed to parse version %s: %w", versionStr, err)
	}

	minVer, err := version.ParseVersion(minVersionStr)
	if err != nil {
		return fmt.Errorf("failed to parse minimum version %s: %w", minVersionStr, err)
	}

	// Compare the versions
	if ver.LessThan(&minVer) {
		return fmt.Errorf("version %s is below the minimum required version %s for component type %s", 
			versionStr, minVersionStr, componentType)
	}

	return nil

// ValidateUpdatePackage validates that an update package meets the security policy requirements
func (dp *DowngradeProtection) ValidateUpdatePackage(pkg *UpdatePackage) error {
	// Check if signature verification is required
	if dp.Policy.RequireSignatureVerification && pkg.Manifest.Signature == "" {
		return errors.New("update package signature is required by the security policy")
	}

	// Validate the binary version
	if pkg.Manifest.Components.Binary.Version != "" {
		if err := dp.ValidateVersion("binary", pkg.Manifest.Components.Binary.Version); err != nil {
			return err
		}
	}

	// Validate the templates version
	if pkg.Manifest.Components.Templates.Version != "" {
		if err := dp.ValidateVersion("templates", pkg.Manifest.Components.Templates.Version); err != nil {
			return err
		}
	}

	// Validate module versions
	for _, module := range pkg.Manifest.Components.Modules {
		if err := dp.ValidateVersion("modules", module.Version); err != nil {
			return fmt.Errorf("module %s: %w", module.ID, err)
		}
	}

	return nil

// EnforceSecurityPolicy applies the security policy to a connection security options object
func (dp *DowngradeProtection) EnforceSecurityPolicy(options *ConnectionSecurityOptions) {
	// Enforce minimum TLS version
	if options.MinTLSVersion < dp.Policy.MinimumTLSVersion {
		options.MinTLSVersion = dp.Policy.MinimumTLSVersion
	}

	// Enforce allowed cipher suites
	if options.CipherSuites == nil || len(options.CipherSuites) == 0 {
		options.CipherSuites = dp.Policy.AllowedCipherSuites
	} else {
		// Filter out disallowed cipher suites
		allowedSuites := make([]uint16, 0, len(options.CipherSuites))
		for _, suite := range options.CipherSuites {
			allowed := false
			for _, allowedSuite := range dp.Policy.AllowedCipherSuites {
				if suite == allowedSuite {
					allowed = true
					break
				}
			}
			if allowed {
				allowedSuites = append(allowedSuites, suite)
			}
		}
		options.CipherSuites = allowedSuites
	}

	// Enforce certificate pinning
	if dp.Policy.RequireCertificatePinning {
		options.EnableCertificatePinning = true
	}

	// Enforce revocation checking
	if dp.Policy.RequireRevocationCheck {
		options.CheckRevocation = true
	}

// UpdateMinimumVersion updates the minimum version requirement for a component type
func (dp *DowngradeProtection) UpdateMinimumVersion(componentType, versionStr string) error {
	// Parse the version to ensure it's valid
	_, err := version.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version %s: %w", versionStr, err)
	}

	// Update the minimum version
	dp.Policy.MinimumVersions[componentType] = versionStr

	// Save the policy
	return dp.SavePolicy()

// UpdateAllowedSignatureAlgorithms updates the allowed signature algorithms
func (dp *DowngradeProtection) UpdateAllowedSignatureAlgorithms(algorithms []SignatureAlgorithm) error {
	// Validate the algorithms
	for _, algorithm := range algorithms {
		switch algorithm {
		case Ed25519Algorithm, RSAAlgorithm, ECDSAAlgorithm:
			// Valid algorithm
		default:
			return fmt.Errorf("invalid signature algorithm: %s", algorithm)
		}
	}

	// Update the allowed algorithms
	dp.Policy.AllowedSignatureAlgorithms = algorithms

	// Save the policy
	return dp.SavePolicy()

// UpdateMinimumKeySize updates the minimum key size for an algorithm
func (dp *DowngradeProtection) UpdateMinimumKeySize(algorithm string, keySize int) error {
	// Validate the key size
	if keySize <= 0 {
		return fmt.Errorf("invalid key size: %d", keySize)
	}

	// Update the minimum key size
	dp.Policy.MinimumKeySize[strings.ToLower(algorithm)] = keySize

	// Save the policy
	return dp.SavePolicy()

// CreateSecureClient creates a secure HTTP client that complies with the security policy
func (dp *DowngradeProtection) CreateSecureClient() (*SecureClient, error) {
	// Create default options
	options := DefaultConnectionSecurityOptions()

	// Apply the security policy
	dp.EnforceSecurityPolicy(options)

	// Create the secure client
