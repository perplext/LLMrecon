// Example program demonstrating the usage of downgrade protection
package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	"github.com/perplext/LLMrecon/src/security/keystore"
	"github.com/perplext/LLMrecon/src/update"
)

func main() {
	// Create a directory for the example
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	exampleDir := filepath.Join(homeDir, ".LLMrecon", "downgrade-example")
	if err := os.MkdirAll(exampleDir, 0700); err != nil {
		log.Fatalf("Failed to create example directory: %v", err)
	}

	// Create paths for files
	policyPath := filepath.Join(exampleDir, "security_policy.json")
	keyStorePath := filepath.Join(exampleDir, "keystore.json")

	// Initialize key store
	log.Println("Initializing key store...")
	keyStore, err := initializeKeyStore(keyStorePath)
	if err != nil {
		log.Fatalf("Failed to initialize key store: %v", err)
	}
	defer func() { if err := keyStore.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Initialize downgrade protection
	log.Println("Initializing downgrade protection...")
	dp, err := update.NewDowngradeProtection(policyPath, keyStore)
	if err != nil {
		log.Fatalf("Failed to initialize downgrade protection: %v", err)
	}

	// Generate and store keys for policy signing and verification
	log.Println("Generating policy signing and verification keys...")
	signingKeyID, verificationKeyID, err := generatePolicyKeys(keyStore)
	if err != nil {
		log.Fatalf("Failed to generate policy keys: %v", err)
	}
	log.Printf("Generated signing key: %s", signingKeyID)
	log.Printf("Generated verification key: %s", verificationKeyID)

	// Sign the security policy
	log.Println("Signing security policy...")
	if err := dp.SignPolicy(signingKeyID); err != nil {
		log.Fatalf("Failed to sign policy: %v", err)
	}
	log.Println("Security policy signed successfully")

	// Display the current security policy
	log.Println("Current security policy:")
	displayPolicy(dp.Policy)

	// Demonstrate connection security validation
	log.Println("\nDemonstrating connection security validation...")
	demonstrateConnectionSecurityValidation(dp)

	// Demonstrate signature algorithm validation
	log.Println("\nDemonstrating signature algorithm validation...")
	demonstrateSignatureAlgorithmValidation(dp)

	// Demonstrate key size validation
	log.Println("\nDemonstrating key size validation...")
	demonstrateKeySizeValidation(dp)

	// Demonstrate version validation
	log.Println("\nDemonstrating version validation...")
	demonstrateVersionValidation(dp)

	// Demonstrate update package validation
	log.Println("\nDemonstrating update package validation...")
	demonstrateUpdatePackageValidation(dp)

	// Demonstrate policy enforcement
	log.Println("\nDemonstrating policy enforcement...")
	demonstratePolicyEnforcement(dp)

	// Demonstrate creating a secure client
	log.Println("\nDemonstrating secure client creation...")
	demonstrateSecureClientCreation(dp)

	// Update the security policy
	log.Println("\nUpdating security policy...")
	demonstratePolicyUpdate(dp)

	log.Println("\nDowngrade protection example completed successfully")

// initializeKeyStore initializes the key store
func initializeKeyStore(keyStorePath string) (*keystore.FileKeyStore, error) {
	keyStoreOptions := keystore.KeyStoreOptions{
		StoragePath:           keyStorePath,
		Passphrase:            "example-passphrase", // In production, use a secure passphrase
		AutoSave:              true,
		RotationCheckInterval: time.Hour * 24, // Check for rotation daily
	}

	return keystore.NewFileKeyStore(keyStoreOptions)

// generatePolicyKeys generates keys for policy signing and verification
func generatePolicyKeys(ks *keystore.FileKeyStore) (string, string, error) {
	// Generate Ed25519 key pair
	privateKeyPEM, publicKeyPEM, err := update.GenerateKeyPair(update.Ed25519Algorithm)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create metadata for the signing key
	signingMetadata := &keystore.KeyMetadata{
		Name:            "policy-signing-key",
		Type:            keystore.Ed25519Key,
		Usage:           keystore.SigningKey,
		ProtectionLevel: keystore.SoftwareProtection,
		Tags:            []string{"policy", "signing"},
		Description:     "Key for signing security policies",
		RotationPeriod:  365, // 1 year
	}

	// Import the signing key
	signingKey, err := ks.ImportKey([]byte(privateKeyPEM), "PEM", signingMetadata)
	if err != nil {
		return "", "", fmt.Errorf("failed to import signing key: %w", err)
	}
	// Create metadata for the verification key
	verificationMetadata := &keystore.KeyMetadata{
		Name:            "policy-verification-key",
		Type:            keystore.Ed25519Key,
		Usage:           keystore.SigningKey,
		ProtectionLevel: keystore.SoftwareProtection,
		Tags:            []string{"policy", "verification"},
		Description:     "Key for verifying security policy signatures",
		RotationPeriod:  365, // 1 year
	}

	// Import the verification key
	verificationKey, err := ks.ImportKey([]byte(publicKeyPEM), "PEM", verificationMetadata)
	if err != nil {
		return "", "", fmt.Errorf("failed to import verification key: %w", err)
	}

	return signingKey.Metadata.ID, verificationKey.Metadata.ID, nil

// displayPolicy displays the current security policy
func displayPolicy(policy *update.SecurityPolicy) {
	policyJSON, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal policy: %v", err)
		return
	}
	fmt.Println(string(policyJSON))

// demonstrateConnectionSecurityValidation demonstrates connection security validation
func demonstrateConnectionSecurityValidation(dp *update.DowngradeProtection) {
	// Create valid options
	validOptions := update.DefaultConnectionSecurityOptions()
	err := dp.ValidateConnectionSecurity(validOptions)
	if err != nil {
		log.Printf("Validation failed for valid options: %v", err)
	} else {
		log.Println("Validation succeeded for valid options")
	}

	// Create invalid options (TLS version too low)
	invalidOptions := update.DefaultConnectionSecurityOptions()
	invalidOptions.MinTLSVersion = tls.VersionTLS10
	err = dp.ValidateConnectionSecurity(invalidOptions)
	if err != nil {
		log.Printf("Validation correctly failed for invalid TLS version: %v", err)
	} else {
	}

	// Create invalid options (certificate pinning disabled)
	invalidOptions = update.DefaultConnectionSecurityOptions()
	invalidOptions.EnableCertificatePinning = false
	err = dp.ValidateConnectionSecurity(invalidOptions)
	if err != nil {
		log.Printf("Validation correctly failed for disabled certificate pinning: %v", err)
	} else {
	}

// demonstrateSignatureAlgorithmValidation demonstrates signature algorithm validation
func demonstrateSignatureAlgorithmValidation(dp *update.DowngradeProtection) {
	// Validate valid algorithms
	for _, algorithm := range dp.Policy.AllowedSignatureAlgorithms {
		err := dp.ValidateSignatureAlgorithm(algorithm)
		if err != nil {
			log.Printf("Validation failed for valid algorithm %s: %v", algorithm, err)
		} else {
			log.Printf("Validation succeeded for valid algorithm %s", algorithm)
		}
	}

	// Validate invalid algorithm
	invalidAlgorithm := update.SignatureAlgorithm("invalid")
	err := dp.ValidateSignatureAlgorithm(invalidAlgorithm)
	if err != nil {
		log.Printf("Validation correctly failed for invalid algorithm: %v", err)
	} else {
	}

// demonstrateKeySizeValidation demonstrates key size validation
func demonstrateKeySizeValidation(dp *update.DowngradeProtection) {
	// Validate valid key sizes
	for algorithm, minSize := range dp.Policy.MinimumKeySize {
		err := dp.ValidateKeySize(algorithm, minSize)
		if err != nil {
			log.Printf("Validation failed for valid key size %d for algorithm %s: %v", minSize, algorithm, err)
		} else {
			log.Printf("Validation succeeded for valid key size %d for algorithm %s", minSize, algorithm)
		}

		err = dp.ValidateKeySize(algorithm, minSize+1024)
		if err != nil {
			log.Printf("Validation failed for valid key size %d for algorithm %s: %v", minSize+1024, algorithm, err)
		} else {
			log.Printf("Validation succeeded for valid key size %d for algorithm %s", minSize+1024, algorithm)
		}
	}

	// Validate invalid key size
	for algorithm, minSize := range dp.Policy.MinimumKeySize {
		err := dp.ValidateKeySize(algorithm, minSize-1)
		if err != nil {
			log.Printf("Validation correctly failed for key size %d below minimum %d for algorithm %s: %v", minSize-1, minSize, algorithm, err)
		} else {
		}
	}

// demonstrateVersionValidation demonstrates version validation
func demonstrateVersionValidation(dp *update.DowngradeProtection) {
	// Validate valid versions
	for componentType, minVersion := range dp.Policy.MinimumVersions {
		err := dp.ValidateVersion(componentType, minVersion)
		if err != nil {
			log.Printf("Validation failed for valid version %s for component type %s: %v", minVersion, componentType, err)
		} else {
			log.Printf("Validation succeeded for valid version %s for component type %s", minVersion, componentType)
		}

		err = dp.ValidateVersion(componentType, minVersion+".1")
		if err != nil {
			log.Printf("Validation failed for valid version %s for component type %s: %v", minVersion+".1", componentType, err)
		} else {
			log.Printf("Validation succeeded for valid version %s for component type %s", minVersion+".1", componentType)
		}
	}

	// Validate invalid versions
	err := dp.ValidateVersion("core", "0.9.0")
	if err != nil {
		log.Printf("Validation correctly failed for version below minimum: %v", err)
	} else {
	}

// demonstrateUpdatePackageValidation demonstrates update package validation
func demonstrateUpdatePackageValidation(dp *update.DowngradeProtection) {
	// Create a valid update package
	validPkg := &update.UpdatePackage{
		Path: os.TempDir(),
		Manifest: update.UpdateManifest{
			Signature: "valid-signature",
			Versions: update.VersionInfo{
				Core:      "1.0.0",
				Templates: "1.0.0",
				Modules: map[string]string{
					"module1": "1.0.0",
					"module2": "1.0.0",
				},
			},
		},
	}

	// Validate valid update package
	err := dp.ValidateUpdatePackage(validPkg)
	if err != nil {
		log.Printf("Validation failed for valid update package: %v", err)
	} else {
		log.Println("Validation succeeded for valid update package")
	}

	// Create an invalid update package (missing signature)
	invalidPkg := *validPkg
	invalidPkg.Manifest.Signature = ""

	// Validate invalid update package
	err = dp.ValidateUpdatePackage(&invalidPkg)
	if err != nil {
		log.Printf("Validation correctly failed for update package with missing signature: %v", err)
	} else {
	}

// demonstratePolicyEnforcement demonstrates policy enforcement
func demonstratePolicyEnforcement(dp *update.DowngradeProtection) {
	// Create options with values below the minimum requirements
	options := &update.ConnectionSecurityOptions{
		MinTLSVersion:          tls.VersionTLS10,
		EnableCertificatePinning: false,
		CheckRevocation:        false,
		CipherSuites:           []uint16{tls.TLS_RSA_WITH_RC4_128_SHA}, // Weak cipher suite
	}

	// Display original options
	log.Println("Original connection security options:")
	log.Printf("  MinTLSVersion: %x", options.MinTLSVersion)
	log.Printf("  EnableCertificatePinning: %v", options.EnableCertificatePinning)
	log.Printf("  CheckRevocation: %v", options.CheckRevocation)
	log.Printf("  CipherSuites: %v", options.CipherSuites)

	// Enforce the security policy
	dp.EnforceSecurityPolicy(options)

	// Display enforced options
	log.Println("Enforced connection security options:")
	log.Printf("  MinTLSVersion: %x", options.MinTLSVersion)
	log.Printf("  EnableCertificatePinning: %v", options.EnableCertificatePinning)
	log.Printf("  CheckRevocation: %v", options.CheckRevocation)
	log.Printf("  CipherSuites: %v", options.CipherSuites)

// demonstrateSecureClientCreation demonstrates secure client creation
func demonstrateSecureClientCreation(dp *update.DowngradeProtection) {
	// Create a secure client
	client, err := dp.CreateSecureClient()
	if err != nil {
		log.Printf("Failed to create secure client: %v", err)
		return
	}

	log.Println("Secure client created successfully")

	// Display pinned certificates
	pinnedCerts := client.GetPinnedCertificates()
	log.Printf("Pinned certificates: %d", len(pinnedCerts))

	// Display retry configuration
	retryConfig := client.GetRetryConfig()
	log.Printf("Retry configuration: MaxRetries=%d, InitialDelay=%v", 
		retryConfig.MaxRetries, retryConfig.InitialDelay)

// demonstratePolicyUpdate demonstrates policy update
func demonstratePolicyUpdate(dp *update.DowngradeProtection) {
	// Update minimum version
	log.Println("Updating minimum version for 'core' to 2.0.0...")
	err := dp.UpdateMinimumVersion("core", "2.0.0")
	if err != nil {
		log.Printf("Failed to update minimum version: %v", err)
	} else {
		log.Printf("Minimum version updated successfully: %s", dp.Policy.MinimumVersions["core"])
	}

	// Update allowed signature algorithms
	log.Println("Updating allowed signature algorithms to Ed25519 and ECDSA only...")
	err = dp.UpdateAllowedSignatureAlgorithms([]update.SignatureAlgorithm{
		update.Ed25519Algorithm, 
		update.ECDSAAlgorithm,
	})
	if err != nil {
		log.Printf("Failed to update allowed signature algorithms: %v", err)
	} else {
		log.Printf("Allowed signature algorithms updated successfully: %v", dp.Policy.AllowedSignatureAlgorithms)
	}

	// Update minimum key size
	log.Println("Updating minimum key size for RSA to 4096 bits...")
	err = dp.UpdateMinimumKeySize("rsa", 4096)
	if err != nil {
		log.Printf("Failed to update minimum key size: %v", err)
	} else {
		log.Printf("Minimum key size updated successfully: %d", dp.Policy.MinimumKeySize["rsa"])
	}

	// Display updated policy
	log.Println("Updated security policy:")
	displayPolicy(dp.Policy)
