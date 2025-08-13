// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/mfa"
)

// NewAccessControlSystemWithStores creates a new access control system with the provided stores
func NewAccessControlSystemWithStores(
	config *AccessControlConfig,
	userStore UserStore,
	sessionStore SessionStore,
	auditLogger AuditLogger,
	incidentStore IncidentStore,
	vulnerabilityStore VulnerabilityStore,
) (*AccessControlSystem, error) {
	if config == nil {
		config = DefaultAccessControlConfig()
	}

	// Initialize the audit logger
	if err := auditLogger.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
	}

	// Create the MFA manager
	// Create an MFA store and config for the MFA manager
	mfaStore := mfa.NewInMemoryMFAStore()
	mfaConfig := &mfa.MFAManagerConfig{
		BackupConfig: &mfa.BackupCodeConfig{
			CodeLength: 8,
			CodeCount:  10,
		},
		VerificationExpiration: 15, // 15 minutes
	}
	mfaManager := mfa.NewDefaultMFAManager(mfaStore, mfaConfig)

	// Create the security manager
	// Note: We don't use these managers directly as we create simpleRBACManager and simpleSecurityManager below
	_ = NewSecurityManager(
		config,
		incidentStore,
		vulnerabilityStore,
		auditLogger,
	)

	// Create the RBAC manager
	_ = NewRBACManager(config)

	// Create the auth manager
	// Extract auth config from AccessControlConfig
	authConfig := &AuthConfig{
		SessionTimeout:         time.Duration(config.SessionPolicy.TokenExpiration) * time.Minute,
		SessionMaxInactive:     time.Duration(config.SessionPolicy.InactivityTimeout) * time.Minute,
		PasswordMinLength:      config.PasswordPolicy.MinLength,
		PasswordRequireUpper:   config.PasswordPolicy.RequireUppercase,
		PasswordRequireLower:   config.PasswordPolicy.RequireLowercase,
		PasswordRequireNumber:  config.PasswordPolicy.RequireNumbers,
		PasswordRequireSpecial: config.PasswordPolicy.RequireSpecialChars,
		PasswordMaxAge:         time.Duration(config.PasswordPolicy.MaxAge) * 24 * time.Hour,
		MFAEnabled:             config.EnableMFA,
	}
	authManager, err := NewAuthManager(
		userStore,
		sessionStore,
		auditLogger,
		mfaManager,
		authConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Create the audit manager
	auditManager, err := NewAuditManager(auditLogger, &config.AuditConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit manager: %w", err)
	}

	// Create a simpleRBACManager for the AccessControlSystem
	simpleRBAC := &simpleRBACManager{
		config:          config,
		rolePermissions: make(map[string][]string),
		userRoles:       make(map[string][]string),
	}
	if config.RolePermissions != nil {
		simpleRBAC.rolePermissions = config.RolePermissions
	}

	// Create a simpleSecurityManager for the AccessControlSystem
	simpleSecurity := &simpleSecurityManager{
		auditLogger: auditLogger,
		config:      &SecurityConfig{},
	}

	return &AccessControlSystem{
		authManager:     authManager,
		rbacManager:     simpleRBAC,
		auditManager:    auditManager,
		securityManager: simpleSecurity,
		mfaManager:      mfaManager,
		config:          config,
	}, nil
}

// InMemoryUserStore is a simple in-memory implementation of UserStore
type InMemoryUserStore struct {
	users map[string]*User
	mu    sync.RWMutex
}

// NewInMemoryUserStore creates a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// CreateUser creates a new user
func (s *InMemoryUserStore) CreateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[user.ID]; exists {
		return fmt.Errorf("user with ID %s already exists", user.ID)
	}

	// Check if username is taken
	for _, existingUser := range s.users {
		if existingUser.Username == user.Username {
			return fmt.Errorf("username %s is already taken", user.Username)
		}
		if existingUser.Email == user.Email {
			return fmt.Errorf("email %s is already in use", user.Email)
		}
	}

	// Create a copy of the user
	newUser := *user

	// Set created and updated timestamps if not set
	if newUser.CreatedAt.IsZero() {
		newUser.CreatedAt = time.Now()
	}
	if newUser.UpdatedAt.IsZero() {
		newUser.UpdatedAt = time.Now()
	}

	// Store the user
	s.users[user.ID] = &newUser

	return nil
}

