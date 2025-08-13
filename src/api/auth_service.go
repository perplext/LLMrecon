package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrAPIKeyNotFound    = errors.New("API key not found")
	ErrAPIKeyExpired     = errors.New("API key expired")
	ErrAPIKeyRevoked     = errors.New("API key revoked")
)

// AuthService handles authentication operations
type AuthService interface {
	// JWT operations
	GenerateJWT(userID string, claims map[string]interface{}) (string, error)
	ValidateJWT(token string) (*JWTClaims, error)
	RefreshJWT(token string) (string, error)
	
	// API key operations
	CreateAPIKey(request CreateAPIKeyRequest) (*APIKey, error)
	GetAPIKey(keyID string) (*APIKey, error)
	ListAPIKeys(filter APIKeyFilter) ([]APIKey, error)
	RevokeAPIKey(keyID string) error
	ValidateAPIKey(key string) (*APIKey, error)
	
	// User authentication
	Authenticate(username, password string) (*User, error)
	CreateUser(request CreateUserRequest) (*User, error)
	UpdatePassword(userID, oldPassword, newPassword string) error
}

// AuthServiceImpl implements AuthService
type AuthServiceImpl struct {
	jwtSecret     []byte
	jwtExpiration time.Duration
	apiKeys       map[string]*APIKey
	users         map[string]*User
	mu            sync.RWMutex
}

