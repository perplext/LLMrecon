package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// UserStore manages storage of users
type UserStore struct {
	// filePath is the path to the user store file
	filePath string
	
	// users is the in-memory cache of users
	users map[string]*User
	
	// mutex protects the users map
	mutex sync.RWMutex
}

// NewUserStore creates a new user store
func NewUserStore(filePath string) (*UserStore, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	store := &UserStore{
		filePath: filePath,
		users:    make(map[string]*User),
	}
	
	// Load users from file if it exists
	if _, err := os.Stat(filePath); err == nil {
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to load users: %w", err)
		}
	} else {
		// Create default admin user if file doesn't exist
		adminUser := &User{
			ID:          "admin",
			Username:    "admin",
			Role:        AdminRole,
			Permissions: []Permission{ReadPermission, WritePermission, DeletePermission, AdminPermission},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		store.users[adminUser.ID] = adminUser
		
		// Save to file
		if err := store.save(); err != nil {
			return nil, fmt.Errorf("failed to save default admin user: %w", err)
		}
	}
	
	return store, nil
}

// load loads users from the file
func (s *UserStore) load() error {
	// Read file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	
	// Parse JSON
	var storedUsers []*User
	if err := json.Unmarshal(data, &storedUsers); err != nil {
		return err
	}
	
	// Update in-memory cache
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.users = make(map[string]*User)
	for _, user := range storedUsers {
		s.users[user.ID] = user
	}
	
	return nil
}

// save saves users to the file
func (s *UserStore) save() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Convert map to slice
	var storedUsers []*User
	for _, user := range s.users {
		storedUsers = append(storedUsers, user)
	}
	
	// Convert to JSON
	data, err := json.MarshalIndent(storedUsers, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file with secure permissions
	return os.WriteFile(s.filePath, data, 0600)
}

// GetUser gets a user by ID
func (s *UserStore) GetUser(id string) (*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID '%s' not found", id)
	}
	
	return user, nil
}

// SaveUser saves a user
func (s *UserStore) SaveUser(user *User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Set timestamps
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now
	
	// Save to in-memory cache
	s.users[user.ID] = user
	
	// Save to file
	return s.save()
}

// DeleteUser deletes a user by ID
func (s *UserStore) DeleteUser(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if user exists
	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user with ID '%s' not found", id)
	}
	
	// Prevent deleting the last admin user
	if id == "admin" {
		// Count admin users
		adminCount := 0
		for _, user := range s.users {
			if user.Role == AdminRole {
				adminCount++
			}
		}
		
		if adminCount <= 1 {
			return fmt.Errorf("cannot delete the last admin user")
		}
	}
	
	// Delete from in-memory cache
	delete(s.users, id)
	
	// Save to file
	return s.save()
}

// ListUsers lists all users
func (s *UserStore) ListUsers() ([]*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Convert map to slice
	var users []*User
	for _, user := range s.users {
		users = append(users, user)
	}
	
	return users, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (s *UserStore) UpdateLastLogin(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if user exists
	user, exists := s.users[id]
	if !exists {
		return fmt.Errorf("user with ID '%s' not found", id)
	}
	
	// Update last login timestamp
	user.LastLoginAt = time.Now()
	
	// Save to file
	return s.save()
}

// HasPermission checks if a user has a specific permission
func (s *UserStore) HasPermission(user *User, permission Permission) bool {
	// Admin role has all permissions
	if user.Role == AdminRole {
		return true
	}
	
	// Check if user has the specific permission
	for _, p := range user.Permissions {
		if p == permission {
			return true
		}
	}
	
	return false
}
