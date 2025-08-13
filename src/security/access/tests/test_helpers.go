// Package tests provides testing utilities for the access control system
package tests

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	
	"github.com/perplext/LLMrecon/src/security/access"
	"github.com/perplext/LLMrecon/src/security/access/db"
	"github.com/perplext/LLMrecon/src/security/access/models"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// MockAuditLogger implements the interfaces.AuditLogger interface for testing
type MockAuditLogger struct {
	Logs []*interfaces.AuditEvent
}

// LogEvent logs an audit event
func (m *MockAuditLogger) LogEvent(ctx context.Context, event *interfaces.AuditEvent) error {
	if event.ID == "" {
		event.ID = fmt.Sprintf("audit-%d", len(m.Logs)+1)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	m.Logs = append(m.Logs, event)
	return nil
}

// GetEventByID retrieves an audit event by ID
func (m *MockAuditLogger) GetEventByID(ctx context.Context, id string) (*interfaces.AuditEvent, error) {
	for _, event := range m.Logs {
		if event.ID == id {
			return event, nil
		}
	}
	return nil, fmt.Errorf("audit event not found: %s", id)
}

// QueryEvents queries audit events with filtering
func (m *MockAuditLogger) QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*interfaces.AuditEvent, int, error) {
	// For simplicity in tests, just return all logs without filtering
	total := len(m.Logs)
	
	// Apply pagination
	if offset >= total {
		return []*interfaces.AuditEvent{}, total, nil
	}
	
	end := offset + limit
	if end > total {
		end = total
	}
	
	return m.Logs[offset:end], total, nil
}

// ExportEvents exports audit events to a file
func (m *MockAuditLogger) ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error) {
	return "exported_audit_logs." + format, nil
}

// Close closes the audit logger
func (m *MockAuditLogger) Close() error {
	return nil
}

// UserManager defines the interface for user management operations
type UserManager interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *models.User) error
	
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	
	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	
	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *models.User) error
	
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id string) error
	
	// ListUsers lists users with optional filtering
	ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error)
}

// Role represents a role in the system
type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []string
	IsBuiltIn   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RBACManager defines the interface for role-based access control operations
type RBACManager interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, role *Role) error
	
	// GetRoleByID retrieves a role by ID
	GetRoleByID(ctx context.Context, id string) (*Role, error)
	
	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, role *Role) error
	
	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, id string) error
	
	// ListRoles lists roles with optional filtering
	ListRoles(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*Role, int, error)
	
	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	
	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	
	// GetUserRoles gets the roles assigned to a user
	GetUserRoles(ctx context.Context, userID string) ([]*Role, error)
	
	// HasPermission checks if a user has a specific permission
	HasPermission(ctx context.Context, userID, permission string) (bool, error)
}

// SessionManager defines the interface for session management operations
type SessionManager interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*models.Session, error)
	
	// ValidateSession validates a session token
	ValidateSession(ctx context.Context, token string) (*models.Session, error)
	
	// RefreshSession refreshes a session
	RefreshSession(ctx context.Context, refreshToken string) (*models.Session, error)
	
	// InvalidateSession invalidates a session
	InvalidateSession(ctx context.Context, token string) error
	
	// InvalidateAllUserSessions invalidates all sessions for a user
	InvalidateAllUserSessions(ctx context.Context, userID string) error
	
	// GetUserSessions gets all sessions for a user
	GetUserSessions(ctx context.Context, userID string) ([]*models.Session, error)
}

// SecurityManager defines the interface for security incident and vulnerability management
type SecurityManager interface {
	// CreateIncident creates a new security incident
	CreateIncident(ctx context.Context, incident *models.SecurityIncident) error
	
	// GetIncidentByID retrieves a security incident by ID
	GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error)
	
	// UpdateIncident updates an existing security incident
	UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error
	
	// DeleteIncident deletes a security incident
	DeleteIncident(ctx context.Context, id string) error
	
	// ListIncidents lists security incidents with optional filtering
	ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error)
	
	// CreateVulnerability creates a new vulnerability
	CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error
	
	// GetVulnerabilityByID retrieves a vulnerability by ID
	GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error)
	
	// UpdateVulnerability updates an existing vulnerability
	UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error
	
	// UpdateVulnerabilityStatus updates the status of a vulnerability
	UpdateVulnerabilityStatus(ctx context.Context, id string, status models.VulnerabilityStatus) error
	
	// DeleteVulnerability deletes a vulnerability
	DeleteVulnerability(ctx context.Context, id string) error
	
	// ListVulnerabilities lists vulnerabilities with optional filtering
	ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error)
}