// NewAuthService creates a new auth service
func NewAuthService(config *Config) AuthService {
	return &AuthServiceImpl{
		jwtSecret:     []byte(config.JWTSecret),
		jwtExpiration: time.Duration(config.JWTExpiration) * time.Hour,
		apiKeys:       make(map[string]*APIKey),
		users:         make(map[string]*User),
	}
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID   string                 `json:"user_id"`
	Username string                 `json:"username"`
	Role     string                 `json:"role"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// APIKey represents an API key
type APIKey struct {
	ID          string            `json:"id"`
	Key         string            `json:"key,omitempty"` // Only returned on creation
	KeyHash     string            `json:"-"`              // Never exposed
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Scopes      []string          `json:"scopes"`
	RateLimit   int               `json:"rate_limit"` // Requests per minute
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	RevokedAt   *time.Time        `json:"revoked_at,omitempty"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// User represents a user
type User struct {
	ID           string            `json:"id"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	PasswordHash string            `json:"-"` // Never exposed
	Role         string            `json:"role"`
	Active       bool              `json:"active"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description,omitempty"`
	Scopes      []string          `json:"scopes,omitempty"`
	RateLimit   int               `json:"rate_limit,omitempty"`
	ExpiresIn   int               `json:"expires_in,omitempty"` // Days
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// APIKeyFilter represents API key filter criteria
type APIKeyFilter struct {
	Active   *bool
	ExpiredOnly bool
	RevokedOnly bool
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username string            `json:"username" validate:"required"`
	Email    string            `json:"email" validate:"required,email"`
	Password string            `json:"password" validate:"required,min=8"`
	Role     string            `json:"role,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GenerateJWT generates a new JWT token
func (s *AuthServiceImpl) GenerateJWT(userID string, claims map[string]interface{}) (string, error) {
	s.mu.RLock()
	user, exists := s.users[userID]
	s.mu.RUnlock()
	
	if !exists {
		return "", errors.New("user not found")
	}
	
	now := time.Now()
	jwtClaims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Extra:    claims,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiration)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "LLMrecon",
			Subject:   userID,
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(s.jwtSecret)
}

// ValidateJWT validates a JWT token
func (s *AuthServiceImpl) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	
	if err != nil {
		return nil, ErrTokenInvalid
	}
	
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrTokenInvalid
}

// RefreshJWT refreshes a JWT token
func (s *AuthServiceImpl) RefreshJWT(tokenString string) (string, error) {
	claims, err := s.ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}
	
	// Generate new token with same claims but new expiration
	return s.GenerateJWT(claims.UserID, claims.Extra)
}

// CreateAPIKey creates a new API key
func (s *AuthServiceImpl) CreateAPIKey(request CreateAPIKeyRequest) (*APIKey, error) {
	// Generate secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	key := base64.URLEncoding.EncodeToString(keyBytes)
	
	// Hash the key for storage
	keyHash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}
	
	// Calculate expiration
	var expiresAt *time.Time
	if request.ExpiresIn > 0 {
		exp := time.Now().AddDate(0, 0, request.ExpiresIn)
		expiresAt = &exp
	}
	
	// Default rate limit
	rateLimit := request.RateLimit
	if rateLimit <= 0 {
		rateLimit = 60 // 60 requests per minute
	}
	
	apiKey := &APIKey{
		ID:          generateID(),
		Key:         key, // Will be removed before storage
		KeyHash:     string(keyHash),
		Name:        request.Name,
		Description: request.Description,
		Scopes:      request.Scopes,
		RateLimit:   rateLimit,
		Metadata:    request.Metadata,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	s.mu.Lock()
	s.apiKeys[apiKey.ID] = apiKey
	s.mu.Unlock()
	
	// Return copy with key included (only time it's exposed)
	result := *apiKey
	return &result, nil
}

// GetAPIKey retrieves an API key by ID
func (s *AuthServiceImpl) GetAPIKey(keyID string) (*APIKey, error) {
	s.mu.RLock()
	apiKey, exists := s.apiKeys[keyID]
	s.mu.RUnlock()
	
	if !exists {
		return nil, ErrAPIKeyNotFound
	}
	
	// Return copy without key
	result := *apiKey
	result.Key = ""
	return &result, nil
}

// ListAPIKeys lists API keys based on filter
func (s *AuthServiceImpl) ListAPIKeys(filter APIKeyFilter) ([]APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var results []APIKey
	now := time.Now()
	
	for _, apiKey := range s.apiKeys {
		// Apply filters
		if filter.Active != nil {
			isActive := apiKey.RevokedAt == nil && (apiKey.ExpiresAt == nil || apiKey.ExpiresAt.After(now))
			if *filter.Active != isActive {
				continue
			}
		}
		
		if filter.ExpiredOnly && (apiKey.ExpiresAt == nil || apiKey.ExpiresAt.After(now)) {
			continue
		}
		
		if filter.RevokedOnly && apiKey.RevokedAt == nil {
			continue
		}
		
		// Return copy without key
		result := *apiKey
		result.Key = ""
		results = append(results, result)
	}
	
	return results, nil
}

// RevokeAPIKey revokes an API key
func (s *AuthServiceImpl) RevokeAPIKey(keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	apiKey, exists := s.apiKeys[keyID]
	if !exists {
		return ErrAPIKeyNotFound
	}
	
	now := time.Now()
	apiKey.RevokedAt = &now
	apiKey.UpdatedAt = now
	
	return nil
}

// ValidateAPIKey validates an API key
func (s *AuthServiceImpl) ValidateAPIKey(key string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Normalize the key
	key = normalizeAPIKey(key)
	
	// Check all API keys
	for _, apiKey := range s.apiKeys {
		// Compare using bcrypt
		if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(key)); err == nil {
			// Check if revoked
			if apiKey.RevokedAt != nil {
				return nil, ErrAPIKeyRevoked
			}
			
			// Check if expired
			if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
				return nil, ErrAPIKeyExpired
			}
			
			// Update last used
			now := time.Now()
			apiKey.LastUsedAt = &now
			
			// Return copy without key
			result := *apiKey
			result.Key = ""
			return &result, nil
		}
	}
	
	return nil, ErrInvalidCredentials
}

// Authenticate authenticates a user
func (s *AuthServiceImpl) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Find user by username
	var user *User
	for _, u := range s.users {
		if u.Username == username {
			user = u
			break
		}
	}
	
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	
	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// Check if active
	if !user.Active {
		return nil, errors.New("user account is inactive")
	}
	
	// Return copy
	result := *user
	return &result, nil
}

// CreateUser creates a new user
func (s *AuthServiceImpl) CreateUser(request CreateUserRequest) (*User, error) {
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Default role
	role := request.Role
	if role == "" {
		role = "user"
	}
	
	user := &User{
		ID:           generateID(),
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: string(passwordHash),
		Role:         role,
		Active:       true,
		Metadata:     request.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	s.mu.Lock()
	
	// Check if username already exists
	for _, u := range s.users {
		if u.Username == user.Username {
			s.mu.Unlock()
			return nil, errors.New("username already exists")
		}
	}
	
	s.users[user.ID] = user
	s.mu.Unlock()
	
	// Return copy
	result := *user
	return &result, nil
}

// UpdatePassword updates a user's password
func (s *AuthServiceImpl) UpdatePassword(userID, oldPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	
	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}
	
	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	user.PasswordHash = string(passwordHash)
	user.UpdatedAt = time.Now()
	
	return nil
}

// normalizeAPIKey normalizes an API key for comparison
func normalizeAPIKey(key string) string {
	// Remove any whitespace and ensure consistent format
	return strings.TrimSpace(key)
}

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Helper function for constant-time string comparison
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}