// GetUserByID retrieves a user by ID
func (s *InMemoryUserStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %s not found", id)
	}

	// Return a copy of the user
	userCopy := *user
	return &userCopy, nil
}

// GetUserByUsername retrieves a user by username
func (s *InMemoryUserStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			// Return a copy of the user
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, fmt.Errorf("user with username %s not found", username)
}

// GetUserByEmail retrieves a user by email
func (s *InMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			// Return a copy of the user
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, fmt.Errorf("user with email %s not found", email)
}

// UpdateUser updates an existing user
func (s *InMemoryUserStore) UpdateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	existingUser, exists := s.users[user.ID]
	if !exists {
		return fmt.Errorf("user with ID %s not found", user.ID)
	}

	// Check if username is taken by another user
	for id, otherUser := range s.users {
		if id != user.ID && otherUser.Username == user.Username {
			return fmt.Errorf("username %s is already taken", user.Username)
		}
		if id != user.ID && otherUser.Email == user.Email {
			return fmt.Errorf("email %s is already in use", user.Email)
		}
	}

	// Create a copy of the user
	updatedUser := *user

	// Preserve created timestamp
	updatedUser.CreatedAt = existingUser.CreatedAt

	// Update timestamp
	updatedUser.UpdatedAt = time.Now()

	// Store the updated user
	s.users[user.ID] = &updatedUser

	return nil
}

// DeleteUser deletes a user by ID
func (s *InMemoryUserStore) DeleteUser(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user with ID %s not found", id)
	}

	// Delete the user
	delete(s.users, id)

	return nil
}

// ListUsers lists all users
func (s *InMemoryUserStore) ListUsers(ctx context.Context) ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a list of users
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		// Create a copy of the user
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users, nil
}

// Close closes the user store
func (s *InMemoryUserStore) Close() error {
	// Nothing to close for in-memory store
	return nil
}

// InMemorySessionStore is a simple in-memory implementation of SessionStore
type InMemorySessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session
func (s *InMemorySessionStore) CreateSession(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session already exists
	if _, exists := s.sessions[session.ID]; exists {
		return fmt.Errorf("session with ID %s already exists", session.ID)
	}

	// Create a copy of the session
	newSession := *session

	// Set created timestamp if not set
	if newSession.CreatedAt.IsZero() {
		newSession.CreatedAt = time.Now()
	}

	// Store the session
	s.sessions[session.ID] = &newSession

	return nil
}

// GetSession retrieves a session by ID
func (s *InMemorySessionStore) GetSession(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", id)
	}

	// Return a copy of the session
	sessionCopy := *session
	return &sessionCopy, nil
}

// UpdateSession updates an existing session
func (s *InMemorySessionStore) UpdateSession(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	if _, exists := s.sessions[session.ID]; !exists {
		return fmt.Errorf("session with ID %s not found", session.ID)
	}

	// Create a copy of the session
	updatedSession := *session

	// Store the updated session
	s.sessions[session.ID] = &updatedSession

	return nil
}