// BoundaryEnforcer defines the interface for boundary enforcement operations
type BoundaryEnforcer interface {
	// EnforceIPRestriction enforces IP restrictions
	EnforceIPRestriction(ctx context.Context, userID, ipAddress string) (bool, error)
	
	// EnforceTimeRestriction enforces time restrictions
	EnforceTimeRestriction(ctx context.Context, userID string) (bool, error)
	
	// EnforceLocationRestriction enforces location restrictions
	EnforceLocationRestriction(ctx context.Context, userID, location string) (bool, error)
	
	// EnforceDeviceRestriction enforces device restrictions
	EnforceDeviceRestriction(ctx context.Context, userID, deviceID string) (bool, error)
	
	// EnforceRateLimiting enforces rate limiting
	EnforceRateLimiting(ctx context.Context, userID, action string) (bool, error)
}

// AuthManager defines the interface for authentication operations
type AuthManager interface {
	// Login authenticates a user
	Login(ctx context.Context, username, password, ipAddress, userAgent string) (*models.Session, error)
	
	// Logout logs out a user
	Logout(ctx context.Context, token string) error
	
	// VerifyMFA verifies multi-factor authentication
	VerifyMFA(ctx context.Context, userID, code string) (bool, error)
	
	// ResetPassword resets a user's password
	ResetPassword(ctx context.Context, userID, newPassword string) error
	
	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

// MockUserStore implements the interfaces.UserStore interface for testing
type MockUserStore struct {
	users           map[string]*models.User
	usersByUsername map[string]*models.User
}

func (m *MockUserStore) CreateUser(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	m.usersByUsername[user.Username] = user
	return nil
}

func (m *MockUserStore) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return user, nil
}

func (m *MockUserStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, ok := m.usersByUsername[username]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	return user, nil
}

func (m *MockUserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found with email: %s", email)
}

func (m *MockUserStore) UpdateUser(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	m.usersByUsername[user.Username] = user
	return nil
}

func (m *MockUserStore) DeleteUser(ctx context.Context, id string) error {
	user, ok := m.users[id]
	if !ok {
		return fmt.Errorf("user not found: %s", id)
	}
	delete(m.usersByUsername, user.Username)
	delete(m.users, id)
	return nil
}

func (m *MockUserStore) ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error) {
	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	total := len(users)

	// Apply pagination
	if offset >= total {
		return []*models.User{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return users[offset:end], total, nil
}

func (m *MockUserStore) Close() error {
	return nil
}

// MockSessionStore implements the interfaces.SessionStore interface for testing
type MockSessionStore struct {
	sessions map[string]*models.Session
}

func (m *MockSessionStore) CreateSession(ctx context.Context, session *models.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionStore) GetSessionByID(ctx context.Context, id string) (*models.Session, error) {
	session, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

func (m *MockSessionStore) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	for _, session := range m.sessions {
		if session.Token == token {
			return session, nil
		}
	}
	return nil, fmt.Errorf("session not found with token: %s", token)
}

func (m *MockSessionStore) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	for _, session := range m.sessions {
		if session.RefreshToken == refreshToken {
			return session, nil
		}
	}
	return nil, fmt.Errorf("session not found with refresh token: %s", refreshToken)
}

func (m *MockSessionStore) UpdateSession(ctx context.Context, session *models.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionStore) DeleteSession(ctx context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *MockSessionStore) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *MockSessionStore) ListSessionsByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	sessions := make([]*models.Session, 0)
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *MockSessionStore) CleanExpiredSessions(ctx context.Context) (int, error) {
	count := 0
	now := time.Now()
	for id, session := range m.sessions {
		if session.ExpiresAt.Before(now) {
			delete(m.sessions, id)
			count++
		}
	}
	return count, nil
}

func (m *MockSessionStore) Close() error {
	return nil
}

// MockIncidentStore implements the interfaces.IncidentStore interface for testing
type MockIncidentStore struct {
	incidents map[string]*models.SecurityIncident
}

func (m *MockIncidentStore) CreateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	m.incidents[incident.ID] = incident
	return nil
}

func (m *MockIncidentStore) GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error) {
	incident, ok := m.incidents[id]
	if !ok {
		return nil, fmt.Errorf("incident not found: %s", id)
	}
	return incident, nil
}

func (m *MockIncidentStore) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	m.incidents[incident.ID] = incident
	return nil
}

func (m *MockIncidentStore) DeleteIncident(ctx context.Context, id string) error {
	delete(m.incidents, id)
	return nil
}

func (m *MockIncidentStore) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	incidents := make([]*models.SecurityIncident, 0, len(m.incidents))
	for _, incident := range m.incidents {
		incidents = append(incidents, incident)
	}
	total := len(incidents)

	// Apply pagination
	if offset >= total {
		return []*models.SecurityIncident{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return incidents[offset:end], total, nil
}

func (m *MockIncidentStore) Close() error {
	return nil
}

// MockVulnerabilityStore implements the interfaces.VulnerabilityStore interface for testing
type MockVulnerabilityStore struct {
	vulnerabilities map[string]*models.Vulnerability
}

func (m *MockVulnerabilityStore) CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	m.vulnerabilities[vulnerability.ID] = vulnerability
	return nil
}

