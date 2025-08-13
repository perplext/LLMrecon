// Package keystore provides secure storage for cryptographic keys and sensitive materials.
package keystore

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ExportKey exports a key in the specified format
func (ks *FileKeyStore) ExportKey(id string, format string, includePrivate bool) ([]byte, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key for export: %w", err)
	}

	// Check if private key is requested but not available
	if includePrivate && !key.Metadata.HasPrivateKey {
		return nil, errors.New("private key requested for export but not available")
	}

	// Log the export attempt
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("export", key.Metadata.ID, 
			fmt.Sprintf("Exporting key in %s format (includePrivate=%v)", format, includePrivate))
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.ExportKey(id, format, includePrivate)
	}

	// Check if format matches the key's format
	if format != key.Material.Format && format != "PEM" && format != "DER" {
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	switch format {
	case "PEM":
		// If already in PEM format, return directly
		if key.Material.Format == "PEM" {
			if includePrivate {
				return key.Material.Private, nil
			}
			return key.Material.Public, nil
		}

		// Convert from DER to PEM
		if key.Material.Format == "DER" {
			var blockType string
			var keyBytes []byte

			if includePrivate {
				keyBytes = key.Material.Private
				switch key.Metadata.Type {
				case RSAKey:
					blockType = "RSA PRIVATE KEY"
				case ECDSAKey:
					blockType = "EC PRIVATE KEY"
				default:
					blockType = "PRIVATE KEY"
				}
			} else {
				keyBytes = key.Material.Public
				switch key.Metadata.Type {
				case RSAKey:
					blockType = "RSA PUBLIC KEY"
				case ECDSAKey:
					blockType = "EC PUBLIC KEY"
				default:
					blockType = "PUBLIC KEY"
				}
			}

			pemBlock := &pem.Block{
				Type:  blockType,
				Bytes: keyBytes,
			}
			return pem.EncodeToMemory(pemBlock), nil
		}

	case "DER":
		// If already in DER format, return directly
		if key.Material.Format == "DER" {
			if includePrivate {
				return key.Material.Private, nil
			}
			return key.Material.Public, nil
		}

		// Convert from PEM to DER
		if key.Material.Format == "PEM" {
			var pemData []byte
			if includePrivate {
				pemData = key.Material.Private
			} else {
				pemData = key.Material.Public
			}

			block, _ := pem.Decode(pemData)
			if block == nil {
				return nil, errors.New("failed to decode PEM block")
			}
			return block.Bytes, nil
		}

	case "RAW":
		// Only for symmetric keys
		if key.Metadata.Type != SymmetricKey {
			return nil, errors.New("RAW format only supported for symmetric keys")
		}
		return key.Material.Private, nil
	}

	return nil, fmt.Errorf("unsupported export format: %s", format)
}

// ImportKey imports a key from the specified format
func (ks *FileKeyStore) ImportKey(keyData []byte, format string, metadata *KeyMetadata) (*Key, error) {
	// Validate metadata
	if metadata == nil {
		return nil, errors.New("metadata is required for key import")
	}
	if metadata.ID == "" {
		metadata.ID = generateKeyID(metadata.Name, metadata.Type)
	}
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}
	if metadata.UpdatedAt.IsZero() {
		metadata.UpdatedAt = time.Now()
	}

	// Log the import attempt
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("import", metadata.ID, 
			fmt.Sprintf("Importing key in %s format", format))
	}

	// Handle HSM-protected keys
	if metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.ImportKey(keyData, format, metadata)
	}

	// Process key data based on format and key type
	var keyMaterial KeyMaterial
	var err error

	switch format {
	case "PEM":
		keyMaterial, err = processPEMKey(keyData, metadata.Type)
	case "DER":
		keyMaterial, err = processDERKey(keyData, metadata.Type)
	case "RAW":
		// Only for symmetric keys
		if metadata.Type != SymmetricKey {
			return nil, errors.New("RAW format only supported for symmetric keys")
		}
		keyMaterial = KeyMaterial{
			Private: keyData,
			Format:  "RAW",
		}
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process key data: %w", err)
	}

	// Calculate fingerprint
	var fingerprintData []byte
	if len(keyMaterial.Public) > 0 {
		fingerprintData = keyMaterial.Public
	} else {
		fingerprintData = keyMaterial.Private
	}

	fingerprint, err := calculateKeyFingerprint(fingerprintData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate key fingerprint: %w", err)
	}
	metadata.Fingerprint = fingerprint

	// Set key flags
	metadata.HasPrivateKey = len(keyMaterial.Private) > 0
	metadata.HasPublicKey = len(keyMaterial.Public) > 0

	// Create the key
	key := &Key{
		Metadata: *metadata,
		Material: keyMaterial,
	}

	// Store the key
	if err := ks.StoreKey(key); err != nil {
		return nil, fmt.Errorf("failed to store imported key: %w", err)
	}

	// Log the successful import
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("import", metadata.ID, 
			"Key imported successfully")
	}

	return key, nil
}

