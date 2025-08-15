package bundle

import (
	"archive/zip"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/klauspost/compress/zstd"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
)

// CompressionHandler handles various compression algorithms
type CompressionHandler interface {
	Compress(src io.Reader, dst io.Writer) error
	Decompress(src io.Reader, dst io.Writer) error
	GetExtension() string

// EncryptionHandler handles encryption operations
type EncryptionHandler interface {
	Encrypt(data []byte, password string) ([]byte, error)
	Decrypt(data []byte, password string) ([]byte, error)
	EncryptStream(src io.Reader, dst io.Writer, password string) error
	DecryptStream(src io.Reader, dst io.Writer, password string) error

// CompressionFactory creates compression handlers
type CompressionFactory struct {
	handlers map[CompressionType]CompressionHandler
}

// NewCompressionFactory creates a new compression factory
func NewCompressionFactory() *CompressionFactory {
	factory := &CompressionFactory{
		handlers: make(map[CompressionType]CompressionHandler),
	}

	// Register default handlers
	factory.RegisterHandler(CompressionGzip, &GzipHandler{})
	factory.RegisterHandler(CompressionZstd, &ZstdHandler{})
	factory.RegisterHandler(CompressionNone, &NoCompressionHandler{})

	return factory

// RegisterHandler registers a compression handler
func (f *CompressionFactory) RegisterHandler(compressionType CompressionType, handler CompressionHandler) {
	f.handlers[compressionType] = handler

// GetHandler returns a compression handler for the given type
func (f *CompressionFactory) GetHandler(compressionType CompressionType) (CompressionHandler, error) {
	handler, exists := f.handlers[compressionType]
	if !exists {
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}
	return handler, nil

// GzipHandler handles gzip compression
type GzipHandler struct {
	Level int
}

func (h *GzipHandler) Compress(src io.Reader, dst io.Writer) error {
	var gzWriter *gzip.Writer
	var err error

	if h.Level == 0 {
		gzWriter = gzip.NewWriter(dst)
	} else {
		gzWriter, err = gzip.NewWriterLevel(dst, h.Level)
		if err != nil {
			return err
		}
	}
	defer func() { if err := gzWriter.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	_, err = io.Copy(gzWriter, src)
	return err

func (h *GzipHandler) Decompress(src io.Reader, dst io.Writer) error {
	gzReader, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer func() { if err := gzReader.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	_, err = io.Copy(dst, gzReader)
	return err

func (h *GzipHandler) GetExtension() string {
	return ".gz"

// ZstdHandler handles Zstandard compression
type ZstdHandler struct {
	Level int
	
}

func (h *ZstdHandler) Compress(src io.Reader, dst io.Writer) error {
	encoder, err := zstd.NewWriter(dst, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(h.Level)))
	if err != nil {
		return err
	}
	defer func() { if err := encoder.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	_, err = io.Copy(encoder, src)
	return err
	

func (h *ZstdHandler) Decompress(src io.Reader, dst io.Writer) error {
	decoder, err := zstd.NewReader(src)
	if err != nil {
		return err
	}
	defer func() { if err := decoder.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	_, err = io.Copy(dst, decoder)
	return err
func (h *ZstdHandler) GetExtension() string {
	return ".zst"
	

// NoCompressionHandler handles no compression (passthrough)
type NoCompressionHandler struct{}

func (h *NoCompressionHandler) Compress(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err

func (h *NoCompressionHandler) Decompress(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err

func (h *NoCompressionHandler) GetExtension() string {
	return ""

// EncryptionFactory creates encryption handlers
type EncryptionFactory struct {
	handlers map[string]EncryptionHandler
}

// NewEncryptionFactory creates a new encryption factory
func NewEncryptionFactory() *EncryptionFactory {
	factory := &EncryptionFactory{
		handlers: make(map[string]EncryptionHandler),
	}

	// Register default handlers
	factory.RegisterHandler("aes-256-gcm", &AESGCMHandler{})
	factory.RegisterHandler("chacha20-poly1305", &ChaCha20Handler{})

	return factory

// RegisterHandler registers an encryption handler
func (f *EncryptionFactory) RegisterHandler(algorithm string, handler EncryptionHandler) {
	f.handlers[algorithm] = handler

// GetHandler returns an encryption handler for the given algorithm
func (f *EncryptionFactory) GetHandler(algorithm string) (EncryptionHandler, error) {
	handler, exists := f.handlers[algorithm]
	if !exists {
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}
	return handler, nil
// AESGCMHandler handles AES-256-GCM encryption
type AESGCMHandler struct{}
func (h *AESGCMHandler) Encrypt(plaintext []byte, password string) ([]byte, error) {
	// Derive key from password
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(result, salt)
	copy(result[len(salt):], nonce)
	copy(result[len(salt)+len(nonce):], ciphertext)

	return result, nil

func (h *AESGCMHandler) Decrypt(ciphertext []byte, password string) ([]byte, error) {
	// Extract salt
	if len(ciphertext) < 32 {
		return nil, fmt.Errorf("ciphertext too short")
	}
	salt := ciphertext[:32]

	// Derive key
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < 32+nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[32 : 32+nonceSize]
	ciphertextData := ciphertext[32+nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
func (h *AESGCMHandler) EncryptStream(src io.Reader, dst io.Writer, password string) error {
	// Read all data (not ideal for large files)
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	encrypted, err := h.Encrypt(data, password)
	if err != nil {
		return err
	}

	_, err = dst.Write(encrypted)
	return err

func (h *AESGCMHandler) DecryptStream(src io.Reader, dst io.Writer, password string) error {
	// Read all data (not ideal for large files)
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	decrypted, err := h.Decrypt(data, password)
	if err != nil {
		return err
	}

	_, err = dst.Write(decrypted)
	return err

// ChaCha20Handler handles ChaCha20-Poly1305 encryption
type ChaCha20Handler struct{}

func (h *ChaCha20Handler) Encrypt(plaintext []byte, password string) ([]byte, error) {
	// Derive key using Argon2
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Create cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	// Encrypt
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	// Combine salt + nonce + ciphertext
	result := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(result, salt)
	copy(result[len(salt):], nonce)
	copy(result[len(salt)+len(nonce):], ciphertext)

	return result, nil
	

func (h *ChaCha20Handler) Decrypt(ciphertext []byte, password string) ([]byte, error) {
	if len(ciphertext) < 16 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract salt
	salt := ciphertext[:16]

	// Derive key
	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Create cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Extract nonce and ciphertext
	nonceSize := aead.NonceSize()
	if len(ciphertext) < 16+nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[16 : 16+nonceSize]
	ciphertextData := ciphertext[16+nonceSize:]
	// Decrypt
	plaintext, err := aead.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil

func (h *ChaCha20Handler) EncryptStream(src io.Reader, dst io.Writer, password string) error {
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	encrypted, err := h.Encrypt(data, password)
	if err != nil {
		return err
	}

	_, err = dst.Write(encrypted)
	return err

func (h *ChaCha20Handler) DecryptStream(src io.Reader, dst io.Writer, password string) error {
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	decrypted, err := h.Decrypt(data, password)
	if err != nil {
		return err
	}

	_, err = dst.Write(decrypted)
	return err
	

// BundleCompressor handles bundle compression and encryption
type BundleCompressor struct {
	compressionFactory *CompressionFactory
	encryptionFactory  *EncryptionFactory

// NewBundleCompressor creates a new bundle compressor
func NewBundleCompressor() *BundleCompressor {
	return &BundleCompressor{
		compressionFactory: NewCompressionFactory(),
		encryptionFactory:  NewEncryptionFactory(),
	}

// CompressBundle compresses a bundle directory
func (c *BundleCompressor) CompressBundle(bundlePath string, outputPath string, options CompressOptions) error {
	// Get compression handler
	handler, err := c.compressionFactory.GetHandler(options.Compression)
	if err != nil {
		return err
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { if err := outputFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Create compression writer
	var writer io.WriteCloser = outputFile

	// Add encryption layer if requested
	if options.Encryption != nil {
		_, err := c.encryptionFactory.GetHandler(options.Encryption.Algorithm)
		if err != nil {
			return err
		}

		// For simplicity, we'll encrypt after compression
		// In production, you might want streaming encryption
		tempFile, err := os.CreateTemp("", "bundle-compress-*")
		if err != nil {
			return err
		}
		defer os.Remove(tempFile.Name())
		writer = tempFile
	}

	// Apply compression
	if options.Format == FormatZip {
		err = c.createZipArchive(bundlePath, writer, handler)
	} else {
		err = c.createTarArchive(bundlePath, writer, handler)
	}
	if err != nil {
		return err
	}
	// Apply encryption if requested
	if options.Encryption != nil {
		writer.Close()
		
		// Read compressed data
		compressedFile, err := os.Open(filepath.Clean(writer.(*os.File)).Name())
		if err != nil {
			return err
		}
		defer func() { if err := compressedFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

		// Encrypt and write to final output
		encHandler, _ := c.encryptionFactory.GetHandler(options.Encryption.Algorithm)
		err = encHandler.EncryptStream(compressedFile, outputFile, options.Encryption.Password)
		if err != nil {
			return err
		}

		// Add encryption header
		c.writeEncryptionHeader(outputPath, options.Encryption)
	}

	return nil

// DecompressBundle decompresses a bundle
func (c *BundleCompressor) DecompressBundle(archivePath string, outputPath string, options DecompressOptions) error {
	// Open archive file
	archiveFile, err := os.Open(filepath.Clean(archivePath))
	if err != nil {
		return err
	}
	defer func() { if err := archiveFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	var reader io.Reader = archiveFile

	// Check for encryption
	if c.isEncrypted(archivePath) {
		if options.Password == "" {
			return fmt.Errorf("archive is encrypted but no password provided")
		}

		// Read encryption header
		encInfo, err := c.readEncryptionHeader(archivePath)
		if err != nil {
			return err
		}

		// Get encryption handler
		encHandler, err := c.encryptionFactory.GetHandler(encInfo.Algorithm)
		if err != nil {
			return err
		}

		// Decrypt to temporary file
		tempFile, err := os.CreateTemp("", "bundle-decrypt-*")
		if err != nil {
			return err
		}
		defer os.Remove(tempFile.Name())
		// Skip header and decrypt
		archiveFile.Seek(int64(encInfo.HeaderSize), 0)
		err = encHandler.DecryptStream(archiveFile, tempFile, options.Password)
		if err != nil {
			return err
		}

		tempFile.Close()
		reader, err = os.Open(filepath.Clean(tempFile.Name()))
		if err != nil {
			return err
		}
		defer func() { if err := reader.(*os.File).Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	}

	// Detect compression type
	compressionType := c.detectCompressionType(archivePath)
	handler, err := c.compressionFactory.GetHandler(compressionType)
	if err != nil {
		return err
	}

	// Create output directory
	if err := os.MkdirAll(outputPath, 0700); err != nil {
		return err
	}

	// Decompress
	if strings.HasSuffix(archivePath, ".zip") {
		return c.extractZipArchive(reader, outputPath)
	} else {
		return c.extractTarArchive(reader, outputPath, handler)
	}

// CompressOptions defines options for compression
type CompressOptions struct {
	Format      ExportFormat
	Compression CompressionType
	Encryption  *EncryptionOptions
	Level       int // Compression level
	
}

// DecompressOptions defines options for decompression
type DecompressOptions struct {
	Password string
	Validate bool // Validate checksums after decompression

// EncryptionHeader contains encryption metadata
type EncryptionHeader struct {
	Algorithm  string `json:"algorithm"`
	HeaderSize int    `json:"headerSize"`
	Version    string `json:"version"`
// Helper methods
}

func (c *BundleCompressor) createZipArchive(bundlePath string, output io.Writer, handler CompressionHandler) error {
	zipWriter := zip.NewWriter(output)
	defer func() { if err := zipWriter.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	return filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(bundlePath, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}
		// Create zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		// Create file in zip
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy file content if not a directory
		if !info.IsDir() {
			file, err := os.Open(filepath.Clean(path))
			if err != nil {
				return err
			}
			defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})

func (c *BundleCompressor) createTarArchive(bundlePath string, output io.Writer, handler CompressionHandler) error {
	// Implementation would create a tar archive with the specified compression
	// This is simplified for brevity
	return fmt.Errorf("tar archive creation not fully implemented")

func (c *BundleCompressor) extractZipArchive(input io.Reader, outputPath string) error {
	// Implementation would extract a zip archive
	// This is simplified for brevity
	return fmt.Errorf("zip extraction not fully implemented")

func (c *BundleCompressor) extractTarArchive(input io.Reader, outputPath string, handler CompressionHandler) error {
	// Implementation would extract a tar archive with decompression
	// This is simplified for brevity
	return fmt.Errorf("tar extraction not fully implemented")

func (c *BundleCompressor) detectCompressionType(path string) CompressionType {
	switch {
	case strings.HasSuffix(path, ".gz"):
		return CompressionGzip
	case strings.HasSuffix(path, ".zst"):
		return CompressionZstd
	default:
		return CompressionNone
	}

func (c *BundleCompressor) isEncrypted(path string) bool {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Check for encryption header
	header := make([]byte, 16)
	if _, err := file.Read(header); err != nil {
		return false
	}

	return string(header[:8]) == "LLMR-ENC"

func (c *BundleCompressor) writeEncryptionHeader(path string, options *EncryptionOptions) error {
	// Prepend encryption header to file
	header := EncryptionHeader{
		Algorithm: options.Algorithm,
		Version:   "1.0",
	}

	headerData, err := json.Marshal(header)
	if err != nil {
		return err
	}

	// Create header with fixed size
	fullHeader := make([]byte, 256)
	copy(fullHeader, []byte("LLMR-ENC"))
	copy(fullHeader[8:], headerData)
	header.HeaderSize = 256

	// Read existing file
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}

	// Write header + content
	return os.WriteFile(filepath.Clean(path, append(fullHeader, content...)), 0600)
	

func (c *BundleCompressor) readEncryptionHeader(path string) (*EncryptionHeader, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Read header
	headerData := make([]byte, 256)
	if _, err := file.Read(headerData); err != nil {
		return nil, err
	}

	// Verify magic bytes
	if string(headerData[:8]) != "LLMR-ENC" {
		return nil, fmt.Errorf("invalid encryption header")
	}

	// Parse header
	var header EncryptionHeader
	if err := json.Unmarshal(headerData[8:], &header); err != nil {
		return nil, err
	}

	header.HeaderSize = 256
	return &header, nil

// PasswordStrengthChecker checks password strength
type PasswordStrengthChecker struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumbers bool
	RequireSpecial bool
}

// CheckPassword validates password strength
func (p *PasswordStrengthChecker) CheckPassword(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters long", p.MinLength)
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if p.RequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if p.RequireNumbers && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if p.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil

// GenerateKeyFromPassword generates an encryption key from a password
func GenerateKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) < 16 {
		return nil, fmt.Errorf("salt must be at least 16 bytes")
	}

	// Use Argon2id for key derivation
	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return key, nil

// GenerateRandomPassword generates a secure random password
func GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	
	password := make([]byte, length)
	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[n.Int64()]
	}
	
	return string(password), nil

// HashPassword creates a secure hash of a password for storage
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	
	// Encode salt and hash together
	result := make([]byte, len(salt)+len(hash))
	copy(result, salt)
	copy(result[len(salt):], hash)
	
	return base64.StdEncoding.EncodeToString(result), nil

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, encodedHash string) bool {
	data, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false
	}

	if len(data) < 48 { // 16 bytes salt + 32 bytes hash
		return false
	}

	salt := data[:16]
	expectedHash := data[16:]
	
	actualHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	
	// Constant time comparison
	if len(expectedHash) != len(actualHash) {
		return false
	}
	
	var result byte
	for i := range expectedHash {
		result |= expectedHash[i] ^ actualHash[i]
	}
	
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