func (m *MockVulnerabilityStore) GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error) {
	vulnerability, ok := m.vulnerabilities[id]
	if !ok {
		return nil, fmt.Errorf("vulnerability not found: %s", id)
	}
	return vulnerability, nil
}

func (m *MockVulnerabilityStore) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	m.vulnerabilities[vulnerability.ID] = vulnerability
	return nil
}

func (m *MockVulnerabilityStore) DeleteVulnerability(ctx context.Context, id string) error {
	delete(m.vulnerabilities, id)
	return nil
}

func (m *MockVulnerabilityStore) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	vulnerabilities := make([]*models.Vulnerability, 0, len(m.vulnerabilities))
	for _, vulnerability := range m.vulnerabilities {
		vulnerabilities = append(vulnerabilities, vulnerability)
	}
	total := len(vulnerabilities)

	// Apply pagination
	if offset >= total {
		return []*models.Vulnerability{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return vulnerabilities[offset:end], total, nil
}

func (m *MockVulnerabilityStore) Close() error {
	return nil
}

// MockUserManager implements the UserManager interface for testing
type MockUserManager struct {
	userStore *MockUserStore
}

func (m *MockUserManager) CreateUser(ctx context.Context, user *models.User) error {
	return m.userStore.CreateUser(ctx, user)
}

func (m *MockUserManager) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return m.userStore.GetUserByID(ctx, id)
}

func (m *MockUserManager) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return m.userStore.GetUserByUsername(ctx, username)
}

func (m *MockUserManager) UpdateUser(ctx context.Context, user *models.User) error {
	return m.userStore.UpdateUser(ctx, user)
}

func (m *MockUserManager) DeleteUser(ctx context.Context, id string) error {
	return m.userStore.DeleteUser(ctx, id)
}

func (m *MockUserManager) ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error) {
	return m.userStore.ListUsers(ctx, filter, offset, limit)
}

// MockRBACManager implements the RBACManager interface for testing
type MockRBACManager struct {
	roles map[string]*Role
	userRoles map[string][]string
	// Map of role name to permissions for string-based roles
	rolePermissions map[string][]string
}