// processPEMKey processes a PEM-encoded key
func processPEMKey(pemData []byte, keyType KeyType) (KeyMaterial, error) {
	// Decode PEM block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return KeyMaterial{}, errors.New("failed to decode PEM block")
	}

	// Process based on block type
	switch block.Type {
	case "RSA PRIVATE KEY", "PRIVATE KEY", "EC PRIVATE KEY":
		// This is a private key
		// Try to extract public key if possible
		var publicKey []byte

		switch keyType {
		case RSAKey:
			var rsaKey *rsa.PrivateKey
			privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				// Try PKCS8 format
				pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					return KeyMaterial{}, fmt.Errorf("failed to parse RSA private key: %w", err)
				}
				var ok bool
				rsaKey, ok = pkcs8Key.(*rsa.PrivateKey)
				if !ok {
					return KeyMaterial{}, errors.New("parsed key is not an RSA private key")
				}
			} else {
				rsaKey = privateKey
			}
			
			// Extract public key
			publicKeyBytes, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
			if err != nil {
				return KeyMaterial{}, fmt.Errorf("failed to marshal RSA public key: %w", err)
			}
			
			publicKeyPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: publicKeyBytes,
			})
			publicKey = publicKeyPEM
		
		case ECDSAKey:
			var ecdsaPrivKey *ecdsa.PrivateKey
			privateKey, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				// Try PKCS8 format
				pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					return KeyMaterial{}, fmt.Errorf("failed to parse ECDSA private key: %w", err)
				}
				var ok bool
				ecdsaPrivKey, ok = pkcs8Key.(*ecdsa.PrivateKey)
				if !ok {
					return KeyMaterial{}, errors.New("parsed key is not an ECDSA private key")
				}
			} else {
				ecdsaPrivKey = privateKey
			}
			
			// Extract public key
			publicKeyBytes, err := x509.MarshalPKIXPublicKey(&ecdsaPrivKey.PublicKey)
			if err != nil {
				return KeyMaterial{}, fmt.Errorf("failed to marshal ECDSA public key: %w", err)
			}
			
			publicKeyPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "EC PUBLIC KEY",
				Bytes: publicKeyBytes,
			})
			publicKey = publicKeyPEM
		
		case Ed25519Key:
			// Try PKCS8 format for Ed25519
			pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return KeyMaterial{}, fmt.Errorf("failed to parse Ed25519 private key: %w", err)
			}
			
			ed25519PrivKey, ok := pkcs8Key.(ed25519.PrivateKey)
			if !ok {
				return KeyMaterial{}, errors.New("parsed key is not an Ed25519 private key")
			}
			
			// Extract public key
			pubKey := ed25519PrivKey.Public().(ed25519.PublicKey)
			publicKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
			if err != nil {
				return KeyMaterial{}, fmt.Errorf("failed to marshal Ed25519 public key: %w", err)
			}
			
			publicKeyPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: publicKeyBytes,
			})
			publicKey = publicKeyPEM
		}

		return KeyMaterial{
			Private: pemData,
			Public:  publicKey,
			Format:  "PEM",
		}, nil

	case "RSA PUBLIC KEY", "PUBLIC KEY", "EC PUBLIC KEY":
		// This is a public key
		return KeyMaterial{
			Public: pemData,
			Format: "PEM",
		}, nil

	case "CERTIFICATE":
		// This is a certificate
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return KeyMaterial{}, fmt.Errorf("failed to parse certificate: %w", err)
		}

		// Extract public key
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
		if err != nil {
			return KeyMaterial{}, fmt.Errorf("failed to marshal public key from certificate: %w", err)
		}

		var blockType string
		switch cert.PublicKey.(type) {
		case *rsa.PublicKey:
			blockType = "RSA PUBLIC KEY"
		case *ecdsa.PublicKey:
			blockType = "EC PUBLIC KEY"
		default:
			blockType = "PUBLIC KEY"
		}

		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  blockType,
			Bytes: publicKeyBytes,
		})

		return KeyMaterial{
			Public:      publicKeyPEM,
			Certificate: pemData,
			Format:      "PEM",
		}, nil

	default:
		return KeyMaterial{}, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

// processDERKey processes a DER-encoded key
func processDERKey(derData []byte, keyType KeyType) (KeyMaterial, error) {
	var blockType string
	var err error

	// Try to determine key type
	switch keyType {
	case RSAKey:
		// Try to parse as private key first
		_, err = x509.ParsePKCS1PrivateKey(derData)
		if err == nil {
			blockType = "RSA PRIVATE KEY"
		} else {
			// Try to parse as public key
			_, err = x509.ParsePKCS1PublicKey(derData)
			if err == nil {
				blockType = "RSA PUBLIC KEY"
			} else {
				// Try to parse as PKIX public key
				pubKey, err := x509.ParsePKIXPublicKey(derData)
				if err == nil {
					_, ok := pubKey.(*rsa.PublicKey)
					if ok {
						blockType = "RSA PUBLIC KEY"
					}
				}
			}
		}

	case ECDSAKey:
		// Try to parse as private key first
		_, err = x509.ParseECPrivateKey(derData)
		if err == nil {
			blockType = "EC PRIVATE KEY"
		} else {
			// Try to parse as PKIX public key
			pubKey, err := x509.ParsePKIXPublicKey(derData)
			if err == nil {
				_, ok := pubKey.(*ecdsa.PublicKey)
				if ok {
					blockType = "EC PUBLIC KEY"
				}
			}
		}

	case Ed25519Key:
		// Try to parse as PKIX public key
		pubKey, err := x509.ParsePKIXPublicKey(derData)
		if err == nil {
			_, ok := pubKey.(ed25519.PublicKey)
			if ok {
				blockType = "PUBLIC KEY"
			}
		}

	case CertificateKey:
		// Try to parse as certificate
		_, err = x509.ParseCertificate(derData)
		if err == nil {
			blockType = "CERTIFICATE"
		}
	}

	if blockType == "" {
		// If we couldn't determine the type, try PKCS8
		_, err = x509.ParsePKCS8PrivateKey(derData)
		if err == nil {
			blockType = "PRIVATE KEY"
		} else {
			return KeyMaterial{}, errors.New("unable to determine key type from DER data")
		}
	}

	// Convert to PEM for storage
	pemBlock := &pem.Block{
		Type:  blockType,
		Bytes: derData,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	// Process the PEM data (which will extract public key if needed)
	return processPEMKey(pemData, keyType)
}