// DeleteSession deletes a session by ID
func (s *InMemorySessionStore) DeleteSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	if _, exists := s.sessions[id]; !exists {
		return fmt.Errorf("session with ID %s not found", id)
	}

	// Delete the session
	delete(s.sessions, id)

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *InMemorySessionStore) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a list of sessions for the user
	sessions := make([]*Session, 0)
	for _, session := range s.sessions {
		if session.UserID == userID {
			// Create a copy of the session
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// CleanExpiredSessions removes all expired sessions
func (s *InMemorySessionStore) CleanExpiredSessions(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current time
	now := time.Now()

	// Remove expired sessions
	for id, session := range s.sessions {
		if session.ExpiresAt.Before(now) {
			delete(s.sessions, id)
		}
	}

	return nil
}

// Close closes the session store
func (s *InMemorySessionStore) Close() error {
	// Nothing to close for in-memory store
	return nil
}

// InMemoryAuditLoggerWithTypes is a simple in-memory implementation of AuditLogger using common.AuditLog
type InMemoryAuditLoggerWithTypes struct {
	logs map[string]*common.AuditLog
	mu   sync.RWMutex
}

// NewInMemoryAuditLoggerWithTypes creates a new in-memory audit logger with types
func NewInMemoryAuditLoggerWithTypes() *InMemoryAuditLoggerWithTypes {
	return &InMemoryAuditLoggerWithTypes{
		logs: make(map[string]*common.AuditLog),
	}
}

// Initialize initializes the audit logger
func (l *InMemoryAuditLoggerWithTypes) Initialize(ctx context.Context) error {
	// Nothing to initialize for in-memory logger
	return nil
}

// LogAudit logs an audit event
func (l *InMemoryAuditLoggerWithTypes) LogAudit(ctx context.Context, log *common.AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a copy of the log
	newLog := *log

	// Set timestamp if not set
	if newLog.Timestamp.IsZero() {
		newLog.Timestamp = time.Now()
	}

	// Store the log
	l.logs[log.ID] = &newLog

	return nil
}

// GetAuditLogs retrieves audit logs
func (l *InMemoryAuditLoggerWithTypes) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*common.AuditLog, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Create a list of logs
	logs := make([]*common.AuditLog, 0, len(l.logs))
	for _, log := range l.logs {
		// Apply filters
		if filter != nil {
			// TODO: Implement filtering
		}

		// Create a copy of the log
		logCopy := *log
		logs = append(logs, &logCopy)
	}

	// Sort logs by timestamp (newest first)
	// TODO: Implement sorting

	// Apply pagination
	total := len(logs)
	if offset >= total {
		return []*common.AuditLog{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return logs[offset:end], total, nil
}

// GetAuditLogByID retrieves an audit log by ID
func (l *InMemoryAuditLoggerWithTypes) GetAuditLogByID(ctx context.Context, id string) (*common.AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	log, exists := l.logs[id]
	if !exists {
		return nil, fmt.Errorf("audit log with ID %s not found", id)
	}

	// Return a copy of the log
	logCopy := *log
	return &logCopy, nil
}

// Close closes the logger
func (l *InMemoryAuditLoggerWithTypes) Close() error {
	// Nothing to close for in-memory logger
	return nil
}

// InMemoryIncidentStore is a simple in-memory implementation of IncidentStore
type InMemoryIncidentStore struct {
	incidents map[string]*common.SecurityIncident
	mu        sync.RWMutex
}

// NewInMemoryIncidentStore creates a new in-memory incident store
func NewInMemoryIncidentStore() *InMemoryIncidentStore {
	return &InMemoryIncidentStore{
		incidents: make(map[string]*common.SecurityIncident),
	}
}

// CreateIncident creates a new security incident
func (s *InMemoryIncidentStore) CreateIncident(ctx context.Context, incident *common.SecurityIncident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if incident already exists
	if _, exists := s.incidents[incident.ID]; exists {
		return fmt.Errorf("incident with ID %s already exists", incident.ID)
	}

	// Create a copy of the incident
	newIncident := *incident

	// Set detected timestamp if not set
	if newIncident.DetectedAt.IsZero() {
		newIncident.DetectedAt = time.Now()
	}

	// Store the incident
	s.incidents[incident.ID] = &newIncident

	return nil
}

// GetIncidentByID retrieves a security incident by ID
func (s *InMemoryIncidentStore) GetIncidentByID(ctx context.Context, id string) (*common.SecurityIncident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, exists := s.incidents[id]
	if !exists {
		return nil, fmt.Errorf("incident with ID %s not found", id)
	}

	// Return a copy of the incident
	incidentCopy := *incident
	return &incidentCopy, nil
}

// UpdateIncident updates an existing security incident
func (s *InMemoryIncidentStore) UpdateIncident(ctx context.Context, incident *common.SecurityIncident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if incident exists
	if _, exists := s.incidents[incident.ID]; !exists {
		return fmt.Errorf("incident with ID %s not found", incident.ID)
	}

	// Create a copy of the incident
	updatedIncident := *incident

	// Store the updated incident
	s.incidents[incident.ID] = &updatedIncident

	return nil
}

// DeleteIncident deletes a security incident by ID
func (s *InMemoryIncidentStore) DeleteIncident(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if incident exists
	if _, exists := s.incidents[id]; !exists {
		return fmt.Errorf("incident with ID %s not found", id)
	}

	// Delete the incident
	delete(s.incidents, id)

	return nil
}

// ListIncidents lists security incidents
func (s *InMemoryIncidentStore) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*common.SecurityIncident, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a list of incidents
	incidents := make([]*common.SecurityIncident, 0, len(s.incidents))
	for _, incident := range s.incidents {
		// Apply filters
		if filter != nil {
			// TODO: Implement filtering
		}

		// Create a copy of the incident
		incidentCopy := *incident
		incidents = append(incidents, &incidentCopy)
	}

	// Sort incidents by detected timestamp (newest first)
	// TODO: Implement sorting

	// Apply pagination
	total := len(incidents)
	if offset >= total {
		return []*common.SecurityIncident{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return incidents[offset:end], total, nil
}

// Close closes the store
func (s *InMemoryIncidentStore) Close() error {
	// Nothing to close for in-memory store
	return nil
}

// InMemoryVulnerabilityStore is a simple in-memory implementation of VulnerabilityStore
type InMemoryVulnerabilityStore struct {
	vulnerabilities map[string]*common.Vulnerability
	mu              sync.RWMutex
}

// NewInMemoryVulnerabilityStore creates a new in-memory vulnerability store
func NewInMemoryVulnerabilityStore() *InMemoryVulnerabilityStore {
	return &InMemoryVulnerabilityStore{
		vulnerabilities: make(map[string]*common.Vulnerability),
	}
}

// CreateVulnerability creates a new vulnerability
func (s *InMemoryVulnerabilityStore) CreateVulnerability(ctx context.Context, vulnerability *common.Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vulnerability already exists
	if _, exists := s.vulnerabilities[vulnerability.ID]; exists {
		return fmt.Errorf("vulnerability with ID %s already exists", vulnerability.ID)
	}

	// Create a copy of the vulnerability
	newVulnerability := *vulnerability

	// Set discovered timestamp if not set
	if newVulnerability.DiscoveredAt.IsZero() {
		newVulnerability.DiscoveredAt = time.Now()
	}

	// Store the vulnerability
	s.vulnerabilities[vulnerability.ID] = &newVulnerability

	return nil
}

// GetVulnerabilityByID retrieves a vulnerability by ID
func (s *InMemoryVulnerabilityStore) GetVulnerabilityByID(ctx context.Context, id string) (*common.Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vulnerability, exists := s.vulnerabilities[id]
	if !exists {
		return nil, fmt.Errorf("vulnerability with ID %s not found", id)
	}

	// Return a copy of the vulnerability
	vulnerabilityCopy := *vulnerability
	return &vulnerabilityCopy, nil
}

// UpdateVulnerability updates an existing vulnerability
func (s *InMemoryVulnerabilityStore) UpdateVulnerability(ctx context.Context, vulnerability *common.Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vulnerability exists
	if _, exists := s.vulnerabilities[vulnerability.ID]; !exists {
		return fmt.Errorf("vulnerability with ID %s not found", vulnerability.ID)
	}

	// Create a copy of the vulnerability
	updatedVulnerability := *vulnerability

	// Store the updated vulnerability
	s.vulnerabilities[vulnerability.ID] = &updatedVulnerability

	return nil
}

// DeleteVulnerability deletes a vulnerability by ID
func (s *InMemoryVulnerabilityStore) DeleteVulnerability(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vulnerability exists
	if _, exists := s.vulnerabilities[id]; !exists {
		return fmt.Errorf("vulnerability with ID %s not found", id)
	}

	// Delete the vulnerability
	delete(s.vulnerabilities, id)

	return nil
}

// ListVulnerabilities lists vulnerabilities
func (s *InMemoryVulnerabilityStore) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*common.Vulnerability, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a list of vulnerabilities
	vulnerabilities := make([]*common.Vulnerability, 0, len(s.vulnerabilities))
	for _, vulnerability := range s.vulnerabilities {
		// Apply filters
		if filter != nil {
			// TODO: Implement filtering
		}

		// Create a copy of the vulnerability
		vulnerabilityCopy := *vulnerability
		vulnerabilities = append(vulnerabilities, &vulnerabilityCopy)
	}

	// Sort vulnerabilities by discovered timestamp (newest first)
	// TODO: Implement sorting

	// Apply pagination
	total := len(vulnerabilities)
	if offset >= total {
		return []*common.Vulnerability{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return vulnerabilities[offset:end], total, nil
}

// Close closes the store
func (s *InMemoryVulnerabilityStore) Close() error {
	// Nothing to close for in-memory store
	return nil
}