// NewMockRBACManager creates a new mock RBAC manager
func NewMockRBACManager() *MockRBACManager {
	m := &MockRBACManager{
		roles:          make(map[string]*Role),
		userRoles:      make(map[string][]string),
		rolePermissions: make(map[string][]string),
	}
	
	// Initialize built-in roles with default permissions
	adminRole := &Role{
		ID:          "admin",
		Name:        "admin",
		Description: "Administrator role with all permissions",
		Permissions: []string{"user:create", "user:read", "user:update", "user:delete"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["admin"] = adminRole
	m.rolePermissions["admin"] = adminRole.Permissions
	
	managerRole := &Role{
		ID:          "manager",
		Name:        "manager",
		Description: "Manager role with management permissions",
		Permissions: []string{"user:read", "user:update"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["manager"] = managerRole
	m.rolePermissions["manager"] = managerRole.Permissions
	
	userRole := &Role{
		ID:          "user",
		Name:        "user",
		Description: "Standard user role",
		Permissions: []string{"user:read"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["user"] = userRole
	m.rolePermissions["user"] = userRole.Permissions
	
	guestRole := &Role{
		ID:          "guest",
		Name:        "guest",
		Description: "Guest role with limited permissions",
		Permissions: []string{},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["guest"] = guestRole
	m.rolePermissions["guest"] = guestRole.Permissions
	
	auditorRole := &Role{
		ID:          "auditor",
		Name:        "auditor",
		Description: "Auditor role with read-only permissions",
		Permissions: []string{"user:read", "audit:read", "audit:list"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["auditor"] = auditorRole
	m.rolePermissions["auditor"] = auditorRole.Permissions
	
	operatorRole := &Role{
		ID:          "operator",
		Name:        "operator",
		Description: "Operator role with operational permissions",
		Permissions: []string{"user:read", "user:update"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["operator"] = operatorRole
	m.rolePermissions["operator"] = operatorRole.Permissions
	
	automationRole := &Role{
		ID:          "automation",
		Name:        "automation",
		Description: "Automation role for system tasks",
		Permissions: []string{"user:read", "user:update"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["automation"] = automationRole
	m.rolePermissions["automation"] = automationRole.Permissions
	
	readonlyRole := &Role{
		ID:          "readonly",
		Name:        "readonly",
		Description: "Read-only role with limited permissions",
		Permissions: []string{"user:read"},
		IsBuiltIn:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.roles["readonly"] = readonlyRole
	m.rolePermissions["readonly"] = readonlyRole.Permissions
	
	return m
}

func (m *MockRBACManager) CreateRole(ctx context.Context, role *Role) error {
	if m.roles == nil {
		m.roles = make(map[string]*Role)
	}
	m.roles[role.ID] = role
	return nil
}

func (m *MockRBACManager) GetRoleByID(ctx context.Context, id string) (*Role, error) {
	if m.roles == nil {
		m.roles = make(map[string]*Role)
	}
	role, ok := m.roles[id]
	if !ok {
		return nil, fmt.Errorf("role not found: %s", id)
	}
	return role, nil
}

func (m *MockRBACManager) UpdateRole(ctx context.Context, role *Role) error {
	if m.roles == nil {
		m.roles = make(map[string]*Role)
	}
	m.roles[role.ID] = role
	return nil
}

func (m *MockRBACManager) DeleteRole(ctx context.Context, id string) error {
	if m.roles == nil {
		m.roles = make(map[string]*Role)
	}
	delete(m.roles, id)
	return nil
}

func (m *MockRBACManager) ListRoles(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*Role, int, error) {
	if m.roles == nil {
		m.roles = make(map[string]*Role)
	}
	roles := make([]*Role, 0, len(m.roles))
	for _, role := range m.roles {
		roles = append(roles, role)
	}
	total := len(roles)

	// Apply pagination
	if offset >= total {
		return []*Role{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return roles[offset:end], total, nil
}

func (m *MockRBACManager) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	if m.userRoles == nil {
		m.userRoles = make(map[string][]string)
	}
	m.userRoles[userID] = append(m.userRoles[userID], roleID)
	return nil
}

func (m *MockRBACManager) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	if m.userRoles == nil {
		m.userRoles = make(map[string][]string)
	}
	roles, ok := m.userRoles[userID]
	if !ok {
		return nil
	}

	newRoles := make([]string, 0, len(roles))
	for _, r := range roles {
		if r != roleID {
			newRoles = append(newRoles, r)
		}
	}
	m.userRoles[userID] = newRoles
	return nil
}

func (m *MockRBACManager) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	if m.userRoles == nil {
		m.userRoles = make(map[string][]string)
	}
	if m.roles == nil {
		m.roles = make(map[string]*Role)
	}

	roleIDs, ok := m.userRoles[userID]
	if !ok {
		return []*Role{}, nil
	}

	roles := make([]*Role, 0, len(roleIDs))
	for _, id := range roleIDs {
		if role, ok := m.roles[id]; ok {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

func (m *MockRBACManager) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	// Check if userID is a valid user ID in our system
	roleIDs, ok := m.userRoles[userID]
	if !ok {
		// If not found directly, it might be a User object ID or some other format
		// For the test environment, we'll check if it matches any known pattern
		
		// Check if it's a User object ID from the access package
		// This is a special case for the tests where we're passing a User object
		// instead of a user ID string
		if strings.HasPrefix(userID, "user-") {
			// It's likely a user ID in the format "user-X"
			roleIDs = []string{"user"} // Default to user role for testing
		} else if userID == "" {
			// Empty user ID, treat as guest
			roleIDs = []string{"guest"}
		} else {
			// For any other unrecognized format, assume it's a test case
			// and give it the basic user role
			roleIDs = []string{"user"}
		}
	}

	// Check if any of the user's roles has the permission
	for _, roleID := range roleIDs {
		// First check in the struct-based roles
		role, ok := m.roles[roleID]
		if ok {
			for _, p := range role.Permissions {
				if p == permission {
					return true, nil
				}
			}
		}

		// Then check in the string-based roles
		permissions, ok := m.rolePermissions[roleID]
		if ok {
			for _, p := range permissions {
				if p == permission {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// MockSessionManager implements the SessionManager interface for testing
type MockSessionManager struct {
	sessionStore *MockSessionStore
}

func (m *MockSessionManager) CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*models.Session, error) {
	session := &models.Session{
		ID:           fmt.Sprintf("session-%d", time.Now().UnixNano()),
		UserID:       userID,
		Token:        fmt.Sprintf("token-%s-%d", userID, time.Now().UnixNano()),
		RefreshToken: fmt.Sprintf("refresh-%s-%d", userID, time.Now().UnixNano()),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
		MFACompleted: false,
		CreatedAt:    time.Now(),
	}

	err := m.sessionStore.CreateSession(ctx, session)
	return session, err
}

func (m *MockSessionManager) ValidateSession(ctx context.Context, token string) (*models.Session, error) {
	session, err := m.sessionStore.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

func (m *MockSessionManager) RefreshSession(ctx context.Context, refreshToken string) (*models.Session, error) {
	session, err := m.sessionStore.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	session.Token = fmt.Sprintf("token-%s-%d", session.UserID, time.Now().UnixNano())
	session.RefreshToken = fmt.Sprintf("refresh-%s-%d", session.UserID, time.Now().UnixNano())
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	session.LastActivity = time.Now()

	err = m.sessionStore.UpdateSession(ctx, session)
	return session, err
}

func (m *MockSessionManager) InvalidateSession(ctx context.Context, token string) error {
	session, err := m.sessionStore.GetSessionByToken(ctx, token)
	if err != nil {
		return err
	}

	return m.sessionStore.DeleteSession(ctx, session.ID)
}

func (m *MockSessionManager) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	return m.sessionStore.DeleteSessionsByUserID(ctx, userID)
}

func (m *MockSessionManager) GetUserSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	return m.sessionStore.ListSessionsByUserID(ctx, userID)
}

// MockSecurityManager implements the SecurityManager interface for testing
type MockSecurityManager struct {
	incidentStore      *MockIncidentStore
	vulnerabilityStore *MockVulnerabilityStore
}

// UpdateVulnerabilityStatus updates the status of a vulnerability
func (m *MockSecurityManager) UpdateVulnerabilityStatus(ctx context.Context, id string, status models.VulnerabilityStatus) error {
	vulnerability, err := m.vulnerabilityStore.GetVulnerabilityByID(ctx, id)
	if err != nil {
		return err
	}
	vulnerability.Status = status
	// Add timestamp to metadata if it doesn't exist
	if vulnerability.Metadata == nil {
		vulnerability.Metadata = make(map[string]interface{})
	}
	vulnerability.Metadata["updated_at"] = time.Now()
	
	// Set appropriate timestamps based on status
	if status == models.VulnerabilityStatusResolved {
		vulnerability.ResolvedAt = time.Now()
	} else if status == models.VulnerabilityStatusMitigated {
		vulnerability.MitigatedAt = time.Now()
	}
	
	return m.vulnerabilityStore.UpdateVulnerability(ctx, vulnerability)
}

func (m *MockSecurityManager) CreateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	return m.incidentStore.CreateIncident(ctx, incident)
}

func (m *MockSecurityManager) GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error) {
	return m.incidentStore.GetIncidentByID(ctx, id)
}

func (m *MockSecurityManager) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	return m.incidentStore.UpdateIncident(ctx, incident)
}

func (m *MockSecurityManager) DeleteIncident(ctx context.Context, id string) error {
	return m.incidentStore.DeleteIncident(ctx, id)
}

func (m *MockSecurityManager) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	return m.incidentStore.ListIncidents(ctx, filter, offset, limit)
}

func (m *MockSecurityManager) CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	return m.vulnerabilityStore.CreateVulnerability(ctx, vulnerability)
}

func (m *MockSecurityManager) GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error) {
	return m.vulnerabilityStore.GetVulnerabilityByID(ctx, id)
}

func (m *MockSecurityManager) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	return m.vulnerabilityStore.UpdateVulnerability(ctx, vulnerability)
}

func (m *MockSecurityManager) DeleteVulnerability(ctx context.Context, id string) error {
	return m.vulnerabilityStore.DeleteVulnerability(ctx, id)
}

func (m *MockSecurityManager) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	return m.vulnerabilityStore.ListVulnerabilities(ctx, filter, offset, limit)
}

// MockBoundaryEnforcer implements the BoundaryEnforcer interface for testing
type MockBoundaryEnforcer struct {}

func (m *MockBoundaryEnforcer) EnforceIPRestriction(ctx context.Context, userID, ipAddress string) (bool, error) {
	return true, nil
}

func (m *MockBoundaryEnforcer) EnforceTimeRestriction(ctx context.Context, userID string) (bool, error) {
	return true, nil
}

func (m *MockBoundaryEnforcer) EnforceLocationRestriction(ctx context.Context, userID, location string) (bool, error) {
	return true, nil
}

func (m *MockBoundaryEnforcer) EnforceDeviceRestriction(ctx context.Context, userID, deviceID string) (bool, error) {
	return true, nil
}

func (m *MockBoundaryEnforcer) EnforceRateLimiting(ctx context.Context, userID, action string) (bool, error) {
	return true, nil
}

// MockAuthManager implements the AuthManager interface for testing
type MockAuthManager struct{}

func (m *MockAuthManager) Login(ctx context.Context, username, password, ipAddress, userAgent string) (*models.Session, error) {
	return &models.Session{
		ID:           fmt.Sprintf("session-%d", time.Now().UnixNano()),
		UserID:       "user-" + username,
		Token:        fmt.Sprintf("token-%s-%d", username, time.Now().UnixNano()),
		RefreshToken: fmt.Sprintf("refresh-%s-%d", username, time.Now().UnixNano()),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
		MFACompleted: false,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockAuthManager) Logout(ctx context.Context, token string) error {
	return nil
}

func (m *MockAuthManager) VerifyMFA(ctx context.Context, userID, code string) (bool, error) {
	return true, nil
}

func (m *MockAuthManager) ResetPassword(ctx context.Context, userID, newPassword string) error {
	return nil
}

func (m *MockAuthManager) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	return nil
}

// TestContext represents a test context with all necessary components
type TestContext struct {
	// Test components
	T         *testing.T
	TempDir   string
	CleanupFn func()

	// Database components
	DB         *sql.DB
	DBFactory  *db.Factory
	DBConfig   *db.DBConfig
	DBFilePath string

	// Access control components
	Manager        *access.DBAccessControlManager
	AuditLogger    interfaces.AuditLogger
	AdminPass      string
	AdminUser      *models.User
	UserManager    UserManager
	RBACManager    RBACManager
	SessionManager SessionManager
	SecurityManager SecurityManager
	BoundaryEnforcer BoundaryEnforcer
	AuthManager    AuthManager

	// Stores
	UserStore        interfaces.UserStore
	SessionStore     interfaces.SessionStore
	IncidentStore    interfaces.IncidentStore
	VulnerabilityStore interfaces.VulnerabilityStore

	// Test data
	TestUsers map[string]*models.User
}

// NewTestContext creates a new test context with an in-memory database
func NewTestContext(t *testing.T) *TestContext {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "access-control-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create database file path
	dbFilePath := filepath.Join(tempDir, "test.db")

	// Create database config
	dbConfig := &db.DBConfig{
		Driver:       "sqlite3",
		DSN:          dbFilePath,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	// Create database factory
	factory, err := db.NewFactory(dbConfig)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create database factory: %v", err)
	}

	// Open database connection
	database, err := sql.Open(dbConfig.Driver, dbConfig.DSN)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create a compatible interfaces.DBConfig from the db.DBConfig
	interfaceDBConfig := &interfaces.DBConfig{
		Driver:       dbConfig.Driver,
		DSN:          dbConfig.DSN,
		MaxOpenConns: dbConfig.MaxOpenConns,
		MaxIdleConns: dbConfig.MaxIdleConns,
	}

	// Create access control manager config
	accessConfig := &access.DBAccessControlConfig{
		DBConfig:             interfaceDBConfig,
		DefaultAdminUsername: "admin",
		DefaultAdminPassword: "Admin123!",
		DefaultAdminEmail:    "admin@example.com",
	}

	// Create access control manager
	manager, err := access.NewDBAccessControlManager(accessConfig)
	if err != nil {
		database.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create access control manager: %v", err)
	}

	// Create cleanup function
	cleanupFn := func() {
		manager.Close()
		database.Close()
		os.RemoveAll(tempDir)
	}

	// Create mock audit logger
	mockAuditLogger := &MockAuditLogger{Logs: make([]*interfaces.AuditEvent, 0)}
	
	// Create test context
	ctx := &TestContext{
		T:                 t,
		TempDir:           tempDir,
		CleanupFn:         cleanupFn,
		DB:                database,
		DBFactory:         factory,
		DBConfig:          dbConfig,
		DBFilePath:        dbFilePath,
		Manager:           manager,
		AuditLogger:       mockAuditLogger,
		AdminPass:         accessConfig.DefaultAdminPassword,
		TestUsers:         make(map[string]*models.User),
	}
	
		// Create mock implementations for testing
	mockUserStore := &MockUserStore{users: make(map[string]*models.User), usersByUsername: make(map[string]*models.User)}
	mockSessionStore := &MockSessionStore{sessions: make(map[string]*models.Session)}
	mockIncidentStore := &MockIncidentStore{incidents: make(map[string]*models.SecurityIncident)}
	mockVulnerabilityStore := &MockVulnerabilityStore{vulnerabilities: make(map[string]*models.Vulnerability)}

	// Set up the interfaces
	mockUserManager := &MockUserManager{userStore: mockUserStore}
	ctx.UserManager = mockUserManager
	
	// Create a real RBAC manager with adapter instead of the mock
	roleStore := access.NewInMemoryRoleStore()
	rbacManager := access.NewRBACManager(mockUserManager, roleStore, mockAuditLogger)
	ctx.RBACManager = adapters.CreateRBACManagerAdapter(rbacManager)
	
	ctx.SessionManager = &MockSessionManager{sessionStore: mockSessionStore}
	ctx.SecurityManager = &MockSecurityManager{incidentStore: mockIncidentStore, vulnerabilityStore: mockVulnerabilityStore}
	ctx.BoundaryEnforcer = &MockBoundaryEnforcer{}
	ctx.AuthManager = &MockAuthManager{}
	
	// Set up the stores
	ctx.UserStore = mockUserStore
	ctx.SessionStore = mockSessionStore
	ctx.IncidentStore = mockIncidentStore
	ctx.VulnerabilityStore = mockVulnerabilityStore
	
	// Create admin user
	ctx.AdminUser, _ = ctx.UserStore.GetUserByUsername(context.Background(), accessConfig.DefaultAdminUsername)

	return ctx
}

// CreateTestUser creates a test user with the given username, password, email, and roles
func (c *TestContext) CreateTestUser(username, password, email string, roles []string) *models.User {
	// Create a new user
	user := &models.User{
		ID:         fmt.Sprintf("user-%d", len(c.TestUsers)+1),
		Username:   username,
		Email:      email,
		Roles:      roles,
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Hash the password
	user.PasswordHash = password // In a real implementation, this would be hashed

	// Store the user in the test context
	c.TestUsers[username] = user

	// Store the user in the user store
	if store, ok := c.UserStore.(*MockUserStore); ok {
		store.users[user.ID] = user
		store.usersByUsername[username] = user
	}

	return user
}

// CreateTestRole creates a test role with the given name, description, and permissions
func (c *TestContext) CreateTestRole(name, description string, permissions []string) *Role {
	// Create a new role
	role := &Role{
		ID:          fmt.Sprintf("role-%d", len(c.TestUsers)+1),
		Name:        name,
		Description: description,
		Permissions: permissions,
		IsBuiltIn:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return role
}

// CreateTestVulnerability creates a test vulnerability
// This is the full signature version
func (c *TestContext) CreateTestVulnerability(title, description, severity, reportedBy string, affectedSystems []string) *models.Vulnerability {
	vulnerability := &models.Vulnerability{
		ID:              fmt.Sprintf("vuln-%d", time.Now().UnixNano()),
		Title:           title,
		Description:     description,
		Severity:        models.VulnerabilitySeverity(severity),
		Status:          models.VulnerabilityStatusOpen,
		ReportedAt:      time.Now(),
		ReportedBy:      reportedBy,
		AffectedSystems: affectedSystems,
		CVE:             "", // Initialize with empty string
		Metadata:        make(map[string]interface{}),
	}

	err := c.SecurityManager.CreateVulnerability(context.Background(), vulnerability)
	if err != nil {
		c.T.Fatalf("Failed to create test vulnerability: %v", err)
	}

	return vulnerability
}

// CreateTestVulnerabilitySimple creates a test vulnerability with a simpler signature
// This is for backward compatibility with existing tests
func (c *TestContext) CreateTestVulnerabilitySimple(title, description, severity, reportedBy string) *models.Vulnerability {
	return c.CreateTestVulnerability(title, description, severity, reportedBy, []string{})
}

// GetVulnerabilityByCVE gets a vulnerability by its CVE ID
func (m *MockSecurityManager) GetVulnerabilityByCVE(ctx context.Context, cve string) (*models.Vulnerability, error) {
	if m.vulnerabilityStore == nil {
		return nil, fmt.Errorf("vulnerability store not initialized")
	}
	
	vulnerabilities, _, err := m.vulnerabilityStore.ListVulnerabilities(ctx, nil, 0, 1000)
	if err != nil {
		return nil, err
	}
	
	for _, v := range vulnerabilities {
		if v.CVE == cve {
			return v, nil
		}
	}
	
	return nil, fmt.Errorf("vulnerability with CVE %s not found", cve)
}

// EscalateIncident escalates a security incident
func (m *MockSecurityManager) EscalateIncident(ctx context.Context, id string, severity models.SecurityIncidentSeverity, reason string) error {
	incident, err := m.incidentStore.GetIncidentByID(ctx, id)
	if err != nil {
		return err
	}
	
	incident.Severity = severity
	incident.Description += "\n\nEscalated: " + reason
	
	if incident.Metadata == nil {
		incident.Metadata = make(map[string]interface{})
	}
	incident.Metadata["escalated_at"] = time.Now()
	incident.Metadata["escalation_reason"] = reason
	
	return m.incidentStore.UpdateIncident(ctx, incident)
}

// AssignIncident assigns a security incident to a user
func (m *MockSecurityManager) AssignIncident(ctx context.Context, id string, assigneeID string) error {
	incident, err := m.incidentStore.GetIncidentByID(ctx, id)
	if err != nil {
		return err
	}
	
	incident.AssignedTo = assigneeID
	
	if incident.Metadata == nil {
		incident.Metadata = make(map[string]interface{})
	}
	incident.Metadata["assigned_at"] = time.Now()
	
	return m.incidentStore.UpdateIncident(ctx, incident)
}

// AddRemediationPlan adds a remediation plan to a vulnerability
func (m *MockSecurityManager) AddRemediationPlan(ctx context.Context, id string, plan string) error {
	vulnerability, err := m.vulnerabilityStore.GetVulnerabilityByID(ctx, id)
	if err != nil {
		return err
	}
	
	if vulnerability.Metadata == nil {
		vulnerability.Metadata = make(map[string]interface{})
	}
	vulnerability.Metadata["remediation_plan"] = plan
	vulnerability.Mitigation = plan
	
	return m.vulnerabilityStore.UpdateVulnerability(ctx, vulnerability)
}

// MarkVulnerabilityRemediated marks a vulnerability as remediated
func (m *MockSecurityManager) MarkVulnerabilityRemediated(ctx context.Context, id string, details string) error {
	vulnerability, err := m.vulnerabilityStore.GetVulnerabilityByID(ctx, id)
	if err != nil {
		return err
	}
	
	vulnerability.Status = models.VulnerabilityStatusResolved
	vulnerability.ResolvedAt = time.Now()
	
	if vulnerability.Metadata == nil {
		vulnerability.Metadata = make(map[string]interface{})
	}
	vulnerability.Metadata["resolution"] = details
	vulnerability.Metadata["resolved_at"] = time.Now()
	
	return m.vulnerabilityStore.UpdateVulnerability(ctx, vulnerability)
}

// WaitForAuditLog waits for an audit log entry with the specified action
func (c *TestContext) WaitForAuditLog(action string, timeout time.Duration) *interfaces.AuditEvent {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		logs, _, _ := c.AuditLogger.QueryEvents(context.Background(), map[string]interface{}{"action": action}, 0, 10)
		if len(logs) > 0 {
			return logs[0]
		}
		time.Sleep(10 * time.Millisecond)
	}
	c.T.Fatalf("Timed out waiting for audit log with action: %s", action)
	return nil
}

// LoginUser logs in a test user
func (c *TestContext) LoginUser(username, password string) (string, error) {
	// For testing purposes, just return a mock session ID
	return "mock-session-" + username, nil
}

// CreateTestIncident creates a test incident
func (c *TestContext) CreateTestIncident(title, description, severity, reportedBy string) *models.SecurityIncident {
	// Create incident
	incident := &models.SecurityIncident{
		Title:       title,
		Description: description,
		Severity:    models.SecurityIncidentSeverity(severity),
		Status:      models.SecurityIncidentStatusOpen,
		ReportedBy:  reportedBy,
		ReportedAt:  time.Now(),
	}
	
	return incident
}

// WaitForAuditEvent waits for an audit event with the specified action (deprecated, use WaitForAuditLog instead)
func (c *TestContext) WaitForAuditEvent(action string, timeout time.Duration) *interfaces.AuditEvent {
	// Create context
	ctx := context.Background()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		events, _, err := c.AuditLogger.QueryEvents(ctx, map[string]interface{}{"action": action}, 0, 10)
		if err != nil {
			c.T.Fatalf("Failed to query audit logs: %v", err)
		}
		
		if len(events) > 0 {
			return events[0]
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	c.T.Fatalf("Timed out waiting for audit event with action: %s", action)
	return nil
}

// AssertEqual asserts that two values are equal
func (c *TestContext) AssertEqual(expected, actual interface{}, message string) {
	if expected != actual {
		c.T.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func (c *TestContext) AssertNotEqual(expected, actual interface{}, message string) {
	if expected == actual {
		c.T.Errorf("%s: expected %v to be different from %v", message, expected, actual)
	}
}

// AssertTrue asserts that a condition is true
func (c *TestContext) AssertTrue(condition bool, message string) {
	if !condition {
		c.T.Errorf("%s: expected true, got false", message)
	}
}

// AssertFalse asserts that a condition is false
func (c *TestContext) AssertFalse(condition bool, message string) {
	if condition {
		c.T.Errorf("%s: expected false, got true", message)
	}
}

// AssertNil asserts that a value is nil
func (c *TestContext) AssertNil(value interface{}, message string) {
	if value != nil {
		c.T.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNotNil asserts that a value is not nil
func (c *TestContext) AssertNotNil(value interface{}, message string) {
	if value == nil {
		c.T.Errorf("%s: expected non-nil value", message)
	}
}

// AssertError asserts that an error is not nil
func (c *TestContext) AssertError(err error, message string) {
	if err == nil {
		c.T.Errorf("%s: expected error, got nil", message)
	}
}

// AssertNoError asserts that an error is nil
func (c *TestContext) AssertNoError(err error, message string) {
	if err != nil {
		c.T.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertContains asserts that a string contains a substring
func (c *TestContext) AssertContains(s, substring string, message string) {
	if !strings.Contains(s, substring) {
		c.T.Errorf("%s: expected %q to contain %q", message, s, substring)
	}
}

// AssertNotContains asserts that a string does not contain a substring
func (c *TestContext) AssertNotContains(s, substring string, message string) {
	if strings.Contains(s, substring) {
		c.T.Errorf("%s: expected %q not to contain %q", message, s, substring)
	}
}

// AssertLen asserts that a slice or map has the expected length
func (c *TestContext) AssertLen(value interface{}, expected int, message string) {
	var actual int
	
	switch v := value.(type) {
	case []interface{}:
		actual = len(v)
	case string:
		actual = len(v)
	case map[string]interface{}:
		actual = len(v)
	case []*models.User:
		actual = len(v)
	case []*access.Role:
		actual = len(v)
	case []*models.AuditLog:
		actual = len(v)
	case []*models.SecurityIncident:
		actual = len(v)
	case []*models.Vulnerability:
		actual = len(v)
	default:
		c.T.Fatalf("AssertLen: unsupported type %T", value)
	}
	
	if actual != expected {
		c.T.Errorf("%s: expected length %d, got %d", message, expected, actual)
	}
}